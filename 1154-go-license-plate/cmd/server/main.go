package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"licenseplate/pkg/api"
	"licenseplate/pkg/license"
)

func validateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req api.ValidateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Plate == "" {
		http.Error(w, "Plate is required", http.StatusBadRequest)
		return
	}

	result := license.Validate(req.Plate)

	resp := api.ValidateResponse{
		Plate:  req.Plate,
		Valid:  result.Valid,
		Type:   string(result.Type),
		Reason: result.Reason,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func batchValidateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req api.BatchValidateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.Plates) == 0 {
		http.Error(w, "Plates is required", http.StatusBadRequest)
		return
	}

	results := make([]api.ValidateResponse, 0, len(req.Plates))
	for _, plate := range req.Plates {
		result := license.Validate(plate)
		results = append(results, api.ValidateResponse{
			Plate:  plate,
			Valid:  result.Valid,
			Type:   string(result.Type),
			Reason: result.Reason,
		})
	}

	resp := api.BatchValidateResponse{
		Results: results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	http.HandleFunc("/validate", validateHandler)
	http.HandleFunc("/batch", batchValidateHandler)
	http.HandleFunc("/health", healthHandler)

	port := ":8503"
	fmt.Printf("License plate validation server starting on %s\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /validate - Validate a single plate")
	fmt.Println("  POST /batch    - Validate multiple plates")
	fmt.Println("  GET  /health   - Health check")

	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
