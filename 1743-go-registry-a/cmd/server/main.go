package main

import (
	"log"
	"net/http"
	"os"
	"registry/registry"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8426"
	}

	reg := registry.New()
	handler := registry.NewHandler(reg)

	mux := http.NewServeMux()
	mux.HandleFunc("/register", handler.Register)
	mux.HandleFunc("/heartbeat", handler.Heartbeat)
	mux.HandleFunc("/discover", handler.Discover)
	mux.HandleFunc("/watch", handler.Watch)
	mux.HandleFunc("/list", handler.List)
	mux.HandleFunc("/get", handler.Get)

	log.Printf("registry server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
