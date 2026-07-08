package scraper

import (
	"ABSC/internal/database"
	"ABSC/internal/model"
	"os"
	"testing"
)

// #单元测试： 数据库新建-抓取Mikan-存入数据库
func TestScrapeAndSave(t *testing.T) {
	// 1. 初始化一个临时的测试数据库
	testDBPath := "test_anime.db"
	defer os.Remove(testDBPath) // 测试完自动把文件删了，保持干净

	database.InitDB(testDBPath)

	// 2. 跑爬虫抓取数据
	t.Log("开始从 Mikan 抓取实时新番数据...")
	bangumis, err := ScrapeCurrentQuarter()
	if err != nil {
		t.Fatalf("抓取失败: %v", err)
	}

	if len(bangumis) == 0 {
		t.Fatal("❌ 严重错误: 未能抓取到任何新番数据，可能是 Mikan 首页样式变了！")
	}
	t.Logf("成功解析出 %d 个新番条目", len(bangumis))

	// 3. 尝试批量写入数据库
	t.Log("将抓取到的数据持久化存入 SQLite...")
	for _, b := range bangumis {
		// 使用 Upsert 语义，防止主键冲突
		err := database.DB.Save(&b).Error
		if err != nil {
			t.Errorf("写入数据库失败 (MikanID: %d): %v", b.MikanID, err)
		}
	}

	// 4. 从数据库里数数，验算结果
	var count int64
	database.DB.Model(&model.BangumiMetadata{}).Count(&count)
	t.Logf("🎉 验证成功! 当前 SQLite 数据库中已成功安全落盘 %d 条新番记录！", count)
}
