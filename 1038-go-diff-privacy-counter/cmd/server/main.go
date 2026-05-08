package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"diff-privacy-counter/pkg/diffprivacy"
	"diff-privacy-counter/pkg/models"
)

type Server struct {
	counter *diffprivacy.DiffPrivacyCounter
}

func NewServer(counter *diffprivacy.DiffPrivacyCounter) *Server {
	return &Server{counter: counter}
}

func (s *Server) handleQueryCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CountQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Category == "" {
		http.Error(w, "category is required", http.StatusBadRequest)
		return
	}

	result, err := s.counter.QueryCount(req.Category, req.Epsilon)
	if err != nil {
		errResp := models.ErrorResponse{Error: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	resp := models.CountQueryResponse{
		Count:   result.FinalCount,
		Epsilon: result.EpsilonUsed,
	}

	if result.IsExact {
		resp.Message = "exact count (no privacy protection)"
	} else {
		resp.Message = "differentially private count"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleGetBudget(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := models.BudgetResponse{
		RemainingBudget: s.counter.RemainingBudget(),
		TotalBudget:     s.counter.TotalBudget(),
		Message:          "current privacy budget status",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleRechargeBudget(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RechargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.counter.RechargeBudget(req.Amount); err != nil {
		errResp := models.ErrorResponse{Error: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	resp := models.BudgetResponse{
		RemainingBudget: s.counter.RemainingBudget(),
		TotalBudget:     s.counter.TotalBudget(),
		Message:          fmt.Sprintf("budget recharged by %.2f", req.Amount),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func initSampleData(counter *diffprivacy.DiffPrivacyCounter) {
	sampleData := []string{
		"apple", "apple", "apple", "apple", "apple",
		"banana", "banana", "banana",
		"cherry", "cherry",
		"date",
	}
	counter.AddRecords(sampleData)
}

func main() {
	initialBudget := 10.0
	if envBudget := os.Getenv("INITIAL_BUDGET"); envBudget != "" {
		if b, err := strconv.ParseFloat(envBudget, 64); err == nil {
			initialBudget = b
		}
	}

	counter := diffprivacy.NewDiffPrivacyCounter(initialBudget)
	initSampleData(counter)

	server := NewServer(counter)

	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/count", http.HandlerFunc(server.handleQueryCount))
	mux.HandleFunc("/api/budget", http.HandlerFunc(server.handleGetBudget))
	mux.HandleFunc("/api/budget/recharge", http.HandlerFunc(server.handleRechargeBudget))

	addr := ":" + port
	log.Printf("Server starting on %s", addr)
	log.Printf("Initial budget: %.2f", initialBudget)
	log.Printf("Categories: %s", strings.Join(counter.AllCategories(), ", "))

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
