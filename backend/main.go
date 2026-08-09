package main

import (
	"ABSC/internal/client"
	"ABSC/internal/database"
	"ABSC/internal/router"
	"ABSC/internal/scraper"
	"ABSC/internal/service"
	"log"
	"os"
	"time"
)

// getEnvOrDefault 辅助函数：读取环境变量，为空则使用默认 fallback 值
func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func main() {
	log.Println("==================================================")
	log.Println("Automatic Banbumi Subscription Center Starting ...")
	log.Println("==================================================")

	// 1. 读取环境配置 (支持环境变量覆盖，无环境变量时使用软路由默认值)
	dbPath := getEnvOrDefault("ABSC_DB_PATH", "absc.db")
	qbURL := getEnvOrDefault("QB_URL", "http://192.168.2.148:8080")
	qbUser := getEnvOrDefault("QB_USER", "admin")
	qbPass := getEnvOrDefault("QB_PASS", "adminadmin")
	tmdbKey := getEnvOrDefault("TMDB_API_KEY", "")
	proxyURL := getEnvOrDefault("PROXY_URL", "")
	seriesDir := getEnvOrDefault("SERIES_DIR", "/downloads/Series")
	incompleteDir := getEnvOrDefault("INCOMPLETE_DIR", "/downloads/incomplete")
	serverPort := getEnvOrDefault("SERVER_PORT", "8899")

	// 2. 初始化数据库连接
	log.Printf("📦 初始化 SQLite 数据库文件: %s", dbPath)
	database.InitDB(dbPath)

	// 3. 初始化底层客户端 (TMDB & qBittorrent)
	log.Println("初始化外部服务客户端: TMDB")
	tmdbClient := scraper.NewTMDBClient(tmdbKey, proxyURL)
	log.Println("初始化外部服务客户端: qBittorrent")
	qbClient, err := client.NewQBitClient(qbURL, qbUser, qbPass)
	if err != nil {
		log.Fatalf("初始化 qBittorrent 客户端失败: %v", err)
	}

	// 4. 初始化核心业务 Service 句柄 (依赖注入)
	bangumiSvc := service.NewBangumiService(tmdbClient)
	subSvc := service.NewSubscriptionService(qbClient, bangumiSvc, seriesDir)
	delSvc := service.NewDeletionService(qbClient)
	renameSvc := service.NewRenameService(qbClient, seriesDir, incompleteDir)

	// 5. 启动后台定时任务（Goroutines 守护进程）
	StartBackgroundTasks(bangumiSvc, renameSvc)

	// 6. 启动 Gin HTTP Server
	r := router.SetupRouter(bangumiSvc, subSvc, delSvc)

	log.Printf("ABSC已成功启动！HTTP 服务监听端口: :%s", serverPort)
	if err := r.Run(":" + serverPort); err != nil {
		log.Fatalf("❌ HTTP 服务异常终止: %v", err)
	}

}

// startBackgroundTasks 启动常驻后台协程任务
func StartBackgroundTasks(bangumiSvc *service.BangumiService, renameSvc *service.RenameService) {
	// -------------------------------------------------------------------------
	// 任务 A：常驻智能重命名与跨季路径纠偏 Worker (每 10 分钟运行一次)
	// -------------------------------------------------------------------------
	go func() {
		// 启动时延迟 10 秒后执行第一次扫描，避开服务刚启动时的网络拥堵
		time.Sleep(10 * time.Second)
		log.Println("[BackgroundTask] 启动后台智能重命名 Worker...")

		// 立即执行一次
		if err := renameSvc.ExecuteRenameTask(25); err != nil {
			log.Printf("[BackgroundTask] 首次执行智能重命名任务失败: %v", err)
		}

		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("[BackgroundTask] 定时触发执行智能重命名任务...")
			if err := renameSvc.ExecuteRenameTask(25); err != nil {
				log.Printf("[BackgroundTask] 定时执行智能重命名任务失败: %v", err)
			}

		}
	}()
	// -------------------------------------------------------------------------
	// 任务 B：定时从 Mikan + TMDB 增量同步当季新番数据 (每 6 小时运行一次)
	// -------------------------------------------------------------------------
	go func() {
		// 启动时延迟 3 秒立即触发一次全量增量刷新
		time.Sleep(3 * time.Second)
		log.Println("[StartBackgroundTasks] 启动首屏增量新番元数据刷新...")
		if err := bangumiSvc.SyncCurrentQuarterBangumi(); err != nil {
			log.Printf("[StartBackgroundTasks] 首屏新番数据同步异常: %v", err)
		}
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("[StartBackgroundTasks] 定时触发执行新番数据同步...")
			if err := bangumiSvc.SyncCurrentQuarterBangumi(); err != nil {
				log.Printf("[StartBackgroundTasks] 定时执行新番数据同步失败: %v", err)
			}
		}
	}()

}
