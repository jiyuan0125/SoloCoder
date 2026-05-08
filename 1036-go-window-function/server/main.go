package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"windowfunc/api"
	"windowfunc/pkg/windowfunc"
)

type Server struct {
	engine *windowfunc.Engine
	mu     sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		engine: windowfunc.NewEngine(),
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, api.ErrorResponse{
		Success: false,
		Error:   msg,
	})
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()
	var req api.UploadRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.engine.UploadDataset(req.Dataset, req.Data); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, api.UploadResponse{
		Success: true,
		Message: fmt.Sprintf("dataset '%s' uploaded successfully with %d records", req.Dataset, len(req.Data)),
	})
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	infos := s.engine.ListDatasets()
	writeJSON(w, http.StatusOK, api.ListResponse{
		Success:  true,
		Datasets: infos,
	})
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()
	var req api.QueryRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result, err := s.engine.Query(req.Dataset, req.Functions)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, api.QueryResponse{
		Success: true,
		Data:    result,
	})
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP server address")
	flag.Parse()
	s := NewServer()
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", s.handleUpload)
	mux.HandleFunc("/list", s.handleList)
	mux.HandleFunc("/query", s.handleQuery)
	fmt.Printf("windowfunc server listening on %s\n", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		fmt.Printf("server error: %v\n", err)
	}
}
