package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"isbn-validator/pkg/api"
	"isbn-validator/pkg/isbn"
)

const port = "8501"

func main() {
	http.HandleFunc("/api/check", handleCheck)
	http.HandleFunc("/api/convert", handleConvert)
	http.HandleFunc("/api/format", handleFormat)

	log.Printf("Server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func handleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithJSON(w, http.StatusMethodNotAllowed, api.CheckResponse{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	var req api.CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, api.CheckResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if strings.TrimSpace(req.ISBN) == "" {
		respondWithJSON(w, http.StatusBadRequest, api.CheckResponse{
			Success: false,
			Message: "ISBN is required",
		})
		return
	}

	result := isbn.Validate(req.ISBN)

	respondWithJSON(w, http.StatusOK, api.CheckResponse{
		Success:      true,
		IsValid:      result.IsValid,
		ISBNType:     string(result.ISBNType),
		CheckDigit:   result.CheckDigit,
		CalculatedCD: result.CalculatedCD,
		Message:      result.Reason,
	})
}

func handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithJSON(w, http.StatusMethodNotAllowed, api.ConvertResponse{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	var req api.ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, api.ConvertResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if strings.TrimSpace(req.ISBN) == "" {
		respondWithJSON(w, http.StatusBadRequest, api.ConvertResponse{
			Success: false,
			Message: "ISBN is required",
		})
		return
	}

	validation := isbn.Validate(req.ISBN)
	if !validation.IsValid {
		respondWithJSON(w, http.StatusBadRequest, api.ConvertResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid ISBN: %s", validation.Reason),
		})
		return
	}

	var converted string
	var convertedType string
	var checkDigit string
	var err error

	cleaned := isbn.CleanInput(req.ISBN)

	if validation.ISBNType == isbn.ISBN10 {
		converted, err = isbn.Convert10To13(req.ISBN)
		if err != nil {
			respondWithJSON(w, http.StatusBadRequest, api.ConvertResponse{
				Success: false,
				Message: fmt.Sprintf("Conversion failed: %v", err),
			})
			return
		}
		convertedType = "ISBN-13"
		checkDigit = string(converted[12])
	} else {
		if !strings.HasPrefix(cleaned, "978") {
			respondWithJSON(w, http.StatusBadRequest, api.ConvertResponse{
				Success: false,
				Message: "ISBN-13 must start with 978 to convert to ISBN-10",
			})
			return
		}
		converted, err = isbn.Convert13To10(req.ISBN)
		if err != nil {
			respondWithJSON(w, http.StatusBadRequest, api.ConvertResponse{
				Success: false,
				Message: fmt.Sprintf("Conversion failed: %v", err),
			})
			return
		}
		convertedType = "ISBN-10"
		checkDigit = string(converted[9])
	}

	respondWithJSON(w, http.StatusOK, api.ConvertResponse{
		Success:       true,
		Original:      cleaned,
		OriginalType:  string(validation.ISBNType),
		Converted:     converted,
		ConvertedType: convertedType,
		CheckDigit:    checkDigit,
		Message:       "Conversion successful",
	})
}

func handleFormat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithJSON(w, http.StatusMethodNotAllowed, api.FormatResponse{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	var req api.FormatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, api.FormatResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if strings.TrimSpace(req.ISBN) == "" {
		respondWithJSON(w, http.StatusBadRequest, api.FormatResponse{
			Success: false,
			Message: "ISBN is required",
		})
		return
	}

	validation := isbn.Validate(req.ISBN)
	if !validation.IsValid {
		respondWithJSON(w, http.StatusBadRequest, api.FormatResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid ISBN: %s", validation.Reason),
		})
		return
	}

	cleaned := isbn.CleanInput(req.ISBN)
	formatted, err := isbn.FormatWithHyphens(req.ISBN)
	if err != nil {
		respondWithJSON(w, http.StatusOK, api.FormatResponse{
			Success:  true,
			Original: cleaned,
			Formatted: cleaned,
			Message:  fmt.Sprintf("Basic format (no hyphens): %v", err),
		})
		return
	}

	respondWithJSON(w, http.StatusOK, api.FormatResponse{
		Success:   true,
		Original:  cleaned,
		Formatted: formatted,
		Message:   "Format successful",
	})
}
