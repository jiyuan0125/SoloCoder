package main

import (
	"clinical-path-backend/api"
	"clinical-path-backend/service"
	"clinical-path-backend/store"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "Server port")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8500"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	dataStore := store.NewStore()
	svc := service.NewService(dataStore)
	handler := api.NewHandler(svc)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/r", handler.EnableCORS(handler.HandlePaths))
	mux.HandleFunc("/api/r/", handler.EnableCORS(handler.HandlePathByID))
	mux.HandleFunc("/api/patients", handler.EnableCORS(handler.HandlePatients))
	mux.HandleFunc("/api/enrollments", handler.EnableCORS(handler.HandleEnrollments))
	mux.HandleFunc("/api/enrollments/", handler.EnableCORS(handler.HandleEnrollmentByID))
	mux.HandleFunc("/api/orders/", handler.EnableCORS(handler.HandleOrderByID))
	mux.HandleFunc("/api/quality", handler.EnableCORS(handler.HandleQualityMetrics))

	fmt.Printf("Clinical Path Management System starting on port %s\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
