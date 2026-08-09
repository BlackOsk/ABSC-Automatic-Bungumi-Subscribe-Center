package service

import (
	"ABSC/internal/client"
	"ABSC/internal/database"
	"ABSC/internal/model"
	"fmt"
	"log"
)

type DeletionService struct {
	QbitClient *client.QBitClient
}

func NewDeletionService(qb *client.QBitClient) *DeletionService {
	return &DeletionService{
		QbitClient: qb,
	}
}

// PurgeSubscription 安全整季下架：清理 qB 种子/文件/RSS，并重置 SQLite 状态
func (s *DeletionService) PurgeSubscription(mikanID int, deleteFiles bool) error {
	// 1. 从 SQLite 查出该番剧的订阅记录
	var sub model.Subscription
	if err := database.DB.Where("mikan_id = ?", mikanID).First(&sub).Error; err != nil {
		return fmt.Errorf("未找到该番剧的订阅记录 (MikanID: %d): %w", mikanID, err)
	}
	log.Printf("启动整季下架清理流程: [%s] ...", sub.QBCategory)

	if err := s.QbitClient.Login(); err != nil {
		return fmt.Errorf("qBittorrent 登录失败: %w", err)
	}

	// 2. 查找属于该分类的所有种子 Hash 列表,并删除
	torrents, err := s.QbitClient.GetTorrents(200)
	if err == nil {
		var targetHashes []string
		for _, t := range torrents {
			if t.Category == sub.QBCategory {
				targetHashes = append(targetHashes, t.Hash)
			}
		}
		if len(targetHashes) > 0 {
			log.Printf("找到 %d 个属于该番剧的种子，准备执行安全删除 (同时删除种子和文件: %v) ...", len(targetHashes), deleteFiles)
			if err := s.QbitClient.DeleteTorrents(targetHashes, deleteFiles); err != nil {
				return fmt.Errorf("删除种子失败: %w", err)
			}
		} else {
			return fmt.Errorf("寻找番剧的种子时出错 (MikanID: %d)", mikanID)
		}
	}

	// 3. 删除对应的 RSS 订阅源和下载规则
	if sub.QBCategory != "" {
		log.Printf("正在移除 qB RSS 节点与规则: %s", sub.QBCategory)
		if err := s.QbitClient.RemoveRSSItem(sub.QBCategory); err != nil {
			return fmt.Errorf("删除 RSS 节点失败: %w", err)
		}
		if err := s.QbitClient.RemoveRSSRule(sub.QBCategory); err != nil {
			return fmt.Errorf("删除 RSS 规则失败: %w", err)
		}
	}

	// 4. 重置 SQLite 中的订阅状态
	err = database.DB.Model(&sub).Updates(map[string]interface{}{
		"status":       "none",
		"qb_category":  "",
		"rss_feed_url": "",
		"save_path":    "",
	}).Error
	if err != nil {
		return fmt.Errorf("重置订阅状态失败: %w", err)
	}
	return nil
}
