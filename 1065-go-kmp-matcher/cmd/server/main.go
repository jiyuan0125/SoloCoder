package main

import (
	"encoding/json"
	"net/http"

	"github.com/kmpmatcher/api"
	"github.com/kmpmatcher/kmp"
)

func singleMatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.SingleMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	opts := kmp.MatchOptions{
		CaseInsensitive: req.CaseInsensitive,
	}
	positions := kmp.SingleMatch(req.Text, req.Pattern, opts)

	resp := api.SingleMatchResponse{Positions: positions}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func multiMatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.MultiMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	opts := kmp.MatchOptions{
		CaseInsensitive: req.CaseInsensitive,
	}
	results := kmp.MultiMatch(req.Text, req.Patterns, opts)

	resp := api.MultiMatchResponse{Results: results}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/single", singleMatchHandler)
	http.HandleFunc("/multi", multiMatchHandler)
	http.ListenAndServe(":8080", nil)
}
