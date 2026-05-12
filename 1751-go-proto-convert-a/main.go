package main

import (
	"log"
	"net/http"
	"os"
	"protocol-converter/config"
	"protocol-converter/handler"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "rules.json"
	}

	if err := config.LoadRules(configPath); err != nil {
		log.Fatalf("Failed to load rules: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.ConvertHandler)

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
