package main

import (
	"encoding/json"
	"net/http"
	"sync"

	"cms/api"
	"cms/countmin"
)

type Server struct {
	mu       sync.RWMutex
	sketches map[string]*countmin.Sketch
}

func NewServer() *Server {
	return &Server{
		sketches: make(map[string]*countmin.Sketch),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.URL.Path {
	case "/create":
		s.handleCreate(w, r)
	case "/add":
		s.handleAdd(w, r)
	case "/query":
		s.handleQuery(w, r)
	case "/merge":
		s.handleMerge(w, r)
	case "/reset":
		s.handleReset(w, r)
	case "/stats":
		s.handleStats(w, r)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sketches[req.Name]; exists {
		writeError(w, http.StatusConflict, "sketch already exists")
		return
	}

	width := countmin.DefaultWidth
	if req.Width > 0 {
		width = req.Width
	}

	depth := countmin.DefaultDepth
	if req.Depth > 0 {
		depth = req.Depth
	}

	sketch, err := countmin.New(width, depth)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.sketches[req.Name] = sketch

	resp := api.CreateResponse{
		Success: true,
		Message: "created",
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	sketch, exists := s.sketches[req.Name]
	if !exists {
		writeError(w, http.StatusNotFound, "sketch not found")
		return
	}

	for _, item := range req.Items {
		sketch.AddString(item.Item, item.Count)
	}

	resp := api.AddResponse{Success: true}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	sketch, exists := s.sketches[req.Name]
	if !exists {
		writeError(w, http.StatusNotFound, "sketch not found")
		return
	}

	results := make([]api.QueryResult, 0, len(req.Items))
	for _, item := range req.Items {
		est := sketch.QueryString(item)
		results = append(results, api.QueryResult{
			Item:       item,
			Frequency:  est.Frequency,
			ErrorBound: est.ErrorBound,
		})
	}

	resp := api.QueryResponse{
		Success: true,
		Results: results,
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleMerge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.MergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Source == "" || req.Target == "" {
		writeError(w, http.StatusBadRequest, "source and target are required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	source, exists := s.sketches[req.Source]
	if !exists {
		writeError(w, http.StatusNotFound, "source sketch not found")
		return
	}

	target, exists := s.sketches[req.Target]
	if !exists {
		writeError(w, http.StatusNotFound, "target sketch not found")
		return
	}

	if err := target.Merge(source); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := api.MergeResponse{
		Success: true,
		Message: "merged",
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.ResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	sketch, exists := s.sketches[req.Name]
	if !exists {
		writeError(w, http.StatusNotFound, "sketch not found")
		return
	}

	sketch.Reset()

	resp := api.ResetResponse{Success: true}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.StatsRequest
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			req.Name = ""
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var sketches []api.SketchStats

	if req.Name != "" {
		sketch, exists := s.sketches[req.Name]
		if !exists {
			writeError(w, http.StatusNotFound, "sketch not found")
			return
		}
		sketches = []api.SketchStats{{
			Name:       req.Name,
			Width:      sketch.Width(),
			Depth:      sketch.Depth(),
			TotalCount: sketch.TotalCount(),
		}}
	} else {
		sketches = make([]api.SketchStats, 0, len(s.sketches))
		for name, sketch := range s.sketches {
			sketches = append(sketches, api.SketchStats{
				Name:       name,
				Width:      sketch.Width(),
				Depth:      sketch.Depth(),
				TotalCount: sketch.TotalCount(),
			})
		}
	}

	resp := api.StatsResponse{
		Success:  true,
		Sketches: sketches,
	}
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	resp := api.ErrorResponse{
		Success: false,
		Error:   msg,
	}
	json.NewEncoder(w).Encode(resp)
}
