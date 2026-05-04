package main

import (
	"log"
	"net/http"
)

func main() {
	router := NewRouter()
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
