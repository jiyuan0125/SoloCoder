package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"organdonation/internal/handler"
	"organdonation/internal/repository"
	"organdonation/internal/service"
)

func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	port := flag.String("port", "8080", "server port")
	flag.Parse()
	return ":" + *port
}

func startScheduler(svc *service.Service) {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			now := time.Now()
			if now.Hour() == 0 {
				svc.UpgradeUrgencyLevel(context.Background())
				log.Println("紧急程度升级检查完成")
			}
			if now.Hour() == 1 && now.Weekday() == time.Monday {
				svc.GenerateWeeklyReport(context.Background())
				log.Println("每周报告已生成")
			}
			svc.CheckColdIschemiaTimeout(context.Background())
		}
	}()
}

func main() {
	port := getPort()

	repo := repository.New()
	svc := service.New(repo)
	h := handler.New(svc)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	h.RegisterRoutes(r)

	startScheduler(svc)

	log.Printf("服务器启动在端口 %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

func init() {
	if val := os.Getenv("GIN_MODE"); val == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	if v := os.Getenv("PORT"); v != "" {
		if _, err := strconv.Atoi(v); err != nil {
			_ = os.Setenv("PORT", "8080")
		}
	}
}
