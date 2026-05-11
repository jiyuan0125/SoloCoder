package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"config-hotload/pkg/api"
	"config-hotload/pkg/config"
)

type server struct {
	manager      *config.Manager
	watchClients map[chan api.WatchEvent]struct{}
	watchMu      sync.RWMutex
}

func newServer(manager *config.Manager) *server {
	s := &server{
		manager:      manager,
		watchClients: make(map[chan api.WatchEvent]struct{}),
	}

	manager.Subscribe("", s.handleConfigChange)
	return s
}

func (s *server) handleConfigChange(change config.Change) {
}

func (s *server) setupHistoryWatcher() {
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		lastCount := len(s.manager.History())
		for range ticker.C {
			history := s.manager.History()
			if len(history) > lastCount {
				for i := lastCount; i < len(history); i++ {
					event := api.WatchEvent{
						Timestamp: history[i].Timestamp.Format("2006-01-02T15:04:05.000Z07:00"),
						Changes:   api.ConvertChanges(history[i].Changes),
					}
					s.broadcastEvent(event)
				}
				lastCount = len(history)
			}
		}
	}()
}

func (s *server) broadcastEvent(event api.WatchEvent) {
	s.watchMu.RLock()
	clients := make([]chan api.WatchEvent, 0, len(s.watchClients))
	for ch := range s.watchClients {
		clients = append(clients, ch)
	}
	s.watchMu.RUnlock()

	for _, ch := range clients {
		select {
		case ch <- event:
		default:
		}
	}
}

func (s *server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cfg := s.manager.Get()
	sendJSON(w, api.GetConfigResponse{
		Success: true,
		Config:  cfg,
	})
}

func (s *server) handleGetField(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		sendError(w, http.StatusBadRequest, "missing 'path' query parameter")
		return
	}

	value, exists := s.manager.GetField(path)
	sendJSON(w, api.GetFieldResponse{
		Success: true,
		Value:   value,
		Exists:  exists,
	})
}

func (s *server) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	before := len(s.manager.History())
	if err := s.manager.Reload(); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	after := s.manager.History()
	var changes []config.Change
	if len(after) > before {
		changes = after[len(after)-1].Changes
	}

	sendJSON(w, api.ReloadResponse{
		Success: true,
		Changes: api.ConvertChanges(changes),
	})
}

func (s *server) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	history := s.manager.History()
	sendJSON(w, api.HistoryResponse{
		Success: true,
		History: api.ConvertHistory(history),
	})
}

func (s *server) handleWatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		sendError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan api.WatchEvent, 10)
	s.watchMu.Lock()
	s.watchClients[ch] = struct{}{}
	s.watchMu.Unlock()

	defer func() {
		s.watchMu.Lock()
		delete(s.watchClients, ch)
		s.watchMu.Unlock()
		close(ch)
	}()

	for {
		select {
		case event := <-ch:
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	var resp interface{}
	switch code {
	case http.StatusMethodNotAllowed:
		resp = api.GetConfigResponse{Success: false, Error: message}
	case http.StatusInternalServerError:
		resp = api.ReloadResponse{Success: false, Error: message}
	default:
		resp = map[string]interface{}{"success": false, "error": message}
	}
	json.NewEncoder(w).Encode(resp)
}

func detectFormat(filePath string) config.ConfigFormat {
	lower := strings.ToLower(filePath)
	if strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml") {
		return config.FormatYAML
	}
	return config.FormatJSON
}

func main() {
	port := flag.String("port", getEnvOrDefault("CONFIG_SERVER_PORT", "8080"), "server port")
	configPath := flag.String("config", "", "path to config file")
	enableRef := flag.Bool("enable-ref", false, "enable variable reference resolution")
	flag.Parse()

	if *configPath == "" {
		log.Fatal("config file path is required (use -config flag)")
	}

	manager, err := config.NewManager(config.Options{
		FilePath:  *configPath,
		Format:    detectFormat(*configPath),
		EnableRef: *enableRef,
	})
	if err != nil {
		log.Fatalf("failed to create config manager: %v", err)
	}

	if err := manager.Start(); err != nil {
		log.Fatalf("failed to start config watcher: %v", err)
	}
	defer manager.Stop()

	server := newServer(manager)
	server.setupHistoryWatcher()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/config", server.handleGetConfig)
	mux.HandleFunc("/api/config/field", server.handleGetField)
	mux.HandleFunc("/api/config/reload", server.handleReload)
	mux.HandleFunc("/api/config/history", server.handleHistory)
	mux.HandleFunc("/api/config/watch", server.handleWatch)

	addr := ":" + *port
	log.Printf("starting config server on %s", addr)
	log.Printf("watching config file: %s", *configPath)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
