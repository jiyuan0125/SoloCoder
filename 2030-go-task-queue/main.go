package main

import (
	"log"
	"task-queue/internal/database"
	"task-queue/internal/handler"
	"task-queue/internal/queue"
	"time"

	"github.com/gin-gonic/gin"
)

func startBackgroundJobs() {
	timeoutTicker := time.NewTicker(1 * time.Minute)
	cleanupTicker := time.NewTicker(1 * time.Hour)

	go func() {
		for range timeoutTicker.C {
			if err := queue.ProcessTimeouts(); err != nil {
				log.Printf("Failed to process timeouts: %v", err)
			}
		}
	}()

	go func() {
		for range cleanupTicker.C {
			if err := queue.CleanupOldTasks(7); err != nil {
				log.Printf("Failed to cleanup old tasks: %v", err)
			}
		}
	}()
}

func main() {
	err := database.InitDB("task_queue.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	startBackgroundJobs()

	r := gin.Default()

	r.GET("/health", handler.Health)
	r.POST("/tasks", handler.SubmitTask)
	r.GET("/tasks/fetch", handler.FetchTasks)
	r.GET("/tasks/:id", handler.GetTask)
	r.POST("/tasks/:id/complete", handler.CompleteTask)

	log.Println("Starting server on :9900")
	if err := r.Run(":9900"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
