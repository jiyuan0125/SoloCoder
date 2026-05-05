package main

import (
	"go-live-chat/internal/server"
	"log"
	"net/http"
	"time"
)

func main() {
	store := server.NewStore()
	service := server.NewService(store)
	handler := server.NewHandler(service)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	go startBackgroundJobs(service)

	log.Println("Live chat server starting on :8739...")
	if err := http.ListenAndServe(":8739", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func startBackgroundJobs(service *server.Service) {
	queueTicker := time.NewTicker(30 * time.Second)
	sessionTicker := time.NewTicker(1 * time.Minute)
	cleanupTicker := time.NewTicker(24 * time.Hour)

	defer queueTicker.Stop()
	defer sessionTicker.Stop()
	defer cleanupTicker.Stop()

	for {
		select {
		case <-queueTicker.C:
			service.CheckQueueReminders()
		case <-sessionTicker.C:
			service.CheckSessionTimeouts()
		case <-cleanupTicker.C:
			service.CleanOldChats()
		}
	}
}
