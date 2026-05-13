package main

import (
	"log"
	"time"
)

const (
	ServerPort  = 9702
	DBPath      = "./notifications.db"
)

func main() {
	if err := InitDB(DBPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	go startRetryScheduler()

	if err := StartServer(ServerPort); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func startRetryScheduler() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("Retry scheduler started (checking every minute)")

	for range ticker.C {
		if err := RetryPendingNotifications(); err != nil {
			log.Printf("Retry scheduler error: %v", err)
		}
	}
}
