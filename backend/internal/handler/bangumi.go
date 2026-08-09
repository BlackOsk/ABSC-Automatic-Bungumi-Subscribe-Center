package handler

import (
	"ABSC/internal/database"
	"ABSC/internal/model"
	"ABSC/internal/scraper"
	"ABSC/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BangumiHandler struct {
	BangumiService *service.BangumiService
}

func NewBangumiHandler(svc *service.BangumiService) *BangumiHandler {
	return &BangumiHandler{
		BangumiService: svc,
	}
}

// BangumiListItem 扩展版新番卡片结构 (带订阅状态)
type BangumiListItem struct {
	model.BangumiMetadata
	SubStatus string `json:"sub_status"` // 'none' | 'subscribing'
}

// GetCurrent 获取当季全量新番列表 (聚合 SQLite 中的 TMDB 元数据与本地订阅状态)
func (h *BangumiHandler) GetCurrent(c *gin.Context) {
	var bangumis []model.BangumiMetadata
	if err := database.DB.Order("broadcast_day ASC, mikan_id DESC").Find(&bangumis).Error; err != nil {
		Fail(c, 50002, "获取本地新番列表失败："+err.Error())
		return
	}

	// 获取订阅状态
	var subs []model.Subscription
	database.DB.Find(&subs)
	subMap := make(map[int]string)
	for _, sub := range subs {
		subMap[sub.MikanID] = "subscribing"
	}

	var result []BangumiListItem
	for _, b := range bangumis {
		status := "none"
		if st, ok := subMap[b.MikanID]; ok {
			status = st
		}
		result = append(result, BangumiListItem{
			BangumiMetadata: b,
			SubStatus:       status,
		})
	}

	Success(c, result)

}

// GetDetail 单部番剧详情下钻 (按需实时懒加载字幕组文件列表及 RSSURL)
func (h *BangumiHandler) GetDetail(c *gin.Context) {
	idStr := c.Param("mikan_id")
	mikanID, err := strconv.Atoi(idStr)
	if err != nil {
		Fail(c, 40001, "无效的 mikan_id 参数")
		return
	}

	// 执行详情抓取
	subResources, err := scraper.ScrapeBangumiDetail(mikanID)
	if err != nil {
		Fail(c, 50002, "获取番剧详情失败："+err.Error())
		return
	}

	Success(c, subResources)
}

// SyncTMDB 手动触发从 Mikan 和 TMDB 增量同步最新番剧
func (h *BangumiHandler) SyncTMDB(c *gin.Context) {
	go func() {
		h.BangumiService.SyncCurrentQuarterBangumi()
	}()
	Success(c, "同步任务已在后台成功启动")
}
