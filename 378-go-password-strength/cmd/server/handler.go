package main

import (
	"encoding/json"
	"net/http"
	"password-strength/pkg/api"
	"password-strength/pkg/password"
)

type EvaluateHandler struct {
	evaluator *password.Evaluator
}

func NewEvaluateHandler() *EvaluateHandler {
	return &EvaluateHandler{
		evaluator: password.NewEvaluator(),
	}
}

func (h *EvaluateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.EvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.evaluator.Evaluate(req.Password)
	if err != nil {
		resp := api.EvaluateResponse{
			Success:     false,
			Level:       "",
			Score:       0,
			Suggestions: []string{},
			Error:       err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := api.EvaluateResponse{
		Success:     true,
		Level:       string(result.Level),
		Score:       result.Score,
		Suggestions: result.Suggestions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
