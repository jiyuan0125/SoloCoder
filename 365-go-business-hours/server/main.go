package main

import (
	"log"
	"net/http"
)

func main() {
	bh := initDefaultBusinessHours()
	handler := NewHandler(bh)
	router := NewRouter(handler)

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
