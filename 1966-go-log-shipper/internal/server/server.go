package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"log-shipper/internal/dispatcher"
	"log-shipper/internal/targets"
	"log-shipper/internal/types"
)

type Server struct {
	dispatcher *dispatcher.Dispatcher
	httpServer *http.Server

	startTime    time.Time
	targetSeq    int64
}

func New(d *dispatcher.Dispatcher) *Server {
	return &Server{
		dispatcher: d,
		startTime:  time.Now(),
	}
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/ingest", s.handleIngest)
	mux.HandleFunc("/targets", s.handleTargets)
	mux.HandleFunc("/targets/", s.handleTargetByID)
	mux.HandleFunc("/stats", s.handleStats)
	mux.HandleFunc("/health", s.handleHealth)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("log-shipper starting on %s", addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var entries []*types.LogEntry
	if strings.HasPrefix(strings.TrimSpace(string(body)), "[") {
		if err := json.Unmarshal(body, &entries); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		var single *types.LogEntry
		if err := json.Unmarshal(body, &single); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		entries = []*types.LogEntry{single}
	}

	for _, entry := range entries {
		if entry.Timestamp.IsZero() {
			entry.Timestamp = time.Now()
		}
		if entry.Level == "" {
			entry.Level = "INFO"
		}
		if entry.Source == "" {
			entry.Source = "ingest"
		}
	}

	s.dispatcher.Dispatch(entries)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"count":   len(entries),
	})
}

func (s *Server) handleTargets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		list := s.dispatcher.ListTargets()
		if list == nil {
			list = []types.TargetStatus{}
		}
		writeJSON(w, http.StatusOK, list)

	case http.MethodPost:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var req struct {
			Type   string          `json:"type"`
			ID     string          `json:"id"`
			Config json.RawMessage `json:"config"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		if req.Type == "" {
			http.Error(w, "missing required field: type", http.StatusBadRequest)
			return
		}

		id := req.ID
		if id == "" {
			id = req.Type + "-" + s.nextID()
		}

		var target types.Target
		switch req.Type {
		case "file":
			var cfg types.FileTargetConfig
			if err := json.Unmarshal(req.Config, &cfg); err != nil {
				http.Error(w, "invalid file config: "+err.Error(), http.StatusBadRequest)
				return
			}
			if cfg.Path == "" {
				http.Error(w, "file config missing: path", http.StatusBadRequest)
				return
			}
			ft, err := targets.NewFileTarget(id, cfg)
			if err != nil {
				http.Error(w, "failed to create file target: "+err.Error(), http.StatusInternalServerError)
				return
			}
			target = ft

		case "http":
			var cfg types.HTTPTargetConfig
			if err := json.Unmarshal(req.Config, &cfg); err != nil {
				http.Error(w, "invalid http config: "+err.Error(), http.StatusBadRequest)
				return
			}
			if cfg.URL == "" {
				http.Error(w, "http config missing: url", http.StatusBadRequest)
				return
			}
			target = targets.NewHTTPTarget(id, cfg)

		case "memory":
			var cfg types.MemoryTargetConfig
			if err := json.Unmarshal(req.Config, &cfg); err != nil {
				http.Error(w, "invalid memory config: "+err.Error(), http.StatusBadRequest)
				return
			}
			target = targets.NewMemoryTarget(id, cfg)

		default:
			http.Error(w, "unknown target type: "+req.Type, http.StatusBadRequest)
			return
		}

		if err := s.dispatcher.AddTarget(target); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"status": "ok",
			"id":     id,
			"type":   req.Type,
		})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTargetByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/targets/")
	id := strings.TrimSpace(path)

	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if id == "" {
		http.Error(w, "missing target id", http.StatusBadRequest)
		return
	}

	if removed := s.dispatcher.RemoveTarget(id); !removed {
		http.Error(w, "target not found: "+id, http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"id":     id,
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := s.dispatcher.Stats()
	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"uptime":    time.Since(s.startTime).String(),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) nextID() string {
	seq := atomic.AddInt64(&s.targetSeq, 1)
	return fmt.Sprintf("%s-%06d", time.Now().Format("20060102"), seq)
}
