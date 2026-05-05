package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"event-collector/common"
	"event-collector/server"
)

func main() {
	port := flag.Int("port", common.DefaultServerPort, "Server port")
	flag.Parse()

	storage := server.NewMemoryStorage()
	validator := server.NewValidator()
	deduplicator := server.NewDeduplicator()
	reporter := server.NewDailyReporter(storage)
	aggregator := server.NewAggregator(storage)

	handler := server.NewHandler(storage, validator, deduplicator, reporter, aggregator)

	mux := http.NewServeMux()
	mux.HandleFunc("/track", handler.Track)
	mux.HandleFunc("/query", handler.Query)
	mux.HandleFunc("/aggregate", handler.Aggregate)
	mux.HandleFunc("/reports", handler.GetReports)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Event collector server starting on port %d...", *port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
