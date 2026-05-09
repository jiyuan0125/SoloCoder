package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"rsa-crypto/common"
	"rsa-crypto/rsacrypto"
)

type Server struct {
	km *KeyManager
}

func NewServer() *Server {
	return &Server{
		km: NewKeyManager(),
	}
}

func main() {
	server := NewServer()
	http.HandleFunc("/api/keys/generate", server.handleGenerateKeys)
	http.HandleFunc("/api/keys/async-generate", server.handleAsyncGenerateKeys)
	http.HandleFunc("/api/keys/task/", server.handleKeyGenTask)
	http.HandleFunc("/api/keys/info", server.handleKeyInfo)
	http.HandleFunc("/api/encrypt", server.handleEncrypt)
	http.HandleFunc("/api/decrypt", server.handleDecrypt)

	addr := ":8080"
	log.Printf("RSA Crypto Server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func (s *Server) handleGenerateKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.KeyPairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.KeySize != 2048 && req.KeySize != 4096 {
		writeJSONError(w, "key size must be 2048 or 4096", http.StatusBadRequest)
		return
	}

	pub, priv, err := rsacrypto.GenerateKeyPair(req.KeySize)
	if err != nil {
		writeJSONError(w, fmt.Sprintf("failed to generate key pair: %v", err), http.StatusInternalServerError)
		return
	}

	pubPEM, err := rsacrypto.ExportPublicKeyPEM(pub)
	if err != nil {
		writeJSONError(w, fmt.Sprintf("failed to export public key: %v", err), http.StatusInternalServerError)
		return
	}

	privPEM, err := rsacrypto.ExportPrivateKeyPEM(priv)
	if err != nil {
		writeJSONError(w, fmt.Sprintf("failed to export private key: %v", err), http.StatusInternalServerError)
		return
	}

	s.km.SetKeys(pub, priv, req.KeySize)

	resp := common.KeyPairResponse{
		PublicKey:  pubPEM,
		PrivateKey: privPEM,
	}
	writeJSON(w, resp, http.StatusOK)
}

func (s *Server) handleAsyncGenerateKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.AsyncKeyGenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.KeySize != 2048 && req.KeySize != 4096 {
		writeJSONError(w, "key size must be 2048 or 4096", http.StatusBadRequest)
		return
	}

	taskID := s.km.CreateTask(req.KeySize)
	s.km.UpdateTask(taskID, common.TaskStatusRunning, "", "", "")

	go func() {
		pub, priv, err := rsacrypto.GenerateKeyPair(req.KeySize)
		if err != nil {
			s.km.UpdateTask(taskID, common.TaskStatusFailed, "", "", err.Error())
			return
		}

		pubPEM, err := rsacrypto.ExportPublicKeyPEM(pub)
		if err != nil {
			s.km.UpdateTask(taskID, common.TaskStatusFailed, "", "", err.Error())
			return
		}

		privPEM, err := rsacrypto.ExportPrivateKeyPEM(priv)
		if err != nil {
			s.km.UpdateTask(taskID, common.TaskStatusFailed, "", "", err.Error())
			return
		}

		s.km.SetKeys(pub, priv, req.KeySize)
		s.km.UpdateTask(taskID, common.TaskStatusCompleted, pubPEM, privPEM, "")
	}()

	resp := common.AsyncKeyGenResponse{TaskID: taskID}
	writeJSON(w, resp, http.StatusAccepted)
}

func (s *Server) handleKeyGenTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/keys/task/")
	if path == "" {
		writeJSONError(w, "task ID not provided", http.StatusBadRequest)
		return
	}

	task, exists := s.km.GetTask(path)
	if !exists {
		writeJSONError(w, "task not found", http.StatusNotFound)
		return
	}

	writeJSON(w, task, http.StatusOK)
}

func (s *Server) handleKeyInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info := s.km.GetKeyInfo()
	writeJSON(w, info, http.StatusOK)
}

func (s *Server) handleEncrypt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.EncryptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	pub, _ := s.km.GetKeys()
	if pub == nil {
		writeJSONError(w, "no public key available, generate keys first", http.StatusBadRequest)
		return
	}

	ciphertext, err := rsacrypto.Encrypt(pub, []byte(req.Plaintext), req.Padding, req.HashAlgo, req.Label)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := common.EncryptResponse{Ciphertext: ciphertext}
	writeJSON(w, resp, http.StatusOK)
}

func (s *Server) handleDecrypt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.DecryptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	_, priv := s.km.GetKeys()
	if priv == nil {
		writeJSONError(w, "no private key available, generate keys first", http.StatusBadRequest)
		return
	}

	plaintext, err := rsacrypto.Decrypt(priv, req.Ciphertext, req.Padding, req.HashAlgo, req.Label)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := common.DecryptResponse{Plaintext: string(plaintext)}
	writeJSON(w, resp, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, v interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	writeJSON(w, common.ErrorResponse{Error: message}, status)
}
