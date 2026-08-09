package handler

import (
	"ABSC/internal/database"
	"ABSC/internal/model"
	"ABSC/internal/service"
	"path"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

type OffsetHandler struct{}

func NewOffsetHandler() *OffsetHandler {
	return &OffsetHandler{}
}

// UpdateOffsetRequest 更改偏移量请求 Payload
type UpdateOffsetRequest struct {
	MikanID     int `json:"mikan_id" binding:"required"`
	Season      int `json:"season" binding:"required"`
	OffsetValue int `json:"offset"`
}

// PreviewRenameRequest 实时预览请求 Payload
type PreviewRenameRequest struct {
	FileName string `json:"file_name" binding:"required"` // 原始文件名，如 "[LoliHouse] Title - 13.mp4"
	Offset   int    `json:"offset"`                       // 偏移量，如 11
}

// PreviewRenameResponse 实时预览响应 Payload
type PreviewRenameResponse struct {
	Matched    bool   `json:"matched"`
	OldName    string `json:"old_name"`
	NewName    string `json:"new_name"`    // 预览改名后，如 "E02.mp4"
	RelativeEp string `json:"relative_ep"` // 如 "E02"
}

// GetOffsets 查询某部番剧所有的季数偏移配置
func (h *OffsetHandler) GetOffsets(c *gin.Context) {
	idStr := c.Param("mikan_id")
	mikanID, err := strconv.Atoi(idStr)
	if err != nil {
		Fail(c, 40001, "无效的 MikanID 参数")
		return
	}

	var offsets []model.EpisodeOffset
	database.DB.Where("mikan_id = ?", mikanID).Order("season ASC").Find(&offsets)
	Success(c, offsets)
}

// UpdateOffset 增加/更新某季的集数偏移数字
func (h *OffsetHandler) UpdateOffset(c *gin.Context) {
	var req UpdateOffsetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 40002, "请求参数不合法")
		return
	}

	offsetRecord := model.EpisodeOffset{
		MikanID:     req.MikanID,
		Season:      req.Season,
		OffsetValue: req.OffsetValue,
	}

	err := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "mikan_id"}, {Name: "season"}},
		DoUpdates: clause.AssignmentColumns([]string{"offset_value", "updated_at"}),
	}).Create(&offsetRecord).Error

	if err != nil {
		Fail(c, 50006, "保存偏移配置失败: "+err.Error())
		return
	}

	Success(c, "偏移量配置更新成功")
}

// PreviewRename **改名实时预览** 核心接口
func (h *OffsetHandler) PreviewRename(c *gin.Context) {
	var req PreviewRenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 40002, "请求参数不合法")
		return
	}

	matched, epStr := service.ExtractAndCalculateEpisode(req.FileName, req.Offset)
	if !matched {
		Success(c, PreviewRenameResponse{
			Matched: false,
			OldName: req.FileName,
			NewName: req.FileName,
		})
		return
	}

	ext := path.Ext(req.FileName)
	newName := epStr + ext

	Success(c, PreviewRenameResponse{
		Matched:    true,
		OldName:    req.FileName,
		NewName:    newName,
		RelativeEp: epStr,
	})
}
