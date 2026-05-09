package main

import (
	"encoding/json"
	"net/http"

	"aes-gcm-crypto/common"
	"aes-gcm-crypto/crypto"
)

func handleEncrypt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.EncryptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.EncryptResponse{Error: "Invalid request body"})
		return
	}

	ciphertext, err := crypto.Encrypt(req.Plaintext, req.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.EncryptResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.EncryptResponse{Ciphertext: ciphertext})
}

func handleDecrypt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.DecryptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.DecryptResponse{Error: "Invalid request body"})
		return
	}

	plaintext, err := crypto.Decrypt(req.Ciphertext, req.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.DecryptResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.DecryptResponse{Plaintext: plaintext})
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ConfigResponse{Error: "Invalid request body"})
		return
	}

	if req.Iterations <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ConfigResponse{Error: "Iterations must be positive"})
		return
	}

	crypto.SetIterations(req.Iterations)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.ConfigResponse{Iterations: req.Iterations})
}

func main() {
	http.HandleFunc("/encrypt", handleEncrypt)
	http.HandleFunc("/decrypt", handleDecrypt)
	http.HandleFunc("/config", handleConfig)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
