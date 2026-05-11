package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"moving-platform/pkg/core"
)

const defaultPort = 9007

func main() {
	port := getPort()

	manager := core.NewBookingManager()
	server := NewServer(manager, port)

	log.Printf("Server starting on port %d...", port)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getPort() int {
	port := defaultPort

	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			port = p
		}
	}

	flagPort := flag.Int("port", port, "Server port")
	flag.Parse()

	if *flagPort > 0 {
		port = *flagPort
	}

	return port
}
