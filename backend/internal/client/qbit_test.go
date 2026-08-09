package client

import (
	"os"
	"testing"
)

func TestQBitPipelineReal(t *testing.T) {
	// 👈 请直接换成你软路由内网中真实的 qB 连接信息进行测试
	qbURL := "http://192.168.2.148:8080"
	username := "1"
	password := "1"

	t.Log("🚀 正在初始化 Go 强类型 qBittorrent 客户端...")
	client, err := NewQBitClient(qbURL, username, password)
	if err != nil {
		t.Fatalf("初始化客户端失败: %v", err)
	}

	// 1. 测试登录接口
	t.Log("🔐 步骤一：尝试登录验证并提取 SID Cookie...")
	if err := client.Login(); err != nil {
		t.Fatalf("❌ 登录失败: %v", err)
	}
	t.Log("✅ 登录成功，会话已激活")

	// 2. 模拟前端点按：订阅一部测试动漫
	testAnimeTitle := "无职转生S3"
	testSavePath := "/downloads/Series/Mushoku Tensei Jobless Reincarnation/Season 3"
	testRSS := "https://mikanani.me/RSS/Bangumi?bangumiId=3995&subgroupid=1252"
	// 3. 测试创建分类
	t.Logf("📂 步骤二：尝试在 qB 中创建测试分类 [%s] 并在 TrueNAS 规划路径...", testAnimeTitle)
	if err := client.CreateCategory(testAnimeTitle, testSavePath); err != nil {
		t.Fatalf("❌ 创建分类失败: %v", err)
	}

	// 4. 测试添加 RSS 订阅流
	t.Logf("📡 步骤三：将 Mikan RSS 链接注入 qB 树状目录...")
	if err := client.AddRSSFeed(testRSS, testAnimeTitle); err != nil {
		t.Fatalf("❌ 添加 RSS 失败: %v", err)
	}

	// 5. 测试构建强类型 RuleDefinition 过滤器并应用
	t.Log("🤖 步骤四：构建强类型 Rule 过滤器并推送给 qB 下载引擎...")
	ruleDef := RuleDefinition{
		Enabled:          true,
		AffectedFeeds:    []string{testRSS},
		AssignedCategory: "", // 保持与你 Python 脚本一致
		MustNotContain:   "繁日",
		TorrentParams: TorrentParams{
			Category:      testAnimeTitle, // 核心绑定：让该 RSS 匹配到的种子自动应用此分类
			OperatingMode: "AutoManaged",
		},
	}

	if err := client.SetRSSRule(testAnimeTitle, ruleDef); err != nil {
		t.Fatalf("❌ 配置自动化下载规则失败: %v", err)
	}

	t.Log("登录 -> 建分类 -> 加RSS -> 设过滤规则 全流水线原子测试在 Go 中完美跑通！")
}

// TestGetQbTorrents 测试获取种子列表
func TestGetQbTorrents(t *testing.T) {
	qbURL := getEnvOrDefault("QB_URL", "http://192.168.2.148:8080")
	qbUser := getEnvOrDefault("QB_USER", "admin")
	qbPass := getEnvOrDefault("QB_PASS", "adminadmin")
	tmdbAPI := getEnvOrDefault("TMDB_API_KEY", "")

	t.Log(qbURL, qbUser, qbPass, tmdbAPI)

	t.Log("🚀 正在初始化 Go 强类型 qBittorrent 客户端...")
	client, err := NewQBitClient(qbURL, qbUser, qbPass)
	if err != nil {
		t.Fatalf("初始化客户端失败: %v", err)
	}

	// 1. 测试登录接口
	t.Log("🔐 步骤一：尝试登录验证并提取 SID Cookie...")
	if err := client.Login(); err != nil {
		t.Fatalf("❌ 登录失败: %v", err)
	}
	t.Log("✅ 登录成功，会话已激活")

	// 2. 获取种子列表
	var torrentinfo []TorrentInfo
	torrentinfo, err = client.GetTorrents(10)
	if err != nil {
		t.Fatalf("❌ 获取种子列表失败: %v", err)
	}
	for _, torrent := range torrentinfo {
		t.Log(torrent.SavePath)
	}
}

// getEnvOrDefault 辅助函数：读取环境变量，为空则使用默认 fallback 值
func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
