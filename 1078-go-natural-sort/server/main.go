package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"naturalsort/common"
	"naturalsort/naturalsort"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	http.HandleFunc("/sort", handleSort)
	http.HandleFunc("/health", handleHealth)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleSort(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req common.SortRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.Strings == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ErrorResponse{Error: "Missing 'strings' field"})
		return
	}

	opts := naturalsort.Options{
		CaseSensitive:    req.CaseSensitive,
		Descending:       req.Descending,
		KeepLeadingZeros: req.KeepLeadingZeros,
	}

	sorted := naturalsort.Sort(req.Strings, opts)

	resp := common.SortResponse{
		Sorted: sorted,
	}

	json.NewEncoder(w).Encode(resp)
}
