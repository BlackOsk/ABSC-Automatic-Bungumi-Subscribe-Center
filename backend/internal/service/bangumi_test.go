package service

import (
	"ABSC/internal/database"
	"ABSC/internal/model"
	"ABSC/internal/scraper"
	"os"
	"testing"
)

func TestSyncCurrentQuarterBangumi(t *testing.T) {

	// 1. 环境参数准备：从环境变量读取 API Key，避免把密钥硬编码提交到 Git
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		// 你也可以在这里临时填入你的 Key 进行本地测试
		apiKey = "YOUR_TEMPORARY_TMDB_API_KEY"
	}

	if apiKey == "YOUR_TEMPORARY_TMDB_API_KEY" || apiKey == "" {
		t.Skip("跳过测试：未检测到有效的 TMDB_API_KEY")
	}

	// 2. 初始化临时sqlite
	testDBPath := "integration_bangumi.db"
	defer os.Remove(testDBPath)
	database.InitDB(testDBPath)

	// 3. 初始化客户端（如果本地电脑需要梯子，填入代理地址，不需要则留空 ""）
	proxy := "" //"http://127.0.0.1:7897"
	client := scraper.NewTMDBClient(apiKey, proxy)
	bangumiService := NewBanbumiService(client)

	// 4. 执行同步函数
	t.Log("[TestSyncCurrentQuarterBangumi] 🔄 开始测试同步当季新番流...")
	err := bangumiService.SyncCurrentQuarterBangumi()

	if err != nil {
		t.Fatalf("❌ 同步当季新番流失败: %v", err)
	}

	// 5. 验证数据库中是否有数据
	var totalCount int64
	var enrichedCount int64

	database.DB.Model(&model.BangumiMetadata{}).Count(&totalCount)
	database.DB.Model(&model.BangumiMetadata{}).Where("tmdb_id IS NOT NULL").Count(&enrichedCount)

	t.Logf("Mikan 发现总番剧 %v 部", totalCount)
	t.Logf("成功通过 TMDB 染色并补全高清海报的番剧数: %d 部", enrichedCount)

	if totalCount == 0 {
		t.Error("❌ 错误：本地数据库空空如也，未完成基本录入")
	}

}
