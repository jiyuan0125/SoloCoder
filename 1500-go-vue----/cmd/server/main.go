package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"oacms/internal/core"
	"oacms/internal/server"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "listening port (default 8080)")
	flag.Parse()

	if port == "" {
		if envPort := os.Getenv("OACMS_PORT"); envPort != "" {
			port = envPort
		}
	}
	if port == "" {
		port = "8080"
	}

	store := core.NewStore()
	ownerSvc := core.NewOwnerService(store)
	electionSvc := core.NewElectionService(store, ownerSvc)
	proposalSvc := core.NewProposalService(store, ownerSvc)
	announcementSvc := core.NewAnnouncementService(store)

	srv := server.NewServer(ownerSvc, electionSvc, proposalSvc, announcementSvc)
	handler := srv.Routes()

	addr := ":" + port
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
