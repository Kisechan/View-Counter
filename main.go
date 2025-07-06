package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"view-counter/config"
	"view-counter/database"
	"view-counter/handler"
	"view-counter/service"
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
	fmt.Println("Database initialized successfully")

	// 初始化归档服务 每天执行一次，保留 30 天数据
	archiver := service.NewArchiver(db, "./views_archive.db", 24*time.Hour, 30*24*time.Hour)
	go archiver.Start()

	// 初始化服务层
	counterService := service.NewCounterService(db)

	// 初始化处理器
	viewsHandler := handler.NewViewsHandler(counterService)

	// 设置路由
	http.HandleFunc("/api/view", viewsHandler.HandleViewsRequest)

	log.Println("Listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
