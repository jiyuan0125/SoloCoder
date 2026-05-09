package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/concurrent-sortedmap/internal/protocol"
	"github.com/concurrent-sortedmap/pkg/skipmap"
)

type Server struct {
	sm *skipmap.SkipMap
}

func NewServer() *Server {
	return &Server{
		sm: skipmap.New(),
	}
}

func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.PutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := protocol.PutResponse{Success: false, Error: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	_, err := s.sm.Put(req.Key, req.Value)
	resp := protocol.PutResponse{Success: err == nil}
	if err != nil {
		resp.Error = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		resp := protocol.GetResponse{Success: false, Error: "missing key"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	value, exists := s.sm.Get(key)
	resp := protocol.GetResponse{Success: true, Exists: exists, Value: value}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := protocol.DeleteResponse{Success: false, Error: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	deleted := s.sm.Delete(req.Key)
	resp := protocol.DeleteResponse{Success: true, Deleted: deleted}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleRange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	it := s.sm.RangeScan(start, end)
	var pairs []protocol.KVPair
	var key, value string
	for it.Next(&key, &value) {
		pairs = append(pairs, protocol.KVPair{Key: key, Value: value})
	}

	resp := protocol.RangeResponse{Success: true, Pairs: pairs}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleSize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	size := s.sm.Size()
	resp := protocol.SizeResponse{Success: true, Size: size}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	server := NewServer()

	http.HandleFunc("/put", server.handlePut)
	http.HandleFunc("/get", server.handleGet)
	http.HandleFunc("/delete", server.handleDelete)
	http.HandleFunc("/range", server.handleRange)
	http.HandleFunc("/size", server.handleSize)

	addr := ":8080"
	fmt.Printf("Server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
