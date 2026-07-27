package service

import (
	"ABSC/internal/client"
	"ABSC/internal/database"
	"ABSC/internal/model"
	"fmt"
	"strings"
)

type SubscribeRequest struct {
	MikanID        int    `json:"mikan_id"`
	SubgroupID     int    `json:"subgroup_id"`
	Season         int    `json:"season"`
	RSSURL         string `json:"rss_url"`
	MustContain    string `json:"must_contain,omitempty"`     // 必须包含的关键字（如 "简中" 或 "1080p"），默认空
	MustNotContain string `json:"must_not_contain,omitempty"` // 不能包含的关键字（如 "720p" 或 "繁中"）,默认空
	CustomOffset   *int   `json:"custom_offset,omitempty"`    // 前端用户微调值，若传 nil 则开启 TMDB 智能推导
}

type SubscriptionService struct {
	QbitClient *client.QBitClient
	BangumiSrv *BangumiService
	SeriesDir  string
}

func NewSubscriptionService(qb *client.QBitClient, bangumiSrv *BangumiService, seriesDir string) *SubscriptionService {
	return &SubscriptionService{
		QbitClient: qb,
		BangumiSrv: bangumiSrv,
		SeriesDir:  strings.TrimSuffix(seriesDir, "/"),
	}
}

// Subscribe 触发完整订阅流水线
func (s *SubscriptionService) Subscribe(req SubscribeRequest) error {
	if req.Season <= 0 {
		req.Season = 1
	}

	//从 SQLite 获取番剧基本信息
	var bangumi model.BangumiMetadata
	if err := database.DB.Where("mikan_id = ?", req.MikanID).First(&bangumi).Error; err != nil {
		return fmt.Errorf("[Subscribe] 找不到对应的番剧元数据 (MikanID: %d): %w", req.MikanID, err)
	}

	// 爬取详情页获取该字幕组的专属 RSS 订阅链接

	return nil

}
