package main

import (
	"encoding/json"
	"net/http"

	"go-unit-converter/pkg/api"
	"go-unit-converter/pkg/unitconv"
)

func ConvertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed. Use POST.")
		return
	}

	var req api.ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := unitconv.Convert(req.Value, req.From, req.To, req.TempDelta)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := api.ConvertResponse{
		Value:    result,
		From:     req.From,
		To:       req.To,
		Original: req.Value,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(api.ErrorResponse{Error: message})
}
