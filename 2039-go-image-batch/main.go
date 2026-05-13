package main

import (
	"image-batch/config"
	"image-batch/handlers"
	"image-batch/repositories"
	"image-batch/services"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	taskRepo, err := repositories.NewTaskRepository()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer taskRepo.Close()

	imageService := services.NewImageService()
	zipService := services.NewZipService()
	batchService := services.NewBatchService(imageService, zipService, taskRepo)

	imageHandler := handlers.NewImageHandler(batchService, imageService)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.MaxMultipartMemory = int64(config.MaxUploadSize)

	api := r.Group("/api")
	{
		api.GET("/health", imageHandler.Health)
		api.GET("/formats", imageHandler.GetSupportedFormats)
		api.GET("/operations", imageHandler.GetOperations)
		api.POST("/process", imageHandler.ProcessBatch)
		api.GET("/tasks/:id", imageHandler.GetTaskStatus)
		api.GET("/tasks/:id/download", imageHandler.DownloadZip)
	}

	log.Printf("Image Batch Processing Server starting on port %s", config.ServerPort)
	log.Printf("Supported formats: %s", config.SupportedFormats)

	if err := r.Run(":" + config.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
