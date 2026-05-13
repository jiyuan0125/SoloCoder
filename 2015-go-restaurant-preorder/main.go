package main

import (
	"fmt"
	"net/http"
	"restaurant-preorder/database"
	"restaurant-preorder/handler"
	"restaurant-preorder/service"
	"strings"
	"time"
)

func main() {
	err := database.InitDB("./restaurant.db")
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}

	go startBackgroundJobs()

	http.HandleFunc("/api/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetAllOrdersHandler(w, r)
		} else if r.Method == http.MethodPost {
			handler.CreateOrderHandler(w, r)
		} else {
			handler.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	http.HandleFunc("/api/orders/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/cancel") {
			handler.CancelOrderHandler(w, r)
		} else if strings.Contains(path, "/status/") {
			handler.UpdateOrderStatusHandler(w, r)
		} else {
			handler.GetOrderHandler(w, r)
		}
	})

	http.HandleFunc("/api/kitchen-summary", handler.GetKitchenSummaryHandler)
	http.HandleFunc("/api/resource-summary", handler.GetResourceSummaryHandler)

	fmt.Println("Server starting on port 8302...")
	if err := http.ListenAndServe(":8302", nil); err != nil {
		panic(fmt.Sprintf("Server failed: %v", err))
	}
}

func startBackgroundJobs() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := service.CheckNoShows()
			if err != nil {
				fmt.Printf("Error checking no-shows: %v\n", err)
			}
		}
	}
}
