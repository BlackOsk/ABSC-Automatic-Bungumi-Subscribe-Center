// 控制一键订阅流水线、整季安全下架清理以及订阅概览
package handler

import (
	"ABSC/internal/database"
	"ABSC/internal/model"
	"ABSC/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	SubSvc *service.SubscriptionService
	DelSvc *service.DeletionService
}

func NewSubscriptionHandler(subSvc *service.SubscriptionService, delSvc *service.DeletionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		SubSvc: subSvc,
		DelSvc: delSvc,
	}
}

// Subscribe 发起订阅
func (h *SubscriptionHandler) Subscribe(c *gin.Context) {
	var req service.SubscribeRequest
	// 解析传入参数
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 40002, "解析订阅参数 JSON 格式失败: "+err.Error())
		return
	}

	if err := h.SubSvc.Subscribe(req); err != nil {
		Fail(c, 50003, "一键订阅失败: "+err.Error())
		return
	}
	Success(c, "一键订阅成功")
}

// Purge 清理整季安全下架
func (h *SubscriptionHandler) Purge(c *gin.Context) {

	idStr := c.Param("mikan_id")
	mikanID, err := strconv.Atoi(idStr)
	if err != nil {
		Fail(c, 40001, "无效的 mikan_id 参数")
		return
	}

	deleteFilesQuery := c.DefaultPostForm("delete_files", "true")
	deleteFiles := deleteFilesQuery == "true"

	if err := h.DelSvc.PurgeSubscription(mikanID, deleteFiles); err != nil {
		Fail(c, 50004, "整季下架清理失败: "+err.Error())
		return
	}
	Success(c, "整季下架清理成功")

}

// List 获取当前已经订阅的番剧清单
func (h *SubscriptionHandler) List(c *gin.Context) {
	var subs []model.Subscription
	if err := database.DB.Where("status = ?", "subscribing").Find(&subs).Error; err != nil {
		Fail(c, 50005, "获取订阅清单失败: "+err.Error())
		return
	}
	Success(c, subs)
}
