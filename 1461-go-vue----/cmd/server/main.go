package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

const defaultPort = 8080

func getPort() int {
	port := defaultPort

	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	flagPort := flag.Int("port", 0, "server port (can also be set via PORT environment variable)")
	flag.Parse()

	if *flagPort != 0 {
		port = *flagPort
	}

	return port
}

func main() {
	port := getPort()
	s := newServer()
	router := s.router()

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Quality Trace Server starting on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
