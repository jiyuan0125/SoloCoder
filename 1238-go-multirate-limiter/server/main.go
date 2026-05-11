package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

const defaultPort = "8080"

func getPort() string {
	if envPort := os.Getenv("LIMITER_PORT"); envPort != "" {
		return envPort
	}

	port := flag.String("port", defaultPort, "server port")
	flag.Parse()
	return *port
}

func main() {
	port := getPort()

	server := NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/allow", server.HandleAllow)
	mux.HandleFunc("/config", server.HandleConfig)
	mux.HandleFunc("/stats", server.HandleStats)

	addr := ":" + port
	log.Printf("Rate limiter server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
