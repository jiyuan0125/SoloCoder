package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"matheval/common"
	"matheval/eval"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) evaluate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var req common.EvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.EvaluateResponse{
			Success: false,
			Errors:  []string{"invalid request body: " + err.Error()},
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	expression := strings.TrimSpace(req.Expression)
	if expression == "" {
		resp := common.EvaluateResponse{
			Success: false,
			Errors:  []string{"expression cannot be empty"},
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	vars := eval.Variables{}
	if req.Variables != nil {
		for k, v := range req.Variables {
			vars[k] = v
		}
	}
	result := eval.ParseAndEvaluateWithVariableCheck(expression, vars)
	resp := common.EvaluateResponse{
		Success: result.Success,
		Value:   result.Value,
		Type:    result.Type,
		Errors:  result.Errors,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	handler := NewHandler()
	http.HandleFunc("/evaluate", handler.evaluate)
	http.HandleFunc("/health", handler.health)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8302"
	}
	addr := ":" + port
	fmt.Printf("Math Expression Evaluator Server running on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
