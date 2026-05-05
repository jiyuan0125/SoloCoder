package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	defaultPort = 8080
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", HealthHandler)
	mux.HandleFunc("/parse", ParseHandler)

	port := defaultPort
	addr := fmt.Sprintf(":%d", port)

	log.Printf("Server starting on port %d...", port)
	log.Printf("Endpoints:")
	log.Printf("  GET  /health - Health check")
	log.Printf("  POST /parse  - Parse card number")
	log.Printf("Server listening on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
