package main

import (
	"encoding/json"
	"net/http"

	"csv-rfc4180/common"
	"csv-rfc4180/csvparser"
)

func handleParse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(common.ParseResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	opt := csvparser.ParserOption{WithHeader: req.WithHeader}
	result, err := csvparser.Parse(req.CSV, opt)
	if err != nil {
		json.NewEncoder(w).Encode(common.ParseResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	resp := common.ParseResponse{
		Success: true,
		Records: result.Records,
		Headers: result.Headers,
		Rows:    result.Rows,
	}
	json.NewEncoder(w).Encode(resp)
}

func handleSerialize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SerializeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(common.SerializeResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	csv := csvparser.Serialize(req.Data)

	resp := common.SerializeResponse{
		Success: true,
		CSV:     csv,
	}
	json.NewEncoder(w).Encode(resp)
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(common.ValidateResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	opt := csvparser.ParserOption{WithHeader: req.WithHeader}
	errs := csvparser.Validate(req.CSV, opt)

	var validateErrs []common.ValidateError
	for _, e := range errs {
		validateErrs = append(validateErrs, common.ValidateError{
			Line:    e.Line,
			Column:  e.Column,
			Message: e.Message,
		})
	}

	resp := common.ValidateResponse{
		Success: true,
		Valid:   len(validateErrs) == 0,
		Errors:  validateErrs,
	}
	json.NewEncoder(w).Encode(resp)
}
