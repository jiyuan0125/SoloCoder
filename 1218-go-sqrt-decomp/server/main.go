package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"sqrtdecomp/api"
	"sqrtdecomp/decomposition"
)

type Server struct {
	data *decomposition.SqrtDecomposition
	mu   sync.RWMutex
}

func NewServer() *Server {
	return &Server{}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{Success: false, Message: "Method not allowed"})
		return
	}

	var req api.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Message: "Invalid request body"})
		return
	}

	if req.Size < 0 {
		respondJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Message: "Size must be non-negative"})
		return
	}

	s.mu.Lock()
	s.data = decomposition.NewSqrtDecomposition(req.Size)
	s.mu.Unlock()

	respondJSON(w, http.StatusOK, api.CreateResponse{Success: true, Message: fmt.Sprintf("Created array of size %d", req.Size)})
}

func (s *Server) handleRangeAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{Success: false, Message: "Method not allowed"})
		return
	}

	if s.data == nil {
		respondJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Message: "Array not created"})
		return
	}

	var req api.RangeAddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Message: "Invalid request body"})
		return
	}

	s.mu.Lock()
	s.data.RangeAdd(req.L, req.R, req.Val)
	s.mu.Unlock()

	respondJSON(w, http.StatusOK, api.RangeAddResponse{Success: true, Message: fmt.Sprintf("Added %d to [%d, %d]", req.Val, req.L, req.R)})
}

func (s *Server) handleRangeSum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{Success: false, Message: "Method not allowed"})
		return
	}

	if s.data == nil {
		respondJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Message: "Array not created"})
		return
	}

	var req api.RangeSumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Message: "Invalid request body"})
		return
	}

	s.mu.RLock()
	sum := s.data.RangeSum(req.L, req.R)
	s.mu.RUnlock()

	respondJSON(w, http.StatusOK, api.RangeSumResponse{Success: true, Sum: sum, Message: "Success"})
}

func (s *Server) handleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{Success: false, Message: "Method not allowed"})
		return
	}

	if s.data == nil {
		respondJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Message: "Array not created"})
		return
	}

	var req api.SetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Message: "Invalid request body"})
		return
	}

	s.mu.Lock()
	s.data.Set(req.Index, req.Val)
	s.mu.Unlock()

	respondJSON(w, http.StatusOK, api.SetResponse{Success: true, Message: fmt.Sprintf("Set index %d to %d", req.Index, req.Val)})
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{Success: false, Message: "Method not allowed"})
		return
	}

	if s.data == nil {
		respondJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Message: "Array not created"})
		return
	}

	var req api.GetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Message: "Invalid request body"})
		return
	}

	s.mu.RLock()
	value := s.data.Get(req.Index)
	s.mu.RUnlock()

	respondJSON(w, http.StatusOK, api.GetResponse{Success: true, Value: value, Message: "Success"})
}

func (s *Server) handleDump(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{Success: false, Message: "Method not allowed"})
		return
	}

	if s.data == nil {
		respondJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Message: "Array not created"})
		return
	}

	s.mu.RLock()
	data := s.data.Dump()
	s.mu.RUnlock()

	respondJSON(w, http.StatusOK, api.DumpResponse{Success: true, Data: data, Message: "Success"})
}

func getPort() string {
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	if len(os.Args) > 1 {
		if _, err := strconv.Atoi(os.Args[1]); err == nil {
			port = os.Args[1]
		}
	}
	return port
}

func main() {
	server := NewServer()

	http.HandleFunc("/create", server.handleCreate)
	http.HandleFunc("/rangeadd", server.handleRangeAdd)
	http.HandleFunc("/rangesum", server.handleRangeSum)
	http.HandleFunc("/set", server.handleSet)
	http.HandleFunc("/get", server.handleGet)
	http.HandleFunc("/dump", server.handleDump)

	port := getPort()
	addr := ":" + port

	log.Printf("Server starting on port %s", port)
	log.Printf("Endpoints: %s", strings.Join([]string{"/create", "/rangeadd", "/rangesum", "/set", "/get", "/dump"}, ", "))
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
