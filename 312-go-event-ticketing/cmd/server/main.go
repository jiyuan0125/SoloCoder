package main

import (
	"flag"
	"log"
	"net/http"

	"event-ticketing/internal/server"
)

func main() {
	port := flag.String("port", "8080", "HTTP server port")
	dataDir := flag.String("data", "./data", "Data directory for persistence")
	flag.Parse()

	store, err := server.NewStore(*dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}

	handler := server.NewHandler(store)

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.CreateEvent(w, r)
		} else if r.Method == http.MethodGet {
			handler.ListEvents(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/tickets/purchase", handler.PurchaseTicket)
	http.HandleFunc("/tickets/checkin", handler.CheckIn)
	http.HandleFunc("/tickets/refund", handler.RefundTicket)
	http.HandleFunc("/events/stats/", handler.GetEventStats)

	log.Printf("Server starting on port %s...", *port)
	log.Printf("Data directory: %s", *dataDir)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
