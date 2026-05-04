package main

import (
	"flag"
	"log"
	"net/http"

	"feedback-system/internal/server/handler"
	"feedback-system/internal/server/store"
)

func main() {
	port := flag.String("port", "8080", "HTTP server port")
	dataPath := flag.String("data", "./data", "Data storage path")
	flag.Parse()

	storage, err := store.NewFileStore(*dataPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	h := handler.NewHandler(storage)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/feedback", h.SubmitFeedback)
	mux.HandleFunc("GET /api/feedback", h.ListFeedbacks)
	mux.HandleFunc("GET /api/feedback/{id}", h.GetFeedback)
	mux.HandleFunc("POST /api/feedback/{id}/note", h.AddNote)
	mux.HandleFunc("PUT /api/feedback/{id}/status", h.UpdateStatus)
	mux.HandleFunc("GET /api/statistics", h.GetStatistics)

	log.Printf("Server starting on port %s...", *port)
	log.Printf("Data will be stored in: %s", *dataPath)

	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
