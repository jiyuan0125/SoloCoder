package main

import (
	"canteen/internal/core"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

const defaultPort = "8080"

func main() {
	port := flag.String("port", "", "Server port (e.g., 8080)")
	flag.Parse()

	if *port == "" {
		*port = os.Getenv("CANTEEN_PORT")
	}
	if *port == "" {
		*port = defaultPort
	}

	service := core.NewCanteenService()
	handler := NewHandler(service)

	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Starting canteen server on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
