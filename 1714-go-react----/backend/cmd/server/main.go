package main

import (
	"flag"
	"log"
	"time"

	"health-supervision-system/internal/config"
	"health-supervision-system/internal/handlers"
	"health-supervision-system/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	portFlag := flag.String("port", "", "Server port")
	flag.Parse()

	cfg := config.LoadConfig(*portFlag)

	if err := storage.InitDB(cfg.DBPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	go startNoticeExpiryChecker()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api")
	{
		units := api.Group("/units")
		{
			units.GET("", handlers.ListUnits)
			units.GET("/warnings", handlers.GetUnitsWithLicenseWarning)
			units.GET("/:id", handlers.GetUnit)
			units.POST("", handlers.CreateUnit)
			units.PUT("/:id", handlers.UpdateUnit)
			units.DELETE("/:id", handlers.DeleteUnit)
		}

		inspections := api.Group("/inspections")
		{
			inspections.GET("/templates", handlers.GetInspectionTemplates)
			inspections.GET("", handlers.ListInspections)
			inspections.GET("/:id", handlers.GetInspection)
			inspections.POST("", handlers.CreateInspection)
			inspections.PUT("/:id/items", handlers.UpdateInspectionItems)
			inspections.DELETE("/:id", handlers.DeleteInspection)
		}

		opinions := api.Group("/opinions")
		{
			opinions.GET("", handlers.ListOpinions)
			opinions.GET("/:id", handlers.GetOpinion)
			opinions.POST("", handlers.IssueOpinion)
			opinions.POST("/:id/review", handlers.ReviewOpinion)
		}

		penalties := api.Group("/penalties")
		{
			penalties.GET("/fine-ranges", handlers.GetFineRanges)
			penalties.GET("", handlers.ListPenalties)
			penalties.GET("/:id", handlers.GetPenalty)
			penalties.POST("", handlers.IssuePenalty)
		}

		notices := api.Group("/notices")
		{
			notices.GET("", handlers.ListNotices)
			notices.GET("/:id", handlers.GetNotice)
			notices.POST("", handlers.PublishNotice)
			notices.PUT("/:id/expiry", handlers.UpdateNoticeExpiry)
			notices.POST("/check-expiry", handlers.CheckAndExpireNotices)
			notices.DELETE("/:id", handlers.DeleteNotice)
		}

		audit := api.Group("/audit")
		{
			audit.GET("", handlers.ListAuditLogs)
			audit.GET("/:id", handlers.GetAuditLog)
			audit.POST("", handlers.BlockAuditModification)
			audit.PUT("/:id", handlers.BlockAuditModification)
			audit.DELETE("/:id", handlers.BlockAuditModification)
		}
	}

	log.Printf("Server starting on %s", cfg.GetAddress())
	if err := r.Run(cfg.GetAddress()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func startNoticeExpiryChecker() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		handlers.ExpireNotices()
	}
}
