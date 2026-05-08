package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/solocoder/file-lock-mutex/api"
	"github.com/solocoder/file-lock-mutex/lock"
)

type Server struct {
	locks map[string]*lock.Lock
	mu    sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		locks: make(map[string]*lock.Lock),
	}
}

func (s *Server) getOrCreateLock(filePath string) *lock.Lock {
	s.mu.Lock()
	defer s.mu.Unlock()

	l, ok := s.locks[filePath]
	if !ok {
		l = lock.NewLock(filePath)
		s.locks[filePath] = l
	}
	return l
}

func (s *Server) handleLock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.LockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.FilePath == "" {
		writeError(w, "file_path is required", http.StatusBadRequest)
		return
	}

	l := s.getOrCreateLock(req.FilePath)

	var err error
	if req.Mode == "exclusive" || req.Mode == "write" {
		err = l.LockExclusive(req.Timeout)
	} else if req.Mode == "shared" || req.Mode == "read" {
		err = l.LockShared(req.Timeout)
	} else {
		writeError(w, "invalid mode: must be 'exclusive' or 'shared'", http.StatusBadRequest)
		return
	}

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, api.LockResponse{
		Success: true,
		LockID:  req.FilePath,
	})
}

func (s *Server) handleUnlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.UnlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.FilePath == "" {
		writeError(w, "file_path is required", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	l, ok := s.locks[req.FilePath]
	s.mu.RUnlock()

	if !ok {
		writeError(w, "lock not found", http.StatusNotFound)
		return
	}

	if err := l.Unlock(); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, api.UnlockResponse{
		Success: true,
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.StatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.FilePath == "" {
		writeError(w, "file_path is required", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	l, ok := s.locks[req.FilePath]
	s.mu.RUnlock()

	var resp api.StatusResponse
	resp.LockFilePath = lock.GetLockFilePathFor(req.FilePath)

	if !ok {
		resp.Success = true
		resp.IsHeld = false
		writeJSON(w, resp)
		return
	}

	resp.Success = true
	resp.IsHeld = l.IsHeld()
	if resp.IsHeld {
		if l.Mode() == lock.LockModeExclusive {
			resp.Mode = "exclusive"
		} else {
			resp.Mode = "shared"
		}
	}

	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(api.ErrorResponse{
		Success: false,
		Error:   msg,
	})
}

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	s := NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/lock", s.handleLock)
	mux.HandleFunc("/unlock", s.handleUnlock)
	mux.HandleFunc("/status", s.handleStatus)

	log.Printf("file lock server listening on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
