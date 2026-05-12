package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	ruleStore := NewRuleStore()
	statsStore := NewStatsStore()
	handler := NewHandler(ruleStore, statsStore)

	mux := http.NewServeMux()
	mux.HandleFunc("/rules", handler.HandleRules)
	mux.HandleFunc("/stats", handler.HandleStats)
	mux.HandleFunc("/check", handler.HandleCheck)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("Throttle service listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
