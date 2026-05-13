package main

import (
	"log"
	"net/http"
	"order-state-machine/database"
	"order-state-machine/handlers"
)

func main() {
	db, err := database.InitDB("./orders.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	orderHandler := handlers.NewOrderHandler(db)
	aftersalesHandler := handlers.NewAfterSalesHandler(db)
	resourceHandler := handlers.NewResourceHandler(db)

	mux := http.NewServeMux()

	mux.HandleFunc("/orders", orderHandler.HandleOrders)
	mux.HandleFunc("/orders/", orderHandler.HandleOrderByID)
	mux.HandleFunc("/orders/history/", orderHandler.HandleOrderHistory)
	mux.HandleFunc("/orders/cancel/", orderHandler.HandleCancelOrder)
	mux.HandleFunc("/orders/status", orderHandler.HandleOrderStatusFlow)

	mux.HandleFunc("/aftersales", aftersalesHandler.HandleAfterSales)
	mux.HandleFunc("/aftersales/", aftersalesHandler.HandleAfterSalesByID)
	mux.HandleFunc("/aftersales/status", aftersalesHandler.HandleAfterSalesStatusFlow)

	mux.HandleFunc("/resources", resourceHandler.HandleResources)
	mux.HandleFunc("/resources/", resourceHandler.HandleResourceByID)
	mux.HandleFunc("/resources/summary/", resourceHandler.HandleResourceSummary)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
