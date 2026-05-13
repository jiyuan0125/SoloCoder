package main

import (
	"asset-depreciation/database"
	"asset-depreciation/router"
	"asset-depreciation/service"
	"log"
	"net/http"
	"time"
)

const (
	port     = ":9700"
	dbPath   = "./assets.db"
)

func main() {
	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.DB.Close()

	log.Println("Database initialized successfully")

	go startScheduler()

	r := router.NewRouter()

	log.Printf("Starting server on port %s...", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func startScheduler() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		<-ticker.C
		log.Println("Running scheduled depreciation check...")
		if err := service.ProcessMonthlyDepreciation(); err != nil {
			log.Printf("Error processing monthly depreciation: %v", err)
		}
	}
}
