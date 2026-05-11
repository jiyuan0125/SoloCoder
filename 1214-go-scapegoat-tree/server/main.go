package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/scapegoat/common"
	"github.com/scapegoat/scapegoat"
)

type Server struct {
	tree *scapegoat.Tree
	mu   sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		tree: scapegoat.NewTree(),
	}
}

func (s *Server) handleInsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.InsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.tree.Insert(req.Key, req.Value)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	keyStr := r.URL.Query().Get("key")
	var key int64
	if _, err := fmt.Sscanf(keyStr, "%d", &key); err != nil {
		writeError(w, "invalid key", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	value, err := s.tree.Search(key)
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(common.SearchResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.SearchResponse{
		Success: true,
		Value:   value,
	})
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	keyStr := r.URL.Query().Get("key")
	var key int64
	if _, err := fmt.Sscanf(keyStr, "%d", &key); err != nil {
		writeError(w, "invalid key", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	err := s.tree.Delete(key)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(common.DeleteResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.DeleteResponse{
		Success: true,
	})
}

func (s *Server) handleSize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	size := s.tree.Size()
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.SizeResponse{Size: size})
}

func writeError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(common.ErrorResponse{
		Success: false,
		Error:   msg,
	})
}

func getPort() string {
	var port string
	flag.StringVar(&port, "port", "", "server port")
	flag.Parse()

	if port == "" {
		port = os.Getenv("SCAPEGOAT_PORT")
	}

	if port == "" {
		port = "8080"
	}

	return port
}

func main() {
	port := getPort()
	server := NewServer()

	http.HandleFunc("/insert", server.handleInsert)
	http.HandleFunc("/search", server.handleSearch)
	http.HandleFunc("/delete", server.handleDelete)
	http.HandleFunc("/size", server.handleSize)

	fmt.Printf("Scapegoat tree server listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("server error: %v\n", err)
		os.Exit(1)
	}
}
