package main

import (
	"log"

	"trial-management-system/internal/api/routes"
	"trial-management-system/internal/config"
	"trial-management-system/internal/database"
)

func main() {
	cfg := config.Load()

	database.Init(cfg.DBPath)

	r := routes.SetupRouter()

	log.Printf("Server starting on port " + cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
