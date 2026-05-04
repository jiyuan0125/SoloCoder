package main

import (
	"delivery-tracking/handler"
	"delivery-tracking/repository"
	"delivery-tracking/service"
	"log"
	"net/http"
	"os"
)

const (
	defaultDBPath = "./delivery_tracking.db"
	defaultPort   = ":8080"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	log.Printf("Initializing database at: %s", dbPath)
	repo, err := repository.NewSQLiteDeliveryRepository(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer repo.Close()

	log.Println("Database initialized successfully")

	deliveryService := service.NewDeliveryService(repo)
	deliveryHandler := handler.NewDeliveryHandler(deliveryService)

	http.HandleFunc("/api/delivery/status", deliveryHandler.ReportStatus)
	http.HandleFunc("/api/delivery/status/batch", deliveryHandler.ReportBatchStatuses)
	http.HandleFunc("/api/delivery/trajectory", deliveryHandler.GetTrajectory)

	log.Printf("Server starting on port %s", port)
	log.Printf("Endpoints:")
	log.Printf("  POST /api/delivery/status - Report single delivery status")
	log.Printf("  POST /api/delivery/status/batch - Report multiple delivery statuses")
	log.Printf("  GET  /api/delivery/trajectory?order_no={order_no} - Get delivery trajectory")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
