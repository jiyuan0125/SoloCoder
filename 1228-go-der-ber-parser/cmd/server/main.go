package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"der-ber-parser/api"
	"der-ber-parser/asn1"
)

func main() {
	port := flag.String("port", "", "server port (default: 8320)")
	flag.Parse()

	serverPort := getPort(*port)

	http.HandleFunc("/encode-der", handleEncodeDER)
	http.HandleFunc("/encode-ber", handleEncodeBER)
	http.HandleFunc("/decode", handleDecode)

	addr := fmt.Sprintf(":%s", serverPort)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func getPort(argPort string) string {
	if argPort != "" {
		return argPort
	}
	if envPort := os.Getenv("PORT"); envPort != "" {
		return envPort
	}
	return "8320"
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func handleEncodeDER(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.EncodeResponse{
			Success: false,
			Error:   "method not allowed, use POST",
		})
		return
	}

	var req api.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.EncodeResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
		return
	}

	hex, err := asn1.EncodeDER(req.Data)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.EncodeResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.EncodeResponse{
		Success: true,
		Hex:     hex,
	})
}

func handleEncodeBER(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.EncodeResponse{
			Success: false,
			Error:   "method not allowed, use POST",
		})
		return
	}

	var req api.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.EncodeResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
		return
	}

	hex, err := asn1.EncodeBER(req.Data)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.EncodeResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.EncodeResponse{
		Success: true,
		Hex:     hex,
	})
}

func handleDecode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.DecodeResponse{
			Success: false,
			Error:   "method not allowed, use POST",
		})
		return
	}

	var req api.DecodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.DecodeResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
		return
	}

	hexStr := strings.TrimSpace(req.Hex)
	if hexStr == "" {
		writeJSON(w, http.StatusBadRequest, api.DecodeResponse{
			Success: false,
			Error:   "hex string is empty",
		})
		return
	}

	data, err := asn1.DecodeHex(hexStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.DecodeResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.DecodeResponse{
		Success: true,
		Data:    data,
	})
}
