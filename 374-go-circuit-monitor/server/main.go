package main

import (
	"flag"
	"log"
	"net/http"

	"circuit-monitor/circuitbreaker"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP server address")
	flag.Parse()

	registry := circuitbreaker.NewRegistry()

	router := NewRouter(registry)

	log.Printf("Starting circuit breaker server on %s", *addr)
	if err := http.ListenAndServe(*addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
