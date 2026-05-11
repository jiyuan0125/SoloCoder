package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"percent-encoding/internal/api"
	"percent-encoding/internal/percent"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "Server port (default: 8080)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/encode/url", handleEncodeURL)
	http.HandleFunc("/encode/uri", handleEncodeURI)
	http.HandleFunc("/encode/form", handleEncodeForm)
	http.HandleFunc("/decode", handleDecode)

	log.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func handleEncodeURL(w http.ResponseWriter, r *http.Request) {
	handleEncode(w, r, percent.ModeURL)
}

func handleEncodeURI(w http.ResponseWriter, r *http.Request) {
	handleEncode(w, r, percent.ModeURI)
}

func handleEncodeForm(w http.ResponseWriter, r *http.Request) {
	handleEncode(w, r, percent.ModeForm)
}

func handleEncode(w http.ResponseWriter, r *http.Request, mode percent.Mode) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.EncodeResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	var req api.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.EncodeResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	var opts percent.EncodeOptions
	if req.Component != "" {
		opts.Component = percent.Component(req.Component)
	}

	output := percent.Encode(req.Input, mode, opts)
	writeJSON(w, http.StatusOK, api.EncodeResponse{
		Success: true,
		Output:  output,
	})
}

func handleDecode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.DecodeResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	var req api.DecodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.DecodeResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	mode := percent.Mode(req.Mode)
	if !percent.IsValidMode(mode) {
		writeJSON(w, http.StatusBadRequest, api.DecodeResponse{
			Success: false,
			Error:   "Invalid mode",
		})
		return
	}

	output, err := percent.Decode(req.Input, mode)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.DecodeResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.DecodeResponse{
		Success: true,
		Output:  output,
	})
}
