package main

import (
	"encoding/json"
	"fmt"
	"iban/internal/api"
	"iban/internal/iban"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/validate", validateHandler)
	http.HandleFunc("/format", formatHandler)
	http.HandleFunc("/batch-validate", batchValidateHandler)
	http.HandleFunc("/health", healthHandler)

	port := ":8604"
	fmt.Printf("IBAN Validation Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.IBAN == "" {
		writeJSONError(w, "IBAN is required", http.StatusBadRequest)
		return
	}

	result := iban.Validate(req.IBAN)
	formatted := iban.Format(req.IBAN)

	response := api.ValidateResponse{
		Valid:         result.Valid,
		CountryCode:   result.CountryCode,
		CountryName:   result.CountryName,
		BBAN:          result.BBAN,
		FormatValid:   result.FormatValid,
		ChecksumValid: result.ChecksumValid,
		Formatted:     formatted,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func formatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.FormatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.IBAN == "" {
		writeJSONError(w, "IBAN is required", http.StatusBadRequest)
		return
	}

	formatted := iban.Format(req.IBAN)

	response := api.FormatResponse{
		Formatted: formatted,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func batchValidateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.BatchValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.IBANs) == 0 {
		writeJSONError(w, "IBAN list is required", http.StatusBadRequest)
		return
	}

	results := make([]api.ValidateResponse, 0, len(req.IBANs))
	for _, ibanStr := range req.IBANs {
		result := iban.Validate(ibanStr)
		formatted := iban.Format(ibanStr)
		results = append(results, api.ValidateResponse{
			Valid:         result.Valid,
			CountryCode:   result.CountryCode,
			CountryName:   result.CountryName,
			BBAN:          result.BBAN,
			FormatValid:   result.FormatValid,
			ChecksumValid: result.ChecksumValid,
			Formatted:     formatted,
		})
	}

	response := api.BatchValidateResponse{
		Results: results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(api.ErrorResponse{Error: message})
}
