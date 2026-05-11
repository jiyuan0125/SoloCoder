package main

import (
	"encoding/json"
	"net/http"

	"green-care-management/pkg/core"
)

type Server struct {
	service *core.Service
	mux     *http.ServeMux
}

func NewServer(service *core.Service) *Server {
	s := &Server{
		service: service,
		mux:     http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]interface{}{
		"success": false,
		"message": message,
	})
}
