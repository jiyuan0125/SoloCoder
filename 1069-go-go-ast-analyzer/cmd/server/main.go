package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/example/go-ast-analyzer/internal/api"
	"github.com/example/go-ast-analyzer/pkg/analyzer"
)

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req api.AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	response := analyzer.Analyze(req.Filename, req.Source)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/analyze", analyzeHandler)
	
	fmt.Println("Server starting on :8303...")
	if err := http.ListenAndServe(":8303", nil); err != nil {
		log.Fatal(err)
	}
}
