package main

import (
	"log"
	"net/http"
)

func main() {
	handler := NewEmojiHandler()

	http.HandleFunc("/api/emoji", handler.HandleEmoji)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
