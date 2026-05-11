package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"go-cross-build/pkg/api"
	"go-cross-build/pkg/core"
)

type Server struct {
	configs map[string]*api.BuildConfig
	mu      sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		configs: make(map[string]*api.BuildConfig),
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.BuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := core.ValidateConfig(&req.Config); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	builder := core.NewBuilder(nil)

	resp, err := builder.BuildAll(req.Config)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleScript(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.ScriptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := core.ValidateConfig(&req.Config); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	content, err := core.GenerateScript(req.Config, req.Format)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, &api.ScriptResponse{
		Content: content,
		Format:  req.Format,
	})
}

func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.ValidatePlatformRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp := &api.ValidatePlatformResponse{
		Results: make(map[string]bool),
	}

	for _, p := range req.Platforms {
		key := p.GOOS + "/" + p.GOARCH
		valid := core.IsValidCombination(p.GOOS, p.GOARCH)
		resp.Results[key] = valid

		if !valid {
			if !core.IsValidGOOS(p.GOOS) {
				resp.Errors = append(resp.Errors, fmt.Sprintf("invalid GOOS: %s", p.GOOS))
			}
			if !core.IsValidGOARCH(p.GOARCH) {
				resp.Errors = append(resp.Errors, fmt.Sprintf("invalid GOARCH: %s", p.GOARCH))
			}
			if core.IsValidGOOS(p.GOOS) && core.IsValidGOARCH(p.GOARCH) {
				resp.Errors = append(resp.Errors, fmt.Sprintf("invalid combination: %s/%s", p.GOOS, p.GOARCH))
			}
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleCommands(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.BuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := core.ValidateConfig(&req.Config); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	results := make([]*api.CommandResult, 0, len(req.Config.Targets))
	for _, target := range req.Config.Targets {
		results = append(results, core.GenerateBuildCommand(target, req.Config))
	}

	writeJSON(w, http.StatusOK, results)
}

func (s *Server) handleConfigCreate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var config api.BuildConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := core.ValidateConfig(&config); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		if config.Name == "" {
			writeError(w, http.StatusBadRequest, "config name is required")
			return
		}

		s.mu.Lock()
		s.configs[config.Name] = &config
		s.mu.Unlock()

		writeJSON(w, http.StatusCreated, map[string]string{
			"status":  "created",
			"name":    config.Name,
		})

	case http.MethodGet:
		s.mu.RLock()
		configs := make([]*api.BuildConfig, 0, len(s.configs))
		for _, c := range s.configs {
			configs = append(configs, c)
		}
		s.mu.RUnlock()

		writeJSON(w, http.StatusOK, configs)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleConfigByName(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/api/configs/"):]
	if name == "" {
		writeError(w, http.StatusBadRequest, "config name is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		config, exists := s.configs[name]
		s.mu.RUnlock()

		if !exists {
			writeError(w, http.StatusNotFound, "config not found")
			return
		}

		writeJSON(w, http.StatusOK, config)

	case http.MethodDelete:
		s.mu.Lock()
		_, exists := s.configs[name]
		if exists {
			delete(s.configs, name)
		}
		s.mu.Unlock()

		if !exists {
			writeError(w, http.StatusNotFound, "config not found")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"status": "deleted",
			"name":   name,
		})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func getPort() string {
	var port int
	flag.IntVar(&port, "port", 0, "server port")
	flag.Parse()

	if port != 0 {
		return fmt.Sprintf(":%d", port)
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		return ":" + envPort
	}

	return ":8080"
}

func main() {
	server := NewServer()

	http.HandleFunc("/health", server.handleHealth)
	http.HandleFunc("/api/build", server.handleBuild)
	http.HandleFunc("/api/script", server.handleScript)
	http.HandleFunc("/api/validate", server.handleValidate)
	http.HandleFunc("/api/commands", server.handleCommands)
	http.HandleFunc("/api/configs", server.handleConfigCreate)
	http.HandleFunc("/api/configs/", server.handleConfigByName)

	addr := getPort()
	log.Printf("server starting on %s", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
