package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/matrix-ops/matrix-service/api"
)

func main() {
	http.HandleFunc("/multiply", handleMultiply)
	http.HandleFunc("/transpose", handleTranspose)
	http.HandleFunc("/determinant", handleDeterminant)

	log.Println("Matrix service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleMultiply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.MultiplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	response, err := api.HandleMultiply(req)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func handleTranspose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.UnaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	response, err := api.HandleTranspose(req)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func handleDeterminant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.UnaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	response, err := api.HandleDeterminant(req)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func writeError(w http.ResponseWriter, message string, status int) {
	errResp := api.MatrixResponse{
		Success: false,
		Error:   strings.TrimSpace(message),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errResp)
}
