package model

import "time"

// BangumiMetadata 对应新番元数据表
type BangumiMetadata struct {
	MikanID       int       `gorm:"primaryKey;column:mikan_id"`
	TMDBID        *int      `gorm:"unique;column:tmdb_id"`
	TitleCN       string    `gorm:"not null;column:title_cn"`
	TitleRaw      string    `gorm:"column:title_raw"`
	Overview      string    `gorm:"column:overview"`
	PosterPath    string    `gorm:"column:poster_path"`
	AirDate       string    `gorm:"column:air_date"`
	BroadcastDay  int       `gorm:"column:broadcast_day"` // 0-6 代表周日-周六
	TotalEpisodes int       `gorm:"default:0;column:total_episodes"`
	CurrentSeason int       `gorm:"default:1;column:current_season"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

// Subscription 对应本地订阅状态表
type Subscription struct {
	MikanID    int             `gorm:"primaryKey;column:mikan_id"`
	Status     string          `gorm:"default:'none';column:status"` // none, subscribing, archived
	QBCategory string          `gorm:"column:qb_category"`
	RSSFeedURL string          `gorm:"column:rss_feed_url"`
	SavePath   string          `gorm:"column:save_path"`
	CreatedAt  time.Time       `gorm:"autoCreateTime"`
	UpdatedAt  time.Time       `gorm:"autoUpdateTime"`
	Metadata   BangumiMetadata `gorm:"foreignKey:MikanID;constraint:OnDelete:CASCADE;"`
}

// EpisodeOffset 对应集数偏移配置表
type EpisodeOffset struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	MikanID     int       `gorm:"uniqueIndex:idx_mikan_season;column:mikan_id"`
	Season      int       `gorm:"default:1;uniqueIndex:idx_mikan_season;column:season"` // idx_mikan_season，联合索引，确保同一季的偏移配置唯一
	OffsetValue int       `gorm:"default:0;column:offset_value"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// MikanSubgroupResource 承载单个字幕组针对该番剧的RSS订阅源
type MikanSubgroupResource struct {
	MikanID      int            `gorm:"primaryKey;column:mikan_id"`    // 番剧ID
	SubgroupID   int            `gorm:"primaryKey;column:subgroup_id"` // 字幕组ID
	SubgroupName string         `gorm:"not null;column:subgroup_name"` // 字幕组中文名
	RSSURL       string         `gorm:"column:rss_url"`                // 字幕组专属RSS链接
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
	Episodes     []MikanEpisode `gorm:"foreignKey:MikanID,SubgroupID;references:MikanID,SubgroupID;constraint:OnDelete:CASCADE;"`
}

// MikanEpisode 记录当前字幕组已经发布过的历史单集种子文件列表
type MikanEpisode struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	MikanID     int       `gorm:"uniqueIndex:idx_ep_unique;column:mikan_id"`
	SubgroupID  int       `gorm:"uniqueIndex:idx_ep_unique;column:subgroup_id"`
	Title       string    `gorm:"uniqueIndex:idx_ep_unique;column:title"` // 种子原始长文件名
	Size        string    `gorm:"column:size"`                            // 文件大小 (如 450.2MB)
	PublishTime string    `gorm:"column:publish_time"`                    // 发布时间 (如 2026/07/18 20:30)
	Magnet      string    `gorm:"column:magnet"`                          // 磁力链接
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}
