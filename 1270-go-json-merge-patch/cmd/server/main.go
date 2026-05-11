package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"jsonmerge/pkg/api"
	"jsonmerge/pkg/merge"
)

func main() {
	port := getPort()
	fmt.Printf("Starting server on :%s\n", port)

	http.HandleFunc("/merge/apply", handleApply)
	http.HandleFunc("/merge/diff", handleDiff)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func getPort() string {
	port := "8080"

	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	flagPort := flag.String("port", "", "Server port (overrides PORT env var)")
	flag.Parse()
	if *flagPort != "" {
		port = *flagPort
	}

	return port
}

func handleApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.ApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: err.Error()})
		return
	}

	result, err := merge.Apply(req.Original, req.Patch)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.ApplyResponse{Result: result})
}

func handleDiff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.DiffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: err.Error()})
		return
	}

	patch, err := merge.Diff(req.Source, req.Target)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.DiffResponse{Patch: patch})
}
