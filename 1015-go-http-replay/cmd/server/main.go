package main

import (
	"flag"
	"log"
	"net/http"

	"http-replay/server"
)

func main() {
	addr := flag.String("addr", ":8080", "Server address")
	sessionsDir := flag.String("sessions", "./sessions", "Directory to store session files")
	flag.Parse()

	handler := server.NewHandler(*sessionsDir)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	log.Printf("HTTP Replay Server starting on %s", *addr)
	log.Printf("Sessions directory: %s", *sessionsDir)

	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
