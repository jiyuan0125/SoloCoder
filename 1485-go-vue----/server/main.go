package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"coldchain/core"
)

const (
	defaultPort = "8080"
)

func main() {
	port := getPort()

	service := core.NewService()
	handlers := NewHandlers(service)
	router := NewRouter(handlers)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("ColdChain server starting on %s", addr)
	log.Printf("Press Ctrl+C to stop")

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func getPort() string {
	port := flag.String("port", "", "Server port (also can be set via PORT environment variable)")
	flag.Parse()

	if *port != "" {
		return *port
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		return envPort
	}

	return defaultPort
}
