package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"snappy-compress/internal/api"
	"snappy-compress/internal/snappy"
)

func main() {
	http.HandleFunc("/compress", handleCompress)
	http.HandleFunc("/decompress", handleDecompress)
	http.ListenAndServe(":8301", nil)
}

func handleCompress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.CompressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		writeError(w, "invalid base64", http.StatusBadRequest)
		return
	}

	compressed := snappy.Encode(data)
	resp := api.CompressResponse{
		Data: base64.StdEncoding.EncodeToString(compressed),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleDecompress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.DecompressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		writeError(w, "invalid base64", http.StatusBadRequest)
		return
	}

	decompressed, err := snappy.Decode(data)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := api.DecompressResponse{
		Data: base64.StdEncoding.EncodeToString(decompressed),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(api.ErrorResponse{Error: msg})
}
