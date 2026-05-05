// Package main provides the config server HTTP API.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"config-loader/common"
	"config-loader/config"
)

const (
	version = "1.0.0"
)

type Server struct {
	config       *config.Config
	loader       *config.Loader
	startTime    time.Time
	configFiles  []string
	mu           sync.RWMutex
	watchEnabled bool
}

func NewServer(configFiles []string, watchEnabled bool) (*Server, error) {
	s := &Server{
		startTime:    time.Now(),
		configFiles:  configFiles,
		watchEnabled: watchEnabled,
	}

	loader := config.NewLoader()
	for _, f := range configFiles {
		loader.AddFile(f)
	}

	cfg, err := loader.LoadToConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	s.config = cfg
	s.loader = loader

	if watchEnabled {
		loader.EnableWatch()
		loader.OnChange(func(oldCfg, newCfg *config.Config, err error) {
			if err != nil {
				fmt.Printf("Warning: config reload failed: %v\n", err)
				return
			}
			s.mu.Lock()
			s.config = newCfg
			s.mu.Unlock()
			fmt.Println("Configuration reloaded successfully")
		})
	}

	return s, nil
}

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	configData := s.config.All()
	s.mu.RUnlock()

	resp := common.HealthResponse{
		Status:      "healthy",
		Uptime:      time.Since(s.startTime).String(),
		Version:     version,
		ConfigCount: countConfigItems(configData),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func countConfigItems(m map[string]interface{}) int {
	count := 0
	for _, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			count += countConfigItems(val)
		default:
			count++
		}
	}
	return count
}

func (s *Server) ConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		configData := s.config.All()
		s.mu.RUnlock()

		resp := common.NewSuccessResponse(configData)
		json.NewEncoder(w).Encode(resp)

	case http.MethodPost:
		var req common.ConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			resp := common.NewErrorResponse(common.ErrCodeInvalidPath, "invalid request body")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
			return
		}

		if req.Path == "" {
			resp := common.NewErrorResponse(common.ErrCodeInvalidPath, "path is required")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
			return
		}

		s.mu.Lock()
		s.config.Set(req.Path, req.Value, config.SourceCLI)
		s.mu.Unlock()

		resp := common.NewSuccessResponse(map[string]interface{}{
			"path":  req.Path,
			"value": req.Value,
		})
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) ConfigByPathHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, common.EndpointConfigByPath)
	if path == "" {
		resp := common.NewErrorResponse(common.ErrCodeInvalidPath, "path is required")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		value := s.config.Get(path)
		s.mu.RUnlock()

		if value == nil {
			resp := common.NewErrorResponse(common.ErrCodeNotFound, fmt.Sprintf("config not found: %s", path))
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(resp)
			return
		}

		resp := common.NewSuccessResponse(value)
		json.NewEncoder(w).Encode(resp)

	case http.MethodDelete:
		s.mu.RLock()
		existing := s.config.Get(path)
		s.mu.RUnlock()

		if existing == nil {
			resp := common.NewErrorResponse(common.ErrCodeNotFound, fmt.Sprintf("config not found: %s", path))
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(resp)
			return
		}

		resp := common.NewSuccessResponse(map[string]string{
			"status": "deleted",
			"path":   path,
		})
		json.NewEncoder(w).Encode(resp)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) ReloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	loader := config.NewLoader()
	for _, f := range s.configFiles {
		loader.AddFile(f)
	}

	newConfig, err := loader.LoadToConfig()
	if err != nil {
		resp := common.NewErrorResponse(common.ErrCodeInternal, fmt.Sprintf("failed to reload config: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	s.mu.Lock()
	s.config = newConfig
	s.mu.Unlock()

	resp := common.NewSuccessResponse(map[string]string{
		"status": "reloaded",
		"time":   time.Now().Format(time.RFC3339),
	})
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) WatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	resp := common.NewSuccessResponse(map[string]interface{}{
		"watch_enabled": s.watchEnabled,
		"config_files":  s.configFiles,
	})
	json.NewEncoder(w).Encode(resp)
}

func main() {
	port := flag.String("port", "8080", "HTTP server port")
	configDir := flag.String("config-dir", ".", "Directory to search for config files")
	configFiles := flag.String("config", "", "Comma-separated list of config files")
	watch := flag.Bool("watch", false, "Enable config file watching")
	flag.Parse()

	var files []string
	if *configFiles != "" {
		files = strings.Split(*configFiles, ",")
		for i, f := range files {
			files[i] = strings.TrimSpace(f)
		}
	} else {
		candidates := []string{
			"config.yaml",
			"config.yml",
			"config.json",
			"config.local.yaml",
			"config.local.yml",
			"config.local.json",
		}
		for _, c := range candidates {
			path := filepath.Join(*configDir, c)
			if _, err := os.Stat(path); err == nil {
				files = append(files, path)
			}
		}
	}

	if len(files) == 0 {
		fmt.Println("Warning: No config files found. Using only environment variables and CLI arguments.")
	} else {
		fmt.Printf("Loading config from: %v\n", files)
	}

	server, err := NewServer(files, *watch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating server: %v\n", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	mux.HandleFunc(common.EndpointHealth, server.HealthHandler)
	mux.HandleFunc(common.EndpointConfig, server.ConfigHandler)
	mux.HandleFunc(common.EndpointConfigByPath, server.ConfigByPathHandler)
	mux.HandleFunc(common.EndpointReload, server.ReloadHandler)
	mux.HandleFunc(common.EndpointWatch, server.WatchHandler)

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Config server starting on %s\n", addr)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  GET  %s - Health check\n", common.EndpointHealth)
	fmt.Printf("  GET  %s - Get all config\n", common.EndpointConfig)
	fmt.Printf("  POST %s - Set config value\n", common.EndpointConfig)
	fmt.Printf("  GET  %s{path} - Get config by path\n", common.EndpointConfigByPath)
	fmt.Printf("  POST %s - Reload config from files\n", common.EndpointReload)
	fmt.Printf("  GET  %s - Watch status\n", common.EndpointWatch)

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
