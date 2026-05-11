package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"lz4service/internal/api"
	"lz4service/internal/lz4"
)

func main() {
	http.HandleFunc("/compress", compressHandler)
	http.HandleFunc("/decompress", decompressHandler)

	fmt.Println("LZ4 compression server starting on port 8080...")
	if err := http.ListenAndServe(":8300", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

func compressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.CompressRequest
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/octet-stream" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			sendError(w, err.Error())
			return
		}
		req.Data = body
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, err.Error())
			return
		}

		if req.Text != "" {
			req.Data = []byte(req.Text)
		} else if req.FilePath != "" {
			absPath, err := filepath.Abs(req.FilePath)
			if err != nil {
				sendError(w, fmt.Sprintf("Invalid file path: %v", err))
				return
			}
			data, err := os.ReadFile(absPath)
			if err != nil {
				sendError(w, fmt.Sprintf("Failed to read file: %v", err))
				return
			}
			req.Data = data
		} else {
			req.Data = []byte{}
		}
	}

	compressed, err := lz4.CompressFrame(req.Data)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	resp := api.CompressResponse{
		Compressed: compressed,
		Success:    true,
		Size:       len(compressed),
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("X-Content-Length", fmt.Sprintf("%d", len(compressed)))
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Compressed)
}

func decompressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.DecompressRequest
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/octet-stream" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			sendError(w, err.Error())
			return
		}
		req.Data = body
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, err.Error())
			return
		}

		if req.FilePath != "" {
			absPath, err := filepath.Abs(req.FilePath)
			if err != nil {
				sendError(w, fmt.Sprintf("Invalid file path: %v", err))
				return
			}
			data, err := os.ReadFile(absPath)
			if err != nil {
				sendError(w, fmt.Sprintf("Failed to read file: %v", err))
				return
			}
			req.Data = data
		} else {
			req.Data = []byte{}
		}
	}

	decompressed, err := lz4.DecompressFrame(req.Data)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	resp := api.DecompressResponse{
		Decompressed: decompressed,
		Success:      true,
		Size:         len(decompressed),
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("X-Content-Length", fmt.Sprintf("%d", len(decompressed)))
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Decompressed)
}

func sendError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(api.CompressResponse{
		Success: false,
		Error:   message,
	})
}
