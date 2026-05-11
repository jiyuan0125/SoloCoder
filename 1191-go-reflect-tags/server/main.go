package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect-tags/api"
	"reflect-tags/tagparser"
	"sync"
)

type Registry struct {
	mu    sync.RWMutex
	defs  map[string]tagparser.StructDef
	tags  map[string][]tagparser.FieldTags
}

func NewRegistry() *Registry {
	return &Registry{
		defs: make(map[string]tagparser.StructDef),
		tags: make(map[string][]tagparser.FieldTags),
	}
}

func (r *Registry) Register(def tagparser.StructDef) {
	r.mu.Lock()
	defer r.mu.Unlock()

	parser := tagparser.NewParser()
	parsedTags := parser.ParseFromDef(def)

	r.defs[def.Name] = def
	r.tags[def.Name] = parsedTags
}

func (r *Registry) Query(structName string, fieldPath string) ([]tagparser.FieldTags, *tagparser.FieldTags, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tags, ok := r.tags[structName]
	if !ok {
		return nil, nil, false
	}

	if fieldPath == "" {
		return tags, nil, true
	}

	var result *tagparser.FieldTags
	for _, t := range tags {
		if t.Path == fieldPath {
			result = &t
			break
		}
	}

	return tags, result, true
}

type Server struct {
	registry *Registry
}

func NewServer(registry *Registry) *Server {
	return &Server{
		registry: registry,
	}
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Struct.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "struct name is required")
		return
	}

	s.registry.Register(req.Struct)

	writeJSON(w, http.StatusOK, api.RegisterResponse{
		Success: true,
		Message: fmt.Sprintf("struct %s registered successfully", req.Struct.Name),
	})
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.StructName == "" {
		writeJSONError(w, http.StatusBadRequest, "struct name is required")
		return
	}

	allTags, field, ok := s.registry.Query(req.StructName, req.FieldPath)
	if !ok {
		writeJSONError(w, http.StatusNotFound, fmt.Sprintf("struct %s not found", req.StructName))
		return
	}

	resp := api.QueryResponse{
		Success: true,
		All:     allTags,
	}

	if req.FieldPath != "" {
		if field == nil {
			writeJSONError(w, http.StatusNotFound, fmt.Sprintf("field %s not found in struct %s", req.FieldPath, req.StructName))
			return
		}
		resp.Field = field
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, api.ErrorResponse{
		Error: message,
	})
}

func getPort() string {
	port := flag.String("port", "", "server port")
	flag.Parse()

	if *port != "" {
		return *port
	}

	if envPort := os.Getenv("SERVER_PORT"); envPort != "" {
		return envPort
	}

	return "8080"
}

func main() {
	registry := NewRegistry()
	server := NewServer(registry)

	http.HandleFunc("/register", server.handleRegister)
	http.HandleFunc("/query", server.handleQuery)

	port := getPort()
	addr := ":" + port

	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
