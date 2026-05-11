package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"urlencoder/common"
	"urlencoder/encoder"
)

func getPort() string {
	port := flag.String("port", "", "server port")
	flag.Parse()

	if *port != "" {
		return ":" + strings.TrimPrefix(*port, ":")
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		return ":" + strings.TrimPrefix(envPort, ":")
	}

	return ":8080"
}

func encodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(common.EncodeResponse{Error: "method not allowed"})
		return
	}

	var req common.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.EncodeResponse{Error: "invalid request body"})
		return
	}

	result := encoder.Encode(req.Text)
	json.NewEncoder(w).Encode(common.EncodeResponse{Result: result})
}

func decodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(common.DecodeResponse{Error: "method not allowed"})
		return
	}

	var req common.DecodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.DecodeResponse{Error: "invalid request body"})
		return
	}

	result, err := encoder.Decode(req.Text)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.DecodeResponse{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(common.DecodeResponse{Result: result})
}

func batchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(common.BatchResponse{Errors: []string{"method not allowed"}})
		return
	}

	var req common.BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.BatchResponse{Errors: []string{"invalid request body"}})
		return
	}

	operation := strings.ToLower(strings.TrimSpace(req.Operation))
	if operation != "encode" && operation != "decode" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.BatchResponse{Errors: []string{"operation must be 'encode' or 'decode'"}})
		return
	}

	results := make([]string, len(req.Texts))
	errors := make([]string, len(req.Texts))
	hasError := false

	for i, text := range req.Texts {
		if operation == "encode" {
			results[i] = encoder.Encode(text)
			errors[i] = ""
		} else {
			result, err := encoder.Decode(text)
			if err != nil {
				results[i] = ""
				errors[i] = err.Error()
				hasError = true
			} else {
				results[i] = result
				errors[i] = ""
			}
		}
	}

	if hasError {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(common.BatchResponse{Results: results, Errors: errors})
}

func main() {
	port := getPort()

	mux := http.NewServeMux()
	mux.HandleFunc("/encode", encodeHandler)
	mux.HandleFunc("/decode", decodeHandler)
	mux.HandleFunc("/batch", batchHandler)

	log.Printf("server listening on %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
