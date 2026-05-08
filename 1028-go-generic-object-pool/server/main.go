package main

import (
	"log"
	"net/http"
)

func main() {
	manager := NewPoolManager()

	http.HandleFunc("/pool/create", createPoolHandler(manager))
	http.HandleFunc("/pool/get", getObjectHandler(manager))
	http.HandleFunc("/pool/put", putObjectHandler(manager))
	http.HandleFunc("/pool/stats", statsHandler(manager))
	http.HandleFunc("/pool/close", closePoolHandler(manager))

	log.Println("server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
