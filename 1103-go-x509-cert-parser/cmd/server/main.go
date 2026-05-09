package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"x509cert/pkg/certparse"
	"x509cert/pkg/common"
)

const (
	defaultPort = "8080"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	http.HandleFunc("/api/parse", handleParse)
	http.HandleFunc("/api/verify", handleVerify)

	addr := ":" + port
	log.Printf("Server starting on %s...", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req common.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := common.ParseResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	certInfo, err := certparse.ParseCertificateFromPEM(req.PEM)
	if err != nil {
		response := common.ParseResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := common.ParseResponse{
		Success:     true,
		Certificate: certInfo,
	}
	json.NewEncoder(w).Encode(response)
}

func handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req common.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := common.VerifyResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	verifyResult, err := certparse.VerifyCertificateChain(req.PEMChain)
	if err != nil {
		response := common.VerifyResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	json.NewEncoder(w).Encode(verifyResult)
}
