package main

import (
	"log"
	"net/http"
	"warranty-claim/internal/server"
)

const (
	port     = ":8080"
	dataFile = "warranty_data.json"
)

func main() {
	store := server.NewStore(dataFile)
	service := server.NewService(store)
	handler := server.NewHandler(service)

	mux := http.NewServeMux()
	handler.SetupRoutes(mux)

	log.Printf("Warranty Claim Server starting on port %s...", port)
	log.Printf("Data will be saved to: %s", dataFile)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
