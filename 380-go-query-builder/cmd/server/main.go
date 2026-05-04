package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	handler := NewHandler()

	http.HandleFunc("/api/query", handler.BuildQuery)

	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
