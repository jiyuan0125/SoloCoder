package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"stun-tool/pkg/common"
	"stun-tool/pkg/stun"
)

var (
	httpAddr string
)

func main() {
	flag.StringVar(&httpAddr, "addr", ":8080", "HTTP server address")
	flag.Parse()

	http.HandleFunc("/api/detect", handleDetect)
	http.HandleFunc("/api/binding", handleBinding)
	http.HandleFunc("/health", handleHealth)

	log.Printf("STUN server starting on %s", httpAddr)
	if err := http.ListenAndServe(httpAddr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func handleDetect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.DetectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	server := req.GetServer()
	creds := req.GetCredentials()

	opts := make([]stun.DetectorOption, 0)
	opts = append(opts, stun.WithServers(server, server))
	if creds != nil {
		opts = append(opts, stun.WithCredentials(creds))
	}

	detector := stun.NewNATDetector(opts...)
	result, err := detector.Detect()

	var resp common.DetectResponse
	resp.FromNATResult(result, err)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleBinding(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.BindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	server := req.GetServer()
	creds := req.GetCredentials()

	opts := &stun.BindingOptions{
		Server:      server,
		Credentials: creds,
	}

	addr, err := stun.SimpleBinding(opts)

	var resp common.BindingResponse
	resp.FromMappedAddress(addr, err)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"service": "stun-server",
	})
}
