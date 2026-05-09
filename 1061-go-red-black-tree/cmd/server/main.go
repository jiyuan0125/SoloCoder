package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/solocoder/rbtree/pkg/api"
	"github.com/solocoder/rbtree/pkg/rbtree"
)

type Server struct {
	tree *rbtree.Tree
	mu   sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		tree: rbtree.New(),
	}
}

func (s *Server) handleInsert(w http.ResponseWriter, r *http.Request) {
	var req api.InsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	s.mu.Lock()
	s.tree.Insert(req.Key)
	s.mu.Unlock()

	resp := api.InsertResponse{Success: true}
	sendJSON(w, http.StatusOK, resp)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	var req api.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	s.mu.Lock()
	removed := s.tree.Delete(req.Key)
	s.mu.Unlock()

	resp := api.DeleteResponse{Success: removed}
	if !removed {
		resp.Message = "key not found"
	}
	sendJSON(w, http.StatusOK, resp)
}

func (s *Server) handleRangeQuery(w http.ResponseWriter, r *http.Request) {
	var req api.RangeQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	s.mu.RLock()
	values := s.tree.RangeQuery(req.Low, req.High)
	s.mu.RUnlock()

	resp := api.RangeQueryResponse{
		Success: true,
		Values:  values,
	}
	sendJSON(w, http.StatusOK, resp)
}

func (s *Server) handleKthLargest(w http.ResponseWriter, r *http.Request) {
	var req api.KthLargestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	s.mu.RLock()
	value, err := s.tree.KthLargest(req.K)
	s.mu.RUnlock()

	if err != nil {
		resp := api.KthLargestResponse{
			Success: false,
			Message: err.Error(),
		}
		sendJSON(w, http.StatusOK, resp)
		return
	}

	resp := api.KthLargestResponse{
		Success: true,
		Value:   value,
	}
	sendJSON(w, http.StatusOK, resp)
}

func (s *Server) handleSize(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	size := s.tree.Size()
	s.mu.RUnlock()

	resp := api.SizeResponse{
		Success: true,
		Size:    size,
	}
	sendJSON(w, http.StatusOK, resp)
}

func sendJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
	})
}

func main() {
	server := NewServer()

	http.HandleFunc("/insert", server.handleInsert)
	http.HandleFunc("/delete", server.handleDelete)
	http.HandleFunc("/range", server.handleRangeQuery)
	http.HandleFunc("/kthlargest", server.handleKthLargest)
	http.HandleFunc("/size", server.handleSize)

	fmt.Println("Server starting on :8080...")
	http.ListenAndServe(":8080", nil)
}
