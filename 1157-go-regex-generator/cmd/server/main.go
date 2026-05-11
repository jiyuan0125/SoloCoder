package main

import (
	"encoding/json"
	"log"
	"net/http"

	"regex-generator/internal/regexgen"
	"regex-generator/pkg/api"
)

func generateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: "method not allowed"})
		return
	}
	
	var req api.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: "invalid request body"})
		return
	}
	
	if len(req.Examples) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: "no examples provided"})
		return
	}
	
	config := regexgen.Config{
		ShortestMatch: req.ShortestMatch,
		AllowOptional: req.AllowOptional,
	}
	
	result, err := regexgen.Generate(req.Examples, config)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: err.Error()})
		return
	}
	
	resp := api.GenerateResponse{
		Regex:       result.Regex,
		Explanation: result.Explanation,
		Examples:    result.Examples,
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: "method not allowed"})
		return
	}
	
	var req api.ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: "invalid request body"})
		return
	}
	
	if req.Regex == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: "no regex provided"})
		return
	}
	
	matched, err := regexgen.Validate(req.Regex, req.TestStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: err.Error()})
		return
	}
	
	resp := api.ValidateResponse{Matched: matched}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/generate", generateHandler)
	http.HandleFunc("/validate", validateHandler)
	
	addr := ":8601"
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
