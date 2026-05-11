package main

import (
	"flag"
	"log"
	"os"

	"green-care-management/pkg/core"
)

func main() {
	var port string

	flag.StringVar(&port, "port", "", "Server port")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	store := core.NewInMemoryStore()
	service := core.NewService(store)

	server := NewServer(service)

	log.Printf("Server starting on port %s", port)
	if err := server.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
