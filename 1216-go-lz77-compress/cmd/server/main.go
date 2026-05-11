package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"lz77-compressor/pkg/api"
	"lz77-compressor/pkg/lz77"
)

func main() {
	port := flag.String("port", "8080", "Server port")
	flag.Parse()

	if envPort := os.Getenv("LZ77_PORT"); envPort != "" {
		*port = envPort
	}

	if !strings.HasPrefix(*port, ":") {
		*port = ":" + *port
	}

	http.HandleFunc("/compress", compressHandler)
	http.HandleFunc("/decompress", decompressHandler)

	log.Printf("LZ77 Compression Server starting on port %s", *port)
	log.Printf("Endpoints: POST /compress, POST /decompress")
	if err := http.ListenAndServe(*port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func compressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.CompressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	originalBytes := []byte(req.Text)
	originalSize := len(originalBytes)

	compressedBytes := lz77.Compress(originalBytes)
	compressedSize := len(compressedBytes)

	ratio := 1.0
	if originalSize > 0 {
		ratio = float64(compressedSize) / float64(originalSize)
	}

	encodedData := base64.StdEncoding.EncodeToString(compressedBytes)

	response := api.CompressResponse{
		CompressedData: encodedData,
		OriginalSize:   originalSize,
		CompressedSize: compressedSize,
		Ratio:          ratio,
		Success:        true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func decompressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.DecompressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	compressedBytes, err := base64.StdEncoding.DecodeString(req.CompressedData)
	if err != nil {
		sendError(w, fmt.Sprintf("Invalid base64 encoding: %v", err), http.StatusBadRequest)
		return
	}

	decompressedBytes := lz77.Decompress(compressedBytes)

	response := api.DecompressResponse{
		Text:    string(decompressedBytes),
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
	})
}
