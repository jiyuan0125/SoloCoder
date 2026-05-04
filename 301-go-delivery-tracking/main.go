package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const dataFilePath = "./delivery_data.json"

func main() {
	store, err := NewDataStore(dataFilePath)
	if err != nil {
		log.Fatalf("Failed to initialize data store: %v", err)
	}

	reportHandler := NewReportHandler(store)
	queryHandler := NewQueryHandler(store)

	http.HandleFunc("/report", reportHandler.Handle)
	http.HandleFunc("/query", queryHandler.Handle)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		if err := store.Save(); err != nil {
			log.Printf("Failed to save data on shutdown: %v", err)
		}
		os.Exit(0)
	}()

	fmt.Println("Delivery tracking service starting on :8080")
	fmt.Println("Available endpoints:")
	fmt.Println("  POST /report - Report delivery status")
	fmt.Println("  GET  /query?order_id=xxx - Query delivery trail")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
