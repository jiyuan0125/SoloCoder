package main

import (
	"blood-management-system/config"
	"blood-management-system/handlers"
	"blood-management-system/models"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	port := config.GetPort()

	db, err := models.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := db.AutoMigrate(
		&models.Donor{},
		&models.HealthCheck{},
		&models.BloodCollection{},
		&models.TestRecord{},
		&models.Inventory{},
		&models.BloodRequest{},
		&models.WorkflowItem{},
		&models.SafetyStock{},
	); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	models.InitSafetyStock(db)

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

	api := r.Group("/api")
	{
		donorHandler := handlers.NewDonorHandler(db)
		api.POST("/donors", donorHandler.Register)
		api.GET("/donors/:id", donorHandler.Get)
		api.GET("/donors", donorHandler.List)

		collectionHandler := handlers.NewCollectionHandler(db)
		api.POST("/collections", collectionHandler.Create)
		api.GET("/collections/:id", collectionHandler.Get)
		api.GET("/collections", collectionHandler.List)

		testHandler := handlers.NewTestHandler(db)
		api.POST("/tests", testHandler.Record)
		api.GET("/tests/:collection_id", testHandler.Get)
		api.PUT("/collections/:id/status", collectionHandler.UpdateStatus)

		inventoryHandler := handlers.NewInventoryHandler(db)
		api.GET("/inventory", inventoryHandler.List)
		api.GET("/inventory/:blood_type/:product_type", inventoryHandler.GetByType)
		api.POST("/inventory/freeze", inventoryHandler.Freeze)
		api.POST("/inventory/unfreeze", inventoryHandler.Unfreeze)

		requestHandler := handlers.NewRequestHandler(db)
		api.POST("/requests", requestHandler.Create)
		api.GET("/requests/:id", requestHandler.Get)
		api.GET("/requests", requestHandler.List)
		api.PUT("/requests/:id/process", requestHandler.Process)

		statsHandler := handlers.NewStatsHandler(db)
		api.GET("/stats/monthly-collection", statsHandler.MonthlyCollection)
		api.GET("/stats/monthly-scrap", statsHandler.MonthlyScrap)
		api.GET("/stats/hospital-ranking", statsHandler.HospitalRanking)
		api.GET("/stats/inventory-turnover", statsHandler.InventoryTurnover)
		api.GET("/stats/dashboard", statsHandler.Dashboard)
	}

	fmt.Printf("Blood Management System starting on port %s...\n", port)
	if err := r.Run(":" + port); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
