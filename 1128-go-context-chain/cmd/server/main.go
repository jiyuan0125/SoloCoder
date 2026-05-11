package main

import (
	"encoding/json"
	"log"
	"net/http"

	"context-chain/pkg/api"
	"context-chain/pkg/ctxmanager"
)

type Server struct {
	manager *ctxmanager.Manager
}

func NewServer() *Server {
	return &Server{
		manager: ctxmanager.NewManager(),
	}
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, status int, err error) {
	s.writeJSON(w, status, &api.ErrorResponse{Error: err.Error()})
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, nil)
		return
	}

	var req api.RequestCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	resp, err := s.manager.CreateContext(r.Context(), req.Timeout, req.Values, req.ParentID)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, nil)
		return
	}

	var req api.RequestCancel
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	resp, err := s.manager.CancelContext(req.CancelID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err)
		return
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleChain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, nil)
		return
	}

	resp := s.manager.GetCancelChain()
	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handlePrecision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, nil)
		return
	}

	var req api.RequestPrecision
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	resp, err := s.manager.TestPrecision(req.TargetTimeout, req.Iterations)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleLeak(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, nil)
		return
	}

	var req api.RequestLeak
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	resp := s.manager.CheckLeak(req.RequestID)
	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, nil)
		return
	}

	resp := s.manager.GetStatus()
	s.writeJSON(w, http.StatusOK, resp)
}

func main() {
	s := NewServer()

	http.HandleFunc("/create", s.handleCreate)
	http.HandleFunc("/cancel", s.handleCancel)
	http.HandleFunc("/chain", s.handleChain)
	http.HandleFunc("/precision", s.handlePrecision)
	http.HandleFunc("/leak", s.handleLeak)
	http.HandleFunc("/status", s.handleStatus)

	addr := ":8080"
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
