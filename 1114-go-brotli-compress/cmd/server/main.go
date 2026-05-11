package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/solocoder/brotli/pkg/api"
	"github.com/solocoder/brotli/pkg/brotli"
)

func main() {
	http.HandleFunc("/compress", handleCompress)
	http.HandleFunc("/decompress", handleDecompress)
	http.HandleFunc("/health", handleHealth)

	port := ":8104"
	fmt.Printf("Brotli compression server listening on %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleCompress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var req api.CompressRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if req.Quality == 0 {
		req.Quality = api.DefaultQuality
	}
	req.Quality = api.ValidateQuality(req.Quality)

	originalSize := len(req.Data)
	compressed, err := brotli.Encode(req.Data, req.Quality)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Compression failed: "+err.Error())
		return
	}

	resp := api.CompressResponse{
		Data:           compressed,
		OriginalSize:   originalSize,
		CompressedSize: len(compressed),
		Quality:        req.Quality,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleDecompress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var req api.DecompressRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	compressedSize := len(req.Data)
	decompressed, err := brotli.Decode(req.Data)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Decompression failed: "+err.Error())
		return
	}

	resp := api.DecompressResponse{
		Data:           decompressed,
		OriginalSize:   len(decompressed),
		CompressedSize: compressedSize,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(api.ErrorResponse{Error: message})
}
