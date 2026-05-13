package main

import (
	"audit-log/archive"
	"audit-log/database"
	"audit-log/handlers"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	if thresholdEnv := os.Getenv("STORAGE_THRESHOLD_GB"); thresholdEnv != "" {
		if gb, err := strconv.ParseInt(thresholdEnv, 10, 64); err == nil && gb > 0 {
			archive.StorageThreshold = gb * 1024 * 1024 * 1024
		}
	}

	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	if err := archive.InitArchive(); err != nil {
		log.Fatalf("Failed to initialize archive: %v", err)
	}

	archive.StartArchiveChecker()

	r := gin.Default()

	api := r.Group("/api/v1/audit-logs")
	{
		api.POST("", handlers.CreateAuditLog)
		api.GET("", handlers.QueryAuditLogs)
		api.GET("/export/csv", handlers.ExportAuditLogsCSV)
		api.GET("/:id", handlers.GetAuditLogByID)

		api.PUT("", handlers.MethodNotAllowed)
		api.PUT("/:id", handlers.MethodNotAllowed)
		api.DELETE("", handlers.MethodNotAllowed)
		api.DELETE("/:id", handlers.MethodNotAllowed)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "not found"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "9802"
	}

	log.Println("Audit log service starting on port", port, "...")
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
