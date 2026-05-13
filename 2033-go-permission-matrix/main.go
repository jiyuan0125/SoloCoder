package main

import (
	"log"
	"net/http"
	"os"
	"permission-matrix/handler"
	"permission-matrix/router"
	"permission-matrix/service"
	"permission-matrix/storage"
	"time"
)

func main() {
	dbPath := "./permission.db"
	if envPath := os.Getenv("DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	if err := storage.InitDB(dbPath); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer storage.CloseDB()

	log.Printf("数据库已初始化: %s", dbPath)

	svc := service.NewPermissionService()
	h := handler.NewHandler(svc)
	r := router.NewRouter(h)

	port := "9902"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("权限矩阵服务启动在端口 %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务启动失败: %v", err)
	}
}
