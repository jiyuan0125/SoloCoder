package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"return-exchange/internal/server"
	"syscall"

	"github.com/gorilla/mux"
)

func main() {
	store := server.NewStore()
	service := server.NewService(store)
	handler := server.NewHandler(service)

	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Printf("Warning: failed to create data directory: %v", err)
	}

	if err := server.LoadData(store, dataDir); err != nil {
		log.Printf("Warning: failed to load data: %v", err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/api/orders/mock", handler.CreateMockOrder).Methods("POST")

	r.HandleFunc("/api/applications/return", handler.SubmitReturn).Methods("POST")
	r.HandleFunc("/api/applications/exchange", handler.SubmitExchange).Methods("POST")
	r.HandleFunc("/api/applications/review", handler.ReviewApplication).Methods("POST")
	r.HandleFunc("/api/applications/refund", handler.ProcessRefund).Methods("POST")
	r.HandleFunc("/api/applications/shipping", handler.CreateShippingOrder).Methods("POST")
	r.HandleFunc("/api/applications/{id}/complete", handler.CompleteApplication).Methods("POST")
	r.HandleFunc("/api/applications/{id}", handler.GetApplication).Methods("GET")
	r.HandleFunc("/api/applications", handler.ListAllApplications).Methods("GET")
	r.HandleFunc("/api/users/{user_id}/applications", handler.ListUserApplications).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %s...", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")

	if err := server.SaveData(store, dataDir); err != nil {
		log.Printf("Warning: failed to save data: %v", err)
	} else {
		log.Println("Data saved successfully")
	}

	log.Println("Server stopped")
}
