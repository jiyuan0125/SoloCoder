package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	
	"github.com/example/benchcompare/pkg/api"
	"github.com/example/benchcompare/pkg/benchmark"
)

func main() {
	port := flag.String("port", "", "HTTP server port (e.g., :8080)")
	flag.Parse()
	
	if *port == "" {
		*port = os.Getenv("BENCH_COMPARE_PORT")
	}
	if *port == "" {
		*port = ":8080"
	}
	
	http.HandleFunc("/compare", handleCompare)
	
	log.Printf("Server starting on %s...", *port)
	if err := http.ListenAndServe(*port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req api.CompareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	response := benchmark.Compare(req.Baseline, req.Current)
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
