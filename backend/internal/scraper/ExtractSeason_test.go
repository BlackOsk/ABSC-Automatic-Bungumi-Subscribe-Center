package scraper

import (
	"ABSC/internal/model"
	"log"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestExtractSeason(t *testing.T) {
	var DB *gorm.DB
	// 现成的数据库文件路径
	dbPath := "..\\service\\integration_bangumi.db"

	var err error
	// 使用 sqlite.Open 打开现有的数据库文件
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("无法连接到现有的数据库文件，错误: %v", err)
	}

	log.Println("成功连接到现有的 SQLite 数据库！")

	var uncompletedBangumis []model.BangumiMetadata

	err = DB.Model(&model.BangumiMetadata{}).
		Where("tmdb_id IS NULL").
		Find(&uncompletedBangumis).Error
	if err != nil {
		t.Logf("查询未完善的动漫失败: %v", err)
	}

	for _, bangumi := range uncompletedBangumis {
		cleanedTitle, season := cleanTitleAndExtractSeason(bangumi.TitleCN)
		t.Logf("\n 原始标题: %s \n 清理后标题: %s \n 提取季数: %d", bangumi.TitleCN, cleanedTitle, season)
	}
}
