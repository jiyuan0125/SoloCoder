package main

import (
	"encoding/json"
	"net/http"

	"phone-formatter/pkg/api"
	"phone-formatter/pkg/phone"
)

func FormatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.FormatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var formatType phone.FormatType
	switch req.FormatType {
	case "domestic":
		formatType = phone.FormatDomestic
	case "international":
		formatType = phone.FormatInternational
	case "pure":
		formatType = phone.FormatPureNumber
	default:
		formatType = phone.FormatDomestic
	}

	result, err := phone.Format(req.Phone, formatType)
	if err != nil {
		json.NewEncoder(w).Encode(api.FormatResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(api.FormatResponse{
		Success: true,
		Result:  result,
	})
}

func ValidateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	valid := phone.Validate(req.Phone)

	json.NewEncoder(w).Encode(api.ValidateResponse{
		Success: true,
		Valid:   valid,
	})
}

func ExtractHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.ExtractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	results := phone.ExtractFromText(req.Text)

	json.NewEncoder(w).Encode(api.ExtractResponse{
		Success: true,
		Results: results,
	})
}
