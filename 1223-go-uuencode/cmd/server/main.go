package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"uuencode/api"
	"uuencode/uuencode"
)

func main() {
	port := flag.String("port", "", "server port")
	flag.Parse()

	if *port == "" {
		*port = os.Getenv("SERVER_PORT")
	}

	if *port == "" {
		*port = "8080"
	}

	http.HandleFunc("/encode", handleEncode)
	http.HandleFunc("/decode", handleDecode)

	fmt.Printf("Server starting on port %s\n", *port)
	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func handleEncode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid base64 data")
		return
	}

	encoded := uuencode.Encode(data, req.Filename, req.Mode)

	resp := api.EncodeResponse{Encoded: encoded}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleDecode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.DecodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := uuencode.Decode(req.Encoded)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	dataBase64 := base64.StdEncoding.EncodeToString(result.Data)

	resp := api.DecodeResponse{
		Data:     dataBase64,
		Filename: result.Filename,
		Mode:     result.Mode,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(api.ErrorResponse{Error: message})
}
