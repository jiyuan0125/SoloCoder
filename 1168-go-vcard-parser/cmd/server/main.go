package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"vcard-parser/api"
	"vcard-parser/vcard"
)

func getPort() string {
	port := os.Getenv("VCARD_SERVER_PORT")
	if port != "" {
		return port
	}
	flagPort := flag.String("port", "8080", "HTTP server port (can also be set via VCARD_SERVER_PORT env var)")
	flag.Parse()
	return *flagPort
}

func parseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(api.ParseResponse{
			Success: false,
			Error:   "method not allowed, use POST",
		})
		return
	}

	contentType := r.Header.Get("Content-Type")
	var vcardText string

	if contentType == "text/vcard" || contentType == "text/plain" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ParseResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to read request body: %v", err),
			})
			return
		}
		vcardText = string(body)
	} else {
		var req api.ParseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ParseResponse{
				Success: false,
				Error:   fmt.Sprintf("invalid JSON request: %v", err),
			})
			return
		}
		vcardText = req.VCardText
	}

	if vcardText == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ParseResponse{
			Success: false,
			Error:   "vcard text is empty",
		})
		return
	}

	contacts, err := vcard.ParseString(vcardText)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(api.ParseResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to parse vcard: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.ParseResponse{
		Success:  true,
		Contacts: contacts,
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func main() {
	port := getPort()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/parse", parseHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	fmt.Printf("vCard parser server starting on port %s...\n", port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  GET  /health    - Health check\n")
	fmt.Printf("  POST /parse     - Parse vCard (JSON or text/vcard)\n")
	fmt.Println()

	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
