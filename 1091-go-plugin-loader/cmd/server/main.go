package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"plugin-system/pkg/common"
	"plugin-system/pkg/config"
	"plugin-system/pkg/plugin"
)

type Server struct {
	manager *plugin.Manager
}

func main() {
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	manager := plugin.NewManager(cfg.PluginDir)
	if err := manager.Scan(); err != nil {
		log.Printf("Warning: initial scan failed: %v", err)
	}

	srv := &Server{manager: manager}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/plugins/list", srv.handleList)
	mux.HandleFunc("/api/plugins/execute", srv.handleExecute)
	mux.HandleFunc("/api/plugins/scan", srv.handleScan)

	log.Printf("Server starting on %s, plugin directory: %s", cfg.Address, cfg.PluginDir)
	if err := http.ListenAndServe(cfg.Address, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	plugins := s.manager.List()
	resp := common.ListResponse{
		Plugins: make([]common.PluginInfo, 0, len(plugins)),
	}
	for _, p := range plugins {
		resp.Plugins = append(resp.Plugins, common.PluginInfo{
			Name:         p.Name,
			APIVersion:   p.APIVersion,
			Capabilities: p.Capabilities,
			Loaded:       true,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "plugin name is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	output, err := s.manager.Execute(ctx, req.Name, req.Input)
	if err != nil {
		if errors.Is(err, plugin.ErrPluginNotFound) || errors.Is(err, plugin.ErrPluginNotLoaded) {
			writeError(w, http.StatusNotFound, err.Error())
		} else if errors.Is(err, plugin.ErrVersionMismatch) {
			writeError(w, http.StatusPreconditionFailed, err.Error())
		} else if errors.Is(err, plugin.ErrTimeout) {
			writeError(w, http.StatusGatewayTimeout, err.Error())
		} else if errors.Is(err, plugin.ErrPanic) {
			writeError(w, http.StatusInternalServerError, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	resp := common.ExecuteResponse{
		Name:   req.Name,
		Output: output,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := s.manager.Scan(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := common.ScanResponse{Message: "scan completed"}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, common.ErrorResponse{
		Code:    status,
		Message: message,
	})
}
