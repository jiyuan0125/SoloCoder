package main

import (
	"log"
	"net/http"
	"os"

	"shift-scheduler/internal/database"
	"shift-scheduler/internal/handler"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
