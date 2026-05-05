package main

import (
	"go-faq-service/internal/server"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	storage := server.NewStorage()
	svc := server.NewService(storage)
	handler := server.NewHandler(svc)

	r := mux.NewRouter()

	r.Use(handler.CORS)
	r.Use(handler.Logging)

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/health", handler.HealthCheck).Methods("GET", "OPTIONS")

	api.HandleFunc("/faqs", handler.ListAllFAQs).Methods("GET", "OPTIONS")
	api.HandleFunc("/faqs", handler.CreateFAQ).Methods("POST", "OPTIONS")
	api.HandleFunc("/faqs/search", handler.SearchFAQ).Methods("GET", "OPTIONS")
	api.HandleFunc("/faqs/batch-import", handler.BatchImportFAQ).Methods("POST", "OPTIONS")
	api.HandleFunc("/faqs/{id}", handler.GetFAQ).Methods("GET", "OPTIONS")
	api.HandleFunc("/faqs/{id}", handler.UpdateFAQ).Methods("PUT", "OPTIONS")
	api.HandleFunc("/faqs/{id}", handler.DeleteFAQ).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/faqs/{id}/enable", handler.SetFAQEnabled).Methods("POST", "OPTIONS")
	api.HandleFunc("/faqs/{id}/pin", handler.SetFAQPinned).Methods("POST", "OPTIONS")
	api.HandleFunc("/faqs/{id}/click", handler.RecordClick).Methods("POST", "OPTIONS")
	api.HandleFunc("/faqs/{id}/history", handler.GetFAQHistory).Methods("GET", "OPTIONS")

	api.HandleFunc("/categories", handler.ListAllCategories).Methods("GET", "OPTIONS")
	api.HandleFunc("/categories", handler.CreateCategory).Methods("POST", "OPTIONS")
	api.HandleFunc("/categories/tree", handler.GetCategoryTree).Methods("GET", "OPTIONS")
	api.HandleFunc("/categories/{id}", handler.UpdateCategory).Methods("PUT", "OPTIONS")
	api.HandleFunc("/categories/{id}", handler.DeleteCategory).Methods("DELETE", "OPTIONS")

	api.HandleFunc("/statistics", handler.GetStatistics).Methods("GET", "OPTIONS")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("FAQ Service starting on port %s...", port)
	log.Printf("API endpoints available at http://localhost:%s/api/v1", port)
	log.Printf("Health check: http://localhost:%s/api/v1/health", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
