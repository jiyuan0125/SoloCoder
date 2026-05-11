package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"ascii85-codec/api"
	"ascii85-codec/ascii85"
)

func main() {
	port := flag.String("port", "", "Server port (or use ASCII85_PORT env var)")
	flag.Parse()

	if *port == "" {
		if envPort := os.Getenv("ASCII85_PORT"); envPort != "" {
			*port = envPort
		} else {
			*port = "8305"
		}
	}

	http.HandleFunc("/encode", handleEncode)
	http.HandleFunc("/decode", handleDecode)

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Ascii85 server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}

func handleEncode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	mode := ascii85.ModeAdobe
	if req.Mode == "btoa" {
		mode = ascii85.ModeBtoa
	}

	encoded, err := ascii85.Encode([]byte(req.Data), mode)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := api.EncodeResponse{
		Encoded: encoded,
		Mode:    string(mode),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleDecode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.DecodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	decoded, mode, err := ascii85.Decode(req.Encoded)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := api.DecodeResponse{
		Decoded: string(decoded),
		Mode:    string(mode),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(api.ErrorResponse{Error: message})
}
