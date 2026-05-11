package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"sync"

	"jsonschema/common"
	"jsonschema/schema"
)

type Server struct {
	schemas map[string][]byte
	mu      sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		schemas: make(map[string][]byte),
	}
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		sendError(w, http.StatusBadRequest, "schema name is required")
		return
	}

	if len(req.Schema) == 0 {
		sendError(w, http.StatusBadRequest, "schema is required")
		return
	}

	_, err := schema.NewValidator(req.Schema)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid schema: "+err.Error())
		return
	}

	s.mu.Lock()
	s.schemas[req.Name] = req.Schema
	s.mu.Unlock()

	sendJSON(w, http.StatusOK, common.RegisterResponse{
		Success: true,
		Message: "schema registered successfully",
	})
}

func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	s.mu.RLock()
	schemaData, exists := s.schemas[req.SchemaName]
	s.mu.RUnlock()

	if !exists {
		sendError(w, http.StatusNotFound, "schema not found: "+req.SchemaName)
		return
	}

	validator, err := schema.NewValidator(schemaData)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to create validator: "+err.Error())
		return
	}

	errors := validator.Validate(req.Data)

	commonErrors := make([]common.ValidationError, 0, len(errors))
	for _, e := range errors {
		commonErrors = append(commonErrors, common.ValidationError{
			Path:       e.Path,
			Constraint: e.Constraint,
			Actual:     e.Actual,
			Expected:   e.Expected,
		})
	}

	sort.Slice(commonErrors, func(i, j int) bool {
		return commonErrors[i].Path < commonErrors[j].Path
	})

	sendJSON(w, http.StatusOK, common.ValidateResponse{
		Valid:  len(errors) == 0,
		Errors: commonErrors,
	})
}

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, status int, message string) {
	sendJSON(w, status, map[string]interface{}{
		"error": message,
	})
}

func main() {
	server := NewServer()

	http.HandleFunc("/register", server.handleRegister)
	http.HandleFunc("/validate", server.handleValidate)

	addr := ":8504"
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
