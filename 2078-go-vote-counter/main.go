package main

import (
	"log"
	"net/http"

	"vote-counter/db"
	"vote-counter/handlers"
)

func main() {
	if err := db.InitDB("./vote_counter.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	mux := http.NewServeMux()

	mux.HandleFunc("/polls", handlers.CreatePoll)
	mux.HandleFunc("/polls/", handlers.PollHandler)

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
