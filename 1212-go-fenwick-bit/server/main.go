package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/solocoder/fenwickbit/api"
	"github.com/solocoder/fenwickbit/fenwickbit"
)

type server struct {
	trees map[string]*fenwickbit.FenwickTree
	mu    sync.RWMutex
}

func newServer() *server {
	return &server{
		trees: make(map[string]*fenwickbit.FenwickTree),
	}
}

func (s *server) getTree(id string) (*fenwickbit.FenwickTree, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tree, exists := s.trees[id]
	return tree, exists
}

func (s *server) setTree(id string, tree *fenwickbit.FenwickTree) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.trees[id] = tree
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, api.ErrorResponse{Error: message})
}

func (s *server) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Size < 0 {
		writeError(w, http.StatusBadRequest, "size must be non-negative")
		return
	}

	id := fmt.Sprintf("tree_%d", len(s.trees)+1)
	tree := fenwickbit.New(req.Size)
	s.setTree(id, tree)

	writeJSON(w, http.StatusOK, api.CreateResponse{ID: id})
}

func (s *server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tree, exists := s.getTree(req.ID)
	if !exists {
		writeError(w, http.StatusNotFound, "tree not found")
		return
	}

	if err := tree.Update(req.Index, req.Delta); err != nil {
		writeJSON(w, http.StatusOK, api.UpdateResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, api.UpdateResponse{Success: true})
}

func (s *server) handleQueryPrefix(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.QueryPrefixRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tree, exists := s.getTree(req.ID)
	if !exists {
		writeError(w, http.StatusNotFound, "tree not found")
		return
	}

	result := tree.QueryPrefix(req.Index)
	writeJSON(w, http.StatusOK, api.QueryPrefixResponse{Result: result})
}

func (s *server) handleQueryRange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.QueryRangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tree, exists := s.getTree(req.ID)
	if !exists {
		writeError(w, http.StatusNotFound, "tree not found")
		return
	}

	result, err := tree.QueryRange(req.L, req.R)
	if err != nil {
		writeJSON(w, http.StatusOK, api.QueryRangeResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, api.QueryRangeResponse{Result: result})
}

func (s *server) handleInversions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.InversionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	count := fenwickbit.CountInversions(req.Array)
	writeJSON(w, http.StatusOK, api.InversionsResponse{Count: count})
}

func main() {
	port := flag.String("port", "8080", "server port")
	flag.Parse()

	if envPort := os.Getenv("SERVER_PORT"); envPort != "" {
		*port = envPort
	}

	srv := newServer()

	http.HandleFunc("/create", srv.handleCreate)
	http.HandleFunc("/update", srv.handleUpdate)
	http.HandleFunc("/query-prefix", srv.handleQueryPrefix)
	http.HandleFunc("/query-range", srv.handleQueryRange)
	http.HandleFunc("/inversions", srv.handleInversions)

	addr := ":" + *port
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
