package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	"ecdh-demo/internal/common"
	"ecdh-demo/pkg/ecdh"
)

func generateSessionID() string {
	buf := make([]byte, 16)
	rand.Read(buf)
	return hex.EncodeToString(buf)
}

func toECDH(c string) ecdh.Curve {
	switch c {
	case common.CurveP256:
		return ecdh.CurveP256
	case common.CurveP384:
		return ecdh.CurveP384
	case common.CurveP521:
		return ecdh.CurveP521
	default:
		return ""
	}
}

func fromECDH(c ecdh.Curve) string {
	switch c {
	case ecdh.CurveP256:
		return common.CurveP256
	case ecdh.CurveP384:
		return common.CurveP384
	case ecdh.CurveP521:
		return common.CurveP521
	default:
		return ""
	}
}

type Server struct {
	sessions *ecdh.SessionManager
}

func NewServer() *Server {
	return &Server{
		sessions: ecdh.NewSessionManager(),
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(common.ErrorResponse{Error: msg})
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.StartNegotiationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	curve := toECDH(req.Curve)
	if err := ecdh.ValidateCurve(curve); err != nil {
		writeError(w, http.StatusBadRequest, "unsupported curve")
		return
	}

	kp, err := ecdh.GenerateKeyPair(curve)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate key pair")
		return
	}

	sessionID := generateSessionID()
	s.sessions.Add(sessionID, kp)

	resp := common.StartNegotiationResponse{
		SessionID: sessionID,
		Curve:     fromECDH(kp.Curve),
		PublicKey: base64.StdEncoding.EncodeToString(kp.PublicKey),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.CompleteNegotiationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	kp, ok := s.sessions.Get(req.SessionID)
	if !ok {
		writeError(w, http.StatusNotFound, "session not found or expired")
		return
	}

	curve := toECDH(req.Curve)
	if err := ecdh.ValidateCurve(curve); err != nil {
		writeError(w, http.StatusBadRequest, "unsupported curve")
		return
	}

	peerPubBytes, err := base64.StdEncoding.DecodeString(req.PublicKey)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid base64 public key")
		return
	}

	sharedSecret, err := ecdh.ComputeSharedSecretFromBytes(kp, curve, peerPubBytes)
	if err != nil {
		switch {
		case errors.Is(err, ecdh.ErrCurveMismatch):
			writeError(w, http.StatusBadRequest, "curve mismatch between session and request")
		case errors.Is(err, ecdh.ErrInvalidPublicKey):
			writeError(w, http.StatusBadRequest, "invalid public key for claimed curve")
		default:
			writeError(w, http.StatusInternalServerError, "failed to compute shared secret")
		}
		return
	}

	s.sessions.Remove(req.SessionID)

	resp := common.CompleteNegotiationResponse{
		SessionID:    req.SessionID,
		SharedSecret: sharedSecret,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func main() {
	server := NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/negotiate/start", server.handleStart)
	mux.HandleFunc("/negotiate/complete", server.handleComplete)

	addr := ":8080"
	http.ListenAndServe(addr, mux)
}
