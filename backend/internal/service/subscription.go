package service

import (
	"ABSC/internal/client"
	"ABSC/internal/database"
	"ABSC/internal/model"
	"fmt"
	"log"
	"path"
	"strings"

	"gorm.io/gorm/clause"
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
	cleanSeriesDir := strings.ReplaceAll(seriesDir, "\\", "/")

	return &SubscriptionService{
		QbitClient: qb,
		BangumiSrv: bangumiSrv,
		SeriesDir:  strings.TrimSuffix(cleanSeriesDir, "/"),
	}
}

// Subscribe 触发完整订阅流水线
func (s *SubscriptionService) Subscribe(req SubscribeRequest) error {
	if req.Season <= 0 {
		req.Season = 1
	}

	targetRSS := strings.TrimSpace(req.RSSURL)
	if targetRSS == "" {
		return fmt.Errorf("[Subscribe]  订阅失败: RSS URL 不能为空")
	}

	// 从 SQLite 获取番剧基本元数据 (干净的中文剧名 TitleCN)
	var bangumi model.BangumiMetadata
	if err := database.DB.Where("mikan_id = ?", req.MikanID).First(&bangumi).Error; err != nil {
		return fmt.Errorf("[Subscribe] 找不到对应的番剧元数据 (MikanID: %d): %w", req.MikanID, err)
	}

	// 计算并持久化集数偏移量 (Offset)
	offsetVal := 0
	if req.CustomOffset != nil {
		offsetVal = *req.CustomOffset
		log.Printf("使用用户自定义设置的集数偏移量: 减去 %d 集", offsetVal)
	} else if bangumi.TMDBID != nil {
		offsetVal = s.BangumiSrv.CalculateAutoOffset(*bangumi.TMDBID, req.Season)
	}
	offsetRecord := model.EpisodeOffset{
		MikanID:     req.MikanID,
		Season:      req.Season,
		OffsetValue: offsetVal,
	}

	// SQLite 等更新偏移量配置表episode_offset
	err := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "mikan_id"}, {Name: "season"}},
		DoUpdates: clause.AssignmentColumns([]string{"offset_value", "updated_at"}),
	}).Create(&offsetRecord).Error

	if err != nil {
		return fmt.Errorf("[Subscribe] 保存集数偏移配置失败: %w", err)
	}

	// 计算 NAS 上的物理存储落盘路径
	categoryName := bangumi.TitleCN
	savePath := path.Join(s.SeriesDir, categoryName, fmt.Sprintf("Season %d", req.Season))

	// 联动 qBittorrent 下发自动化订阅控制命令
	log.Printf("	正在向 qBittorrent 下发化订阅指令...")
	log.Printf("   ├─ 分类名称: %s", categoryName)
	log.Printf("   ├─ 保存路径: %s", savePath)
	log.Printf("   ├─ RSS 地址: %s", targetRSS)
	log.Printf("   └─ 关键字过滤规则: [包含: '%s'] | [排除: '%s']", req.MustContain, req.MustNotContain)

	if err := s.QbitClient.Login(); err != nil {
		return fmt.Errorf("[Subscribe] qBittorrent 登录失败: %w", err)
	}

	// 创建以剧名命名的分类并绑定路径
	if err := s.QbitClient.CreateCategory(categoryName, savePath); err != nil {
		return fmt.Errorf("[Subscribe] 创建分类失败: %w (可能已存在，忽略并继续)", err)
	}

	// 新增 RSS 订阅
	if err := s.QbitClient.AddRSSFeed(targetRSS, categoryName); err != nil {
		return fmt.Errorf("[Subscribe] 新增 RSS 订阅失败: %w (可能已存在，忽略并继续)", err)

	}

	// 配置 RSS 下载器规则
	rulDef := client.RuleDefinition{
		Enabled:        true,
		MustContain:    req.MustContain,    // 注入必须包含关键字
		MustNotContain: req.MustNotContain, // 注入排除包含关键字
		AffectedFeeds:  []string{targetRSS},
		TorrentParams: client.TorrentParams{
			Category:      categoryName,
			OperatingMode: "AutoManaged",
		},
	}

	if err := s.QbitClient.SetRSSRule(categoryName, rulDef); err != nil {
		return fmt.Errorf(" 设置 qB RSS 下载规则失败: %w ", err)
	}

	// 更新 SQLite 本地订阅状态表 (状态置为 'subscribing')
	subRecord := model.Subscription{
		MikanID:    req.MikanID,
		Status:     "subscribing",
		QBCategory: categoryName,
		RSSFeedURL: targetRSS,
		SavePath:   savePath,
	}

	err = database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "mikan_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "qb_category", "rss_feed_url", "save_path", "updated_at"}),
	}).Create(&subRecord).Error
	if err != nil {
		return fmt.Errorf("[Subscribe] 更新本地 SQLite 订阅状态失败: %w", err)
	}
	log.Printf("番剧《%s》 第 %d 季 自动化订阅流程完成", categoryName, req.Season)
	return nil

}
