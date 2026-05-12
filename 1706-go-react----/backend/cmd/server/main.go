package main

import (
	"fmt"
	"hospital-pharmacy/pkg/config"
	"hospital-pharmacy/pkg/database"
	"hospital-pharmacy/pkg/handlers"
	"log"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if len(os.Args) > 1 {
		if p, err := strconv.Atoi(os.Args[1]); err == nil {
			cfg.Port = p
		}
	}

	if err := database.Init(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	handlers.SetupDrugRoutes(r)
	handlers.SetupStockRoutes(r)
	handlers.SetupPrescriptionRoutes(r)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Printf("Server starting on port %d...", cfg.Port)
	if err := r.Run(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
