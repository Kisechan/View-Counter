package main

import (
	"log"
	"time"

	"view-counter/config"
	"view-counter/database"
	"view-counter/handler"
	"view-counter/middleware"
	"view-counter/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	cfg := config.New()

	// 初始化数据库
	db, err := database.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Database initialized successfully")

	// 初始化服务层
	counterService := service.NewCounterService(db)

	// 初始化处理器
	viewsHandler := handler.NewViewsHandler(counterService)

	// 初始化 Gin 引擎
	router := gin.Default()

	// 初始化速率限制中间件 (每分钟最多60次请求)
	rateLimiter := middleware.NewRateLimiter(60, time.Minute)

	// 创建 API 路由组并应用中间件
	api := router.Group("/api")
	api.Use(rateLimiter.Middleware()) // 将中间件应用于 /api 组下的所有路由
	{
		api.POST("/view", viewsHandler.IncrementView)
		api.GET("/view", viewsHandler.GetView)
	}

	log.Println("Listening on :8081")
	// 启动服务
	if err := router.Run(":8081"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
