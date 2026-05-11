package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"disruptor/server"
)

const (
	defaultPort = "8080"
)

func getPort() string {
	port := flag.String("port", "", "server port")
	flag.Parse()

	if *port != "" {
		return *port
	}

	if envPort := os.Getenv("DISRUPTOR_PORT"); envPort != "" {
		return envPort
	}

	return defaultPort
}

func main() {
	port := getPort()
	srv := server.NewServer()
	addr := ":" + port

	log.Printf("disruptor server listening on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
