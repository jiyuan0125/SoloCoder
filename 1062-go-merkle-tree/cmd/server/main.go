package main

import (
	"log"
	"net/http"
)

func main() {
	store := NewTreeStore()
	handler := NewHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/build-tree", handler.BuildTree)
	mux.HandleFunc("/api/verify-proof", handler.VerifyProof)
	mux.HandleFunc("/api/find-differences", handler.FindDifferences)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Server starting on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
