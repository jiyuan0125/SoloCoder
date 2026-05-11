package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"rleencode/pkg/api"
	"rleencode/pkg/rle8"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "server port (e.g. :8210)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("RLE_SERVER_PORT")
	}
	if port == "" {
		port = ":8210"
	}
	if port[0] != ':' {
		port = ":" + port
	}

	http.HandleFunc("/encode", encodeHandler)
	http.HandleFunc("/decode", decodeHandler)

	fmt.Printf("RLE8 server listening on %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func encodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Width <= 0 || req.Height <= 0 {
		http.Error(w, "width and height must be positive", http.StatusBadRequest)
		return
	}

	data, err := rle8.Encode(req.Pixels)
	if err != nil {
		http.Error(w, fmt.Sprintf("encode error: %v", err), http.StatusInternalServerError)
		return
	}

	resp := api.EncodeResponse{
		Data: hex.EncodeToString(data),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func decodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.DecodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Width <= 0 {
		http.Error(w, "width must be positive", http.StatusBadRequest)
		return
	}

	data, err := hex.DecodeString(req.Data)
	if err != nil {
		http.Error(w, "invalid hex string", http.StatusBadRequest)
		return
	}

	pixels, err := rle8.Decode(data, req.Width)
	if err != nil {
		http.Error(w, fmt.Sprintf("decode error: %v", err), http.StatusInternalServerError)
		return
	}

	resp := api.DecodeResponse{
		Pixels: pixels,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
