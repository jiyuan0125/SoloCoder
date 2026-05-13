package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"medical-quality-system/internal/app"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8501, "Server port")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	server := app.NewServer()
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server starting on port %d...", port)
	if err := server.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
