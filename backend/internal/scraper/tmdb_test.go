package scraper

import (
	"log"
	"os"
	"testing"
)

func TestTmdbMessage(t *testing.T) {

	// 从环境变量读取 API Key，避免把密钥硬编码提交到 Git
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		// 你也可以在这里临时填入你的 Key 进行本地测试
		apiKey = "YOUR_TEMPORARY_TMDB_API_KEY"
	}

	if apiKey == "YOUR_TEMPORARY_TMDB_API_KEY" || apiKey == "" {
		t.Skip("跳过测试：未检测到有效的 TMDB_API_KEY")
	}

	// 初始化客户端（如果本地电脑需要梯子，填入代理地址，不需要则留空 ""）
	proxy := ""
	client := NewTMDBClient(apiKey, proxy)

	// 测试搜索当下热门的动漫
	testTitle := "小书痴的下克上"
	t.Logf("开始在 TMDB 中搜索动漫: %s ...", testTitle)

	result, err := client.SearchAnime(testTitle)
	if err != nil {
		t.Fatalf("❌ 刮削失败: %v", err)
	}

	// 断言验证返回的信息
	t.Logf("🎉 刮削成功！")
	t.Logf("TMDB ID: %d", result.ID)
	t.Logf("标准中文名: %s", result.Name)
	t.Logf("原始名: %s", result.OriginalName)
	t.Logf("高清海报路径: https://image.tmdb.org/t/p/w500%s", result.PosterPath)
	t.Logf("中文剧情简介: %s", result.Overview)
	t.Logf("首播日期: %s", result.FirstAirDate)
}

func TestTMDBTVDetailsAndOffsetLogic(t *testing.T) {

	// 从环境变量读取 API Key，避免把密钥硬编码提交到 Git
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		// 你也可以在这里临时填入你的 Key 进行本地测试
		apiKey = "YOUR_TEMPORARY_TMDB_API_KEY"
	}

	if apiKey == "YOUR_TEMPORARY_TMDB_API_KEY" || apiKey == "" {
		t.Skip("跳过测试：未检测到有效的 TMDB_API_KEY")
	}

	// 初始化客户端（如果本地电脑需要梯子，填入代理地址，不需要则留空 ""）
	proxy := ""
	client := NewTMDBClient(apiKey, proxy)

	testID := 94664

	t.Logf("正在测试请求 TMDB ID: %d 的剧集详情...", testID)
	details, err := client.GetTVDetails(testID)
	if err != nil {
		t.Fatalf("❌ 刮削失败: %v", err)
	}

	// 断言验证返回的信息
	t.Logf("总集数 %v", details.NumberOfEpisodes)
	t.Logf("总季数 %v", details.NumberOfSeasons)

	targetSeason := 3
	autoOffset := 0

	// 2. 遍历所有季，计算目标季之前的总集数
	for _, season := range details.Seasons {
		if season.SeasonNumber > 0 && season.SeasonNumber < targetSeason {
			autoOffset += season.EpisodeCount
			log.Printf("自动推导累加：第 %d 季有 %d 集", season.SeasonNumber, season.EpisodeCount)
		}

	}
	log.Printf("动漫 TMDB_ID: %d，当前订阅第 %d 季，推导出的集数偏移量为: 减去 %d 集", testID, targetSeason, autoOffset)
}
