package main

import (
	"log"
	"net/http"
)

const defaultAddress = ":8080"

func main() {
	handler := NewRetryHandler()
	
	mux := http.NewServeMux()
	mux.HandleFunc("POST /execute", handler.Execute)
	mux.HandleFunc("GET /health", handler.Health)
	
	log.Printf("Server starting on %s...", defaultAddress)
	if err := http.ListenAndServe(defaultAddress, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
