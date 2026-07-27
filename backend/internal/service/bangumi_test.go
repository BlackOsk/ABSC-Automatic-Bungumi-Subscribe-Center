package service

import (
	"ABSC/internal/database"
	"ABSC/internal/model"
	"ABSC/internal/scraper"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
)

func TestExt(t *testing.T) {
	// 测试 Ext 函数
	ext := filepath.Ext(".s/df/ssskkjiuyy/123-456-789.mp3")
	fmt.Println(ext) // 输出: 123-456-789
}

func TestGetEps(t *testing.T) {
	var episodePatterns = []*regexp.Regexp{
		regexp.MustCompile(`\s([0-9]{2})\s`),
		regexp.MustCompile(`-([0-9]{2})-`),
		regexp.MustCompile(`\[([0-9]{2})\]`),
		regexp.MustCompile(`\[([0-9]{2})v[0-9]\]`),
		regexp.MustCompile(`\s([0-9]{2})v[0-9]\s`),
		regexp.MustCompile(`(?i)E([0-9]{1,2})`), // 容错支持 E01、E1 格式
	}

	fileName := "E3"

	for _, pattern := range episodePatterns {
		matches := pattern.FindStringSubmatch(fileName)
		if len(matches) >= 2 {
			absEp, err := strconv.Atoi(matches[1])
			if err == nil {
				t.Log(absEp)
			}
			t.Log(absEp)
		}
	}

	//t.Logf("%02d", 6)

}

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

func TestCalculateAutoOffset(t *testing.T) {
	// 1. 环境参数准备：从环境变量读取 API Key，避免把密钥硬编码提交到 Git
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		// 你也可以在这里临时填入你的 Key 进行本地测试
		apiKey = "YOUR_TEMPORARY_TMDB_API_KEY"
	}

	if apiKey == "YOUR_TEMPORARY_TMDB_API_KEY" || apiKey == "" {
		t.Skip("跳过测试：未检测到有效的 TMDB_API_KEY")
	}

	// 3. 初始化客户端（如果本地电脑需要梯子，填入代理地址，不需要则留空 ""）
	proxy := "" //"http://127.0.0.1:7897"
	client := scraper.NewTMDBClient(apiKey, proxy)
	bangumiService := NewBanbumiService(client)
	testSeason := 3

	testOffset := bangumiService.CalculateAutoOffset(94664, testSeason)

	t.Logf("计算第 %d 季的集数偏移为 %d", testSeason, testOffset)
}
