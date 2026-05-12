package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	manager := NewComponentManager()
	checker := NewChecker(manager)
	handler := NewHandler(manager, checker)
	
	mux := http.NewServeMux()
	
	mux.HandleFunc("/health", handler.HealthCheck)
	mux.HandleFunc("/components", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/components" {
			if r.Method == http.MethodPost {
				handler.RegisterComponent(w, r)
			} else if r.Method == http.MethodGet {
				handler.GetComponents(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			handler.UnregisterComponent(w, r)
		}
	})
	
	addr := ":" + port
	log.Printf("health checker server starting on %s", addr)
	
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
