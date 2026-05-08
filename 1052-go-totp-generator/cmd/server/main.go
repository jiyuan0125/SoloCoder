package main

import (
	"encoding/json"
	"log"
	"net/http"

	"totp-generator/internal/common"
	"totp-generator/internal/totp"
)

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Account == "" {
		http.Error(w, "account is required", http.StatusBadRequest)
		return
	}

	secret, err := totp.GenerateSecret()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := common.RegisterResponse{
		Secret:    secret,
		QRCodeURL: totp.GenerateOTPURL(secret, req.Account, req.Issuer),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Secret == "" {
		http.Error(w, "secret is required", http.StatusBadRequest)
		return
	}

	code, err := totp.Generate(req.Secret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := common.GenerateResponse{TOTP: code}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Secret == "" || req.TOTP == "" {
		http.Error(w, "secret and totp are required", http.StatusBadRequest)
		return
	}

	var valid bool
	if req.Window > 0 {
		valid = totp.ValidateWithWindow(req.Secret, req.TOTP, req.Window)
	} else {
		valid = totp.Validate(req.Secret, req.TOTP)
	}

	resp := common.ValidateResponse{Valid: valid}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/generate", generateHandler)
	http.HandleFunc("/validate", validateHandler)

	log.Println("server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
