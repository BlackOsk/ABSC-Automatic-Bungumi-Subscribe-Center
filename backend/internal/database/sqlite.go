package database

import (
	"ABSC/internal/model"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dbPath string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	log.Println("📂 SQLite 数据库连接成功...")

	// 自动迁移表结构
	err = DB.AutoMigrate(&model.BangumiMetadata{}, &model.Subscription{}, &model.EpisodeOffset{})
	if err != nil {
		log.Fatalf("❌ 数据库迁移失败: %v", err)
	}
	log.Println("✅ 数据库表结构同步/建立成功!")
}
