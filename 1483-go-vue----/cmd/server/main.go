package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"insurance-claim/pkg/core"
)

const defaultPort = "8080"

func main() {
	var port string

	flag.StringVar(&port, "port", "", "Server port (overrides environment variable)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("CLAIM_SERVER_PORT")
	}
	if port == "" {
		port = defaultPort
	}

	store := core.NewMemoryStore()
	service := core.NewClaimService(store)

	handler := NewHandler(service)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Insurance Claim Management Server starting on %s...", addr)
	log.Printf("Server listening on port %s", port)

	err := http.ListenAndServe(addr, handler.Router())
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
