package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"ecdsa-sign/pkg/ecdsa"
	"ecdsa-sign/pkg/api"
)

type Server struct {
	signer *ecdsa.Signer
}

func main() {
	port := flag.String("port", "8080", "server port")
	curveName := flag.String("curve", "P-256", "curve name (P-256, P-384, P-521)")
	maxCache := flag.Int("max-cache", 10000, "maximum k-value cache size")
	flag.Parse()

	curve, err := ecdsa.ParseCurveName(*curveName)
	if err != nil {
		log.Fatalf("Failed to parse curve: %v", err)
	}

	signer, err := ecdsa.NewSigner(curve, *maxCache)
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}

	server := &Server{signer: signer}

	mux := http.NewServeMux()
	mux.HandleFunc("/key/generate", server.handleGenerateKey)
	mux.HandleFunc("/key/public", server.handlePublicKey)
	mux.HandleFunc("/sign", server.handleSign)
	mux.HandleFunc("/verify", server.handleVerify)
	mux.HandleFunc("/stats", server.handleStats)

	addr := ":" + *port
	log.Printf("ECDSA Signing Server starting on %s", addr)
	log.Printf("Using curve: %s", server.signer.CurveName())
	log.Printf("Max k-value cache size: %d", *maxCache)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (s *Server) handleGenerateKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, &api.GenerateKeyResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.GenerateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, &api.GenerateKeyResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	curve := s.signer.CurveName()
	if req.Curve != "" {
		var err error
		curve, err = ecdsa.ParseCurveName(req.Curve)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, &api.GenerateKeyResponse{
				Success: false,
				Error:   "invalid curve: " + err.Error(),
			})
			return
		}
	}

	if err := s.signer.GenerateKey(curve); err != nil {
		writeJSON(w, http.StatusInternalServerError, &api.GenerateKeyResponse{
			Success: false,
			Error:   "failed to generate key: " + err.Error(),
		})
		return
	}

	pubKey := s.signer.PublicKey()
	pubKeyStr, err := ecdsa.PublicKeyToBase64(pubKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, &api.GenerateKeyResponse{
			Success: false,
			Error:   "failed to encode public key: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, &api.GenerateKeyResponse{
		Success:   true,
		PublicKey: pubKeyStr,
		Curve:     string(s.signer.CurveName()),
	})
}

func (s *Server) handlePublicKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, &api.PublicKeyResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	pubKey := s.signer.PublicKey()
	if pubKey == nil {
		writeJSON(w, http.StatusNotFound, &api.PublicKeyResponse{
			Success: false,
			Error:   "no key available",
		})
		return
	}

	pubKeyStr, err := ecdsa.PublicKeyToBase64(pubKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, &api.PublicKeyResponse{
			Success: false,
			Error:   "failed to encode public key: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, &api.PublicKeyResponse{
		Success:   true,
		PublicKey: pubKeyStr,
		Curve:     string(s.signer.CurveName()),
	})
}

func (s *Server) handleSign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, &api.SignResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.SignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, &api.SignResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	if req.Message == "" {
		writeJSON(w, http.StatusBadRequest, &api.SignResponse{
			Success: false,
			Error:   "message is required",
		})
		return
	}

	message := []byte(req.Message)
	signature, err := s.signer.Sign(message)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, &api.SignResponse{
			Success: false,
			Error:   "failed to sign: " + err.Error(),
		})
		return
	}

	sigBase64 := base64.StdEncoding.EncodeToString(signature)
	writeJSON(w, http.StatusOK, &api.SignResponse{
		Success:   true,
		Signature: sigBase64,
	})
}

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, &api.VerifyResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, &api.VerifyResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	if req.Message == "" || req.Signature == "" || req.PublicKey == "" {
		writeJSON(w, http.StatusBadRequest, &api.VerifyResponse{
			Success: false,
			Error:   "message, signature, and public_key are required",
		})
		return
	}

	pubKey, err := ecdsa.PublicKeyFromBase64(req.PublicKey)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &api.VerifyResponse{
			Success: false,
			Error:   "invalid public key: " + err.Error(),
		})
		return
	}

	signature, err := base64.StdEncoding.DecodeString(req.Signature)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &api.VerifyResponse{
			Success: false,
			Error:   "invalid signature encoding: " + err.Error(),
		})
		return
	}

	message := []byte(req.Message)
	valid, err := ecdsa.Verify(message, signature, pubKey)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &api.VerifyResponse{
			Success: false,
			Error:   "verification failed: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, &api.VerifyResponse{
		Success: true,
		Valid:   valid,
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, &api.StatsResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	writeJSON(w, http.StatusOK, &api.StatsResponse{
		Success:    true,
		SignCount:  s.signer.SignCount(),
		KCacheSize: s.signer.KCacheSize(),
	})
}
