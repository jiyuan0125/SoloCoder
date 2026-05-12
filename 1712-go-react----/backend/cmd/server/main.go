package main

import (
	"fmt"
	"log"
	"net/http"

	"health-archive/internal/config"
	"health-archive/internal/handlers"
	"health-archive/internal/store"
)

func main() {
	cfg := config.Load()
	s := store.New()
	h := handlers.New(s)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Server starting on port %d", cfg.Port)

	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
