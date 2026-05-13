package main

import (
	"log"
	"net/http"

	"aftersale-ticket/db"
	"aftersale-ticket/handlers"
)

func main() {
	store, err := db.NewStore("./aftersale.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()

	handler := handlers.NewHandler(store)

	mux := http.NewServeMux()
	mux.Handle("/", handler)

	log.Println("Server starting on :8401...")
	log.Println("Available endpoints:")
	log.Println("  POST /tickets          - Create a new ticket")
	log.Println("  GET  /tickets          - List all tickets")
	log.Println("  GET  /tickets/{id}     - Get a specific ticket")
	log.Println("  POST /tickets/{id}/claim   - Claim a ticket")
	log.Println("  POST /tickets/{id}/review  - Review a ticket")
	log.Println("  GET  /tickets/{id}/logistics - Get logistics info")
	log.Println("  POST /tickets/{id}/qc   - Perform QC (return tickets)")
	log.Println("  POST /tickets/{id}/repair - Add repair record")
	log.Println("  GET  /tickets/{id}/repair - List repair records")
	log.Println("  POST /tickets/{id}/status - Update ticket status")

	if err := http.ListenAndServe(":8401", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
