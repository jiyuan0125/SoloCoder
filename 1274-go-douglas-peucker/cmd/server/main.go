package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/solo-coder/douglas-peucker/internal/api"
	"github.com/solo-coder/douglas-peucker/pkg/peucker"
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8501"
	}

	flagPort := flag.String("port", "", "Server port")
	flag.Parse()

	if *flagPort != "" {
		port = *flagPort
	}

	return port
}

func simplifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.SimplifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if len(req.Points) == 0 {
		http.Error(w, "No points provided", http.StatusBadRequest)
		return
	}

	resp := peucker.Simplify(&req)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	port := getPort()

	http.HandleFunc("/simplify", simplifyHandler)
	http.HandleFunc("/health", healthHandler)

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Server starting on port %s...\n", port)

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
