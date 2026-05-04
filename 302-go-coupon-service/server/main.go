package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	port := flag.String("port", "8080", "HTTP server port")
	dataDir := flag.String("data", "./data", "Data directory for persistent storage")
	flag.Parse()

	absDataDir, err := filepath.Abs(*dataDir)
	if err != nil {
		log.Fatalf("Failed to get absolute data directory: %v", err)
	}

	if err := os.MkdirAll(absDataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	store, err := NewStore(absDataDir)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}

	service := NewService(store)
	handler := NewHandler(service)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	addr := ":" + *port
	log.Printf("Coupon service server starting on %s...", addr)
	log.Printf("Data directory: %s", absDataDir)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
