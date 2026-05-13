package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	store := NewMetricStore()
	handler := NewHandler(store)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9201"
	}

	log.Printf("Starting metrics collector on port %s", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
