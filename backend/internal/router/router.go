package router

import (
	"ABSC/internal/handler"
	"ABSC/internal/service"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	bangumiSvc *service.BangumiService,
	subSvc *service.SubscriptionService,
	delSvc *service.DeletionService,

) *gin.Engine {
	r := gin.Default()

	// 1. 配置 CORS 跨域中间件 (方便 React 开发服务 5173 端口无缝请求 Go 8899)
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:   []string{"Content-Length"},
		MaxAge:          12 * time.Hour,
	}))

	// 2. 初始化 Handlers
	bangumiHandler := handler.NewBangumiHandler(bangumiSvc)
	subHandler := handler.NewSubscriptionHandler(subSvc, delSvc)
	offsetHandler := handler.NewOffsetHandler()

	// 3. 注册 API 路由组 v1
	v1 := r.Group("/api/v1")
	{
		// 番剧模块
		bangumiGroup := v1.Group("/bangumi")
		{
			bangumiGroup.GET("/current", bangumiHandler.GetCurrent)
			bangumiGroup.GET("/:mikan_id/detail", bangumiHandler.GetDetail)
			bangumiGroup.POST("/sync", bangumiHandler.SyncTMDB)
		}

		// 订阅模块
		subGroup := v1.Group("/subscriptions")
		{
			subGroup.GET("/list", subHandler.List)
			subGroup.POST("", subHandler.Subscribe)
			subGroup.DELETE("/:mikan_id", subHandler.Purge)
		}

		// 偏移量管理与改名预览模块
		offsetGroup := v1.Group("/offsets")
		{
			offsetGroup.GET("/:mikan_id", offsetHandler.GetOffsets)
			offsetGroup.POST("", offsetHandler.UpdateOffset)
			offsetGroup.POST("/preview", offsetHandler.PreviewRename)
		}
	}

	// 在 router.go 中追加静态资源托管
	r.Static("/assets", "./dist/assets")
	r.StaticFile("/favicon.ico", "./dist/favicon.ico")

	// 兼容 SPA 单页应用路由：任何未匹配到 API 的请求，均返回 index.html
	r.NoRoute(func(c *gin.Context) {
		c.File("./dist/index.html")
	})

	return r

}
