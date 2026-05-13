package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"order-matching/internal/engine"
	"order-matching/internal/handler"
	"order-matching/internal/model"
)

func main() {
	dbPath := "./orderbook.db"
	
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	db, err := model.NewDatabase(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	matchingEngine := engine.NewMatchingEngine(db)
	handlers := handler.NewHandler(matchingEngine)

	http.HandleFunc("/api/order/submit", handlers.SubmitOrder)
	http.HandleFunc("/api/order/cancel", handlers.CancelOrder)
	http.HandleFunc("/api/orderbook", handlers.GetOrderBook)

	port := ":8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = ":" + envPort
	}
	
	fmt.Printf("Server starting on %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
