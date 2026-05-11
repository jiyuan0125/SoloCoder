package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"double-array-trie/pkg/api"
	"double-array-trie/pkg/datrie"
)

type Server struct {
	trie *datrie.Trie
	mu   sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		trie: datrie.NewTrie(),
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

func (s *Server) AddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	var req api.AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.trie.Insert(req.Word); err != nil {
		if err == datrie.ErrWordExists {
			s.writeJSON(w, http.StatusOK, &api.AddResponse{
				Success: false,
				Message: "word already exists",
			})
			return
		}
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.writeJSON(w, http.StatusOK, &api.AddResponse{
		Success: true,
		Message: "word added",
	})
}

func (s *Server) ImportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	var req api.ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	newTrie := datrie.NewTrie()
	added := 0

	for _, word := range req.Words {
		if err := newTrie.Insert(word); err != nil {
			if err == datrie.ErrWordExists {
				continue
			}
			s.writeError(w, http.StatusBadRequest, err)
			return
		}
		added++
	}

	s.trie = newTrie

	s.writeJSON(w, http.StatusOK, &api.ImportResponse{
		Success: true,
		Added:   added,
		Message: "words imported",
	})
}

func (s *Server) SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	word := r.URL.Query().Get("word")
	if word == "" {
		s.writeError(w, http.StatusBadRequest, errors.New("missing word parameter"))
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	exists := s.trie.Search(word)

	s.writeJSON(w, http.StatusOK, &api.SearchResponse{
		Exists: exists,
	})
}

func (s *Server) PrefixHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		s.writeError(w, http.StatusBadRequest, errors.New("missing prefix parameter"))
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	words := s.trie.Prefix(prefix)

	s.writeJSON(w, http.StatusOK, &api.PrefixResponse{
		Success: true,
		Words:   words,
	})
}

func (s *Server) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		s.writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	word := r.URL.Query().Get("word")
	if word == "" {
		s.writeError(w, http.StatusBadRequest, errors.New("missing word parameter"))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.trie.Delete(word); err != nil {
		if err == datrie.ErrWordNotFound {
			s.writeJSON(w, http.StatusOK, &api.DeleteResponse{
				Success: false,
				Message: "word not found",
			})
			return
		}
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.writeJSON(w, http.StatusOK, &api.DeleteResponse{
		Success: true,
		Message: "word deleted",
	})
}

func (s *Server) StatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	stat := s.trie.Stat()

	s.writeJSON(w, http.StatusOK, &api.StatResponse{
		BaseSize:    stat.BaseSize,
		CheckSize:   stat.CheckSize,
		UsedSlots:   stat.UsedSlots,
		WordCount:   stat.WordCount,
		Utilization: stat.Utilization,
	})
}

func getPort() string {
	port := "8403"
	if envPort := os.Getenv("TRIE_PORT"); envPort != "" {
		port = envPort
	}
	flagPort := flag.String("port", "", "server port")
	flag.Parse()
	if *flagPort != "" {
		port = *flagPort
	}
	return port
}

func main() {
	port := getPort()
	server := NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/add", server.AddHandler)
	mux.HandleFunc("/import", server.ImportHandler)
	mux.HandleFunc("/search", server.SearchHandler)
	mux.HandleFunc("/prefix", server.PrefixHandler)
	mux.HandleFunc("/delete", server.DeleteHandler)
	mux.HandleFunc("/stat", server.StatHandler)

	addr := ":" + strings.TrimPrefix(port, ":")
	fmt.Printf("Double-Array Trie server starting on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
