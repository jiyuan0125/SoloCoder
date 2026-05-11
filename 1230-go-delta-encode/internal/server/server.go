package server

import (
	"encoding/json"
	"net/http"

	"delta-encode/pkg/api"
)

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/encode", s.encode)
	mux.HandleFunc("/decode", s.decode)
	mux.HandleFunc("/stats", s.stats)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, api.ErrorResponse{Error: msg})
}
