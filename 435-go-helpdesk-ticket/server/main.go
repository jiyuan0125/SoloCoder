package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	store := NewStore()
	service := NewTicketService(store)
	handler := NewHandler(service)

	go startSLAWatcher(service)

	http.HandleFunc("/tickets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.CreateTicket(w, r)
		} else {
			handler.ListTickets(w, r)
		}
	})

	http.HandleFunc("/tickets/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case path == "/tickets/statistics":
			handler.GetStatistics(w, r)
		case path == "/tickets/handlers":
			handler.GetHandlers(w, r)
		case path == "/tickets/templates":
			handler.GetTemplates(w, r)
		case path == "/tickets/autoclassify":
			handler.AutoClassify(w, r)
		case path == "/tickets/link":
			handler.LinkTickets(w, r)
		case path == "/tickets/":
			handler.ListTickets(w, r)
		default:
			if endsWith(path, "/reassign") {
				handler.ReassignTicket(w, r)
			} else if endsWith(path, "/status") {
				handler.UpdateStatus(w, r)
			} else if endsWith(path, "/rate") {
				handler.RateTicket(w, r)
			} else if endsWith(path, "/transfer") {
				handler.TransferTicket(w, r)
			} else if endsWith(path, "/logs") {
				handler.GetLogs(w, r)
			} else {
				handler.GetTicket(w, r)
			}
		}
	})

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func startSLAWatcher(service *TicketService) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		service.CheckSLAAndEscalate()
	}
}

func endsWith(path, suffix string) bool {
	if len(path) < len(suffix) {
		return false
	}
	return path[len(path)-len(suffix):] == suffix
}
