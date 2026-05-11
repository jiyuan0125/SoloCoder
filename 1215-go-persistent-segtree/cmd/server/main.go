package main

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"os"
	"sync"

	"persistent-segtree/pkg/api"
	"persistent-segtree/pkg/segtree"
)

type Server struct {
	tree *segtree.PersistentSegmentTree
	mu   sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		tree: segtree.New(),
	}
}

func (s *Server) handleBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.BuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.BuildResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.tree.Build(req.Values)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.BuildResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.BuildResponse{
		Success:      true,
		VersionCount: s.tree.VersionCount(),
		Message:      "success",
	})
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.tree.VersionCount() <= 1 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.QueryKthResponse{
			Success: false,
			Message: "segment tree not built yet",
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.QueryKthResponse{
			Success: false,
			Message: "invalid request",
		})
		return
	}

	var req api.QueryKthRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.QueryKthResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	value, err := s.tree.QueryKth(req.L, req.R, req.K)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.QueryKthResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.QueryKthResponse{
		Success: true,
		Value:   value,
		Message: "success",
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	ready := s.tree.VersionCount() > 1
	length := 0
	if ready {
		length = s.tree.VersionCount() - 1
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.StatusResponse{
		Ready:       ready,
		ArrayLength: length,
	})
}

func getPort() string {
	port := flag.String("port", "", "port to listen on (default: 8080)")
	flag.Parse()

	if *port != "" {
		return ":" + *port
	}

	envPort := os.Getenv("PORT")
	if envPort != "" {
		return ":" + envPort
	}

	return ":8080"
}

func main() {
	server := NewServer()
	port := getPort()

	http.HandleFunc("/api/build", server.handleBuild)
	http.HandleFunc("/api/query", server.handleQuery)
	http.HandleFunc("/api/status", server.handleStatus)

	println("server listening on port", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		println("server error:", err.Error())
		os.Exit(1)
	}
}
