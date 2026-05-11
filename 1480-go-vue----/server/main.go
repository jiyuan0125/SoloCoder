package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"vehicle-inspection/core"
)

func main() {
	var port string

	flag.StringVar(&port, "port", "", "Server port (e.g. 8080)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("SERVER_PORT")
		if port == "" {
			port = "8909"
		}
	}

	service := core.NewInspectionService()
	router := NewRouter(service)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on %s...", addr)
	
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
