package main

import (
	"log"
	"net/http"
	"strings"

	"inventory-alert/database"
	"inventory-alert/handlers"
)

func main() {
	err := database.InitDB("inventory.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		handlers.ProductsHandler(w, r)
	})

	mux.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/products/")
		parts := strings.Split(path, "/")

		if len(parts) == 1 || (len(parts) == 2 && parts[1] == "") {
			handlers.ProductHandler(w, r)
		} else if len(parts) >= 2 && parts[1] == "alert" {
			handlers.ProductAlertHandler(w, r)
		} else if len(parts) >= 2 && parts[1] == "purchase" {
			handlers.ProductPurchaseHandler(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/alerts", func(w http.ResponseWriter, r *http.Request) {
		handlers.AlertsHandler(w, r)
	})

	mux.HandleFunc("/alerts/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/alerts/")
		parts := strings.Split(path, "/")

		if len(parts) == 1 || (len(parts) == 2 && parts[1] == "") {
			handlers.AlertHandler(w, r)
		} else if len(parts) >= 2 && parts[1] == "assign" {
			handlers.AlertAssignHandler(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/statistics/categories", handlers.StatisticsHandler)

	log.Println("Inventory Alert Service starting on port 8104...")
	log.Fatal(http.ListenAndServe(":8104", mux))
}
