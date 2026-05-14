package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"log-aggregator/cleaner"
	"log-aggregator/db"
	"log-aggregator/handler"
)

const (
	port    = ":8080"
	dbPath  = "./logs.db"
)

func main() {
	if err := db.Init(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	logCleaner := cleaner.NewCleaner()
	logCleaner.Start()
	defer logCleaner.Stop()

	router := gin.Default()

	router.GET("/health", handler.HealthCheck)

	api := router.Group("/api/v1")
	{
		api.POST("/logs", handler.SubmitLog)
		api.GET("/logs", handler.QueryLogs)
		api.GET("/stats/daily", handler.GetDailyStats)
		api.GET("/config/retention", handler.GetRetentionConfig)
		api.PUT("/config/retention", handler.UpdateRetentionConfig)
	}

	go func() {
		log.Printf("Server starting on port %s", port)
		if err := router.Run(port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
}
