package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"binproto/core"
	"binproto/pkg/proto"
)

type FormatStore struct {
	mu      sync.RWMutex
	formats map[string]string
}

func NewFormatStore() *FormatStore {
	return &FormatStore{
		formats: make(map[string]string),
	}
}

func (s *FormatStore) Register(formatID, dsl string) error {
	if err := core.ValidateDSL(dsl); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.formats[formatID] = dsl
	return nil
}

func (s *FormatStore) Get(formatID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	dsl, ok := s.formats[formatID]
	return dsl, ok
}

type Server struct {
	store *FormatStore
}

func NewServer(store *FormatStore) *Server {
	return &Server{store: store}
}

func (s *Server) RegisterFormat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req proto.RegisterFormatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := proto.RegisterFormatResponse{Success: false, Error: "invalid request body"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if err := s.store.Register(req.FormatID, req.DSL); err != nil {
		resp := proto.RegisterFormatResponse{Success: false, Error: err.Error()}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := proto.RegisterFormatResponse{Success: true}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) Parse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req proto.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := proto.ParseResponse{Success: false, Error: "invalid request body"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	dsl, ok := s.store.Get(req.FormatID)
	if !ok {
		resp := proto.ParseResponse{Success: false, Error: "format not found"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	def, err := core.ParseDSL(dsl)
	if err != nil {
		resp := proto.ParseResponse{Success: false, Error: err.Error()}
		json.NewEncoder(w).Encode(resp)
		return
	}

	result, err := core.ParseBinary(def, req.Data)
	if err != nil {
		resp := proto.ParseResponse{Success: false, Error: err.Error()}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := proto.ParseResponse{Success: true, Data: result.Data}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) Serialize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req proto.SerializeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := proto.SerializeResponse{Success: false, Error: "invalid request body"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	dsl, ok := s.store.Get(req.FormatID)
	if !ok {
		resp := proto.SerializeResponse{Success: false, Error: "format not found"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	def, err := core.ParseDSL(dsl)
	if err != nil {
		resp := proto.SerializeResponse{Success: false, Error: err.Error()}
		json.NewEncoder(w).Encode(resp)
		return
	}

	data, err := core.SerializeBinary(def, req.Data)
	if err != nil {
		resp := proto.SerializeResponse{Success: false, Error: err.Error()}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := proto.SerializeResponse{Success: true, Data: data}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) Validate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req proto.ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := proto.ValidateResponse{Success: false, Error: "invalid request body"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if err := core.ValidateDSL(req.DSL); err != nil {
		resp := proto.ValidateResponse{Success: false, Error: err.Error()}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := proto.ValidateResponse{Success: true}
	json.NewEncoder(w).Encode(resp)
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "server port")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		fmt.Sscanf(envPort, "%d", &port)
	}

	store := NewFormatStore()
	server := NewServer(store)

	http.HandleFunc("/register", server.RegisterFormat)
	http.HandleFunc("/parse", server.Parse)
	http.HandleFunc("/serialize", server.Serialize)
	http.HandleFunc("/validate", server.Validate)

	log.Printf("server listening on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
