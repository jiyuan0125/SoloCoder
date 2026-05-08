package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/example/jsonpatch/pkg/common"
	"github.com/example/jsonpatch/pkg/jsonpatch"
)

func applyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, common.ApplyResponse{
			Success: false,
			Error:   "invalid request body",
		}, http.StatusBadRequest)
		return
	}

	var doc interface{}
	if err := json.Unmarshal(req.Document, &doc); err != nil {
		sendJSON(w, common.ApplyResponse{
			Success: false,
			Error:   "invalid document JSON",
		}, http.StatusBadRequest)
		return
	}

	var patch jsonpatch.Patch
	if err := json.Unmarshal(req.Patch, &patch); err != nil {
		sendJSON(w, common.ApplyResponse{
			Success: false,
			Error:   "invalid patch JSON",
		}, http.StatusBadRequest)
		return
	}

	result, err := jsonpatch.Apply(doc, patch)
	if err != nil {
		sendJSON(w, common.ApplyResponse{
			Success: false,
			Error:   err.Error(),
		}, http.StatusBadRequest)
		return
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		sendJSON(w, common.ApplyResponse{
			Success: false,
			Error:   err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	sendJSON(w, common.ApplyResponse{
		Success:  true,
		Document: resultBytes,
	}, http.StatusOK)
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, common.ValidateResponse{
			Success: false,
			Error:   "invalid request body",
		}, http.StatusBadRequest)
		return
	}

	var patch jsonpatch.Patch
	if err := json.Unmarshal(req.Patch, &patch); err != nil {
		sendJSON(w, common.ValidateResponse{
			Success: true,
			Valid:   false,
			Error:   err.Error(),
		}, http.StatusOK)
		return
	}

	if err := jsonpatch.Validate(patch); err != nil {
		sendJSON(w, common.ValidateResponse{
			Success: true,
			Valid:   false,
			Error:   err.Error(),
		}, http.StatusOK)
		return
	}

	sendJSON(w, common.ValidateResponse{
		Success: true,
		Valid:   true,
	}, http.StatusOK)
}

func sendJSON(w http.ResponseWriter, response interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/apply", applyHandler)
	http.HandleFunc("/validate", validateHandler)

	fmt.Println("JSON Patch server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("server error: %v\n", err)
	}
}
