package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/lzwtool/lzwcompress/pkg/api"
	"github.com/lzwtool/lzwcompress/pkg/lzw"
)

func compressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.CompressRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.MinCodeSize < 2 || req.MinCodeSize > 8 {
		resp := api.CompressResponse{Error: "minCodeSize must be between 2 and 8"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	var buf bytes.Buffer
	enc, err := lzw.NewEncoder(req.MinCodeSize, &buf)
	if err != nil {
		resp := api.CompressResponse{Error: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	err = enc.Compress(bytes.NewReader(req.Data))
	if err != nil {
		resp := api.CompressResponse{Error: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := api.CompressResponse{
		Data:        buf.Bytes(),
		MinCodeSize: req.MinCodeSize,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func decompressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.DecompressRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.MinCodeSize < 2 || req.MinCodeSize > 8 {
		resp := api.DecompressResponse{Error: "minCodeSize must be between 2 and 8"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	var buf bytes.Buffer
	dec, err := lzw.NewDecoder(req.MinCodeSize, bytes.NewReader(req.Data))
	if err != nil {
		resp := api.DecompressResponse{Error: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	err = dec.Decompress(&buf)
	if err != nil {
		resp := api.DecompressResponse{Error: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := api.DecompressResponse{Data: buf.Bytes()}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc(api.EndpointCompress, compressHandler)
	http.HandleFunc(api.EndpointDecompress, decompressHandler)

	println("server listening on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
