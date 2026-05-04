package main

import (
	"log"
	"net/http"
	"os"

	"notification-hub/internal/server/handler"
	"notification-hub/internal/server/service"
	"notification-hub/internal/server/store"
)

func main() {
	dataFile := getEnv("DATA_FILE", "notifications.json")
	port := getEnv("PORT", "8080")

	s := store.NewStore(dataFile)
	svc := service.NewNotificationService(s)
	h := handler.NewHandler(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	log.Printf("Server starting on port %s...", port)
	log.Printf("Data file: %s", dataFile)
	log.Printf("Available users: user1, user2, user3, user4, user5")

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
