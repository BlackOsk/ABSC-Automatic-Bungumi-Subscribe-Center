package database

import (
	"ABSC/internal/model"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化数据库连接并自动迁移表结构

func InitDB(dbPath string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ [database.InitDB] 数据库连接失败: %v", err)
	}

	log.Println("📂 [database.InitDB] SQLite 数据库连接成功...")

	// 自动迁移表结构
	err = DB.AutoMigrate(&model.BangumiMetadata{}, &model.Subscription{}, &model.EpisodeOffset{}, &model.MikanSubgroupResource{}, &model.MikanEpisode{})
	if err != nil {
		log.Fatalf("❌ [database.InitDB] 数据库表结构迁移失败: %v", err)
	}
	log.Println("✅ [database.InitDB] 数据库表结构同步/建立成功!")
}
