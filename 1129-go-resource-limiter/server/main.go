package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"resource-limiter/common"
	"resource-limiter/limiter"
)

type Server struct {
	limiter *limiter.Limiter
}

func NewServer() *Server {
	config := limiter.DefaultConfig()
	return &Server{
		limiter: limiter.NewLimiter(config),
	}
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.RequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	acquired, release := s.limiter.Acquire(ctx, req.Key)
	if !acquired {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(common.RequestResponse{
			Success: false,
			Message: "Rate limit exceeded",
		})
		return
	}
	defer release()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.RequestResponse{
		Success: true,
	})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		config := s.limiter.Config()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(common.ConfigResponse{
			GlobalLimit: config.GlobalLimit,
			KeyLimit:    config.KeyLimit,
			Mode:        common.LimitMode(config.Mode),
		})

	case http.MethodPost:
		var req common.ConfigUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.GlobalLimit != nil {
			s.limiter.SetGlobalLimit(*req.GlobalLimit)
		}

		if req.KeyLimit != nil {
			if req.Key != "" {
				s.limiter.SetKeyLimit(req.Key, *req.KeyLimit)
			} else {
				s.limiter.SetDefaultKeyLimit(*req.KeyLimit)
			}
		}

		if req.Mode != nil {
			s.limiter.SetMode(limiter.LimitMode(*req.Mode))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleMode(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		config := s.limiter.Config()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mode": common.LimitMode(config.Mode),
		})

	case http.MethodPost:
		var req common.ModeUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		s.limiter.SetMode(limiter.LimitMode(req.Mode))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mode": req.Mode,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := s.limiter.Stats()
	
	keyStats := make(map[string]common.KeyStats)
	for k, v := range stats.KeyStats {
		keyStats[k] = common.KeyStats{
			Key:       v.Key,
			InUse:     v.InUse,
			Capacity:  v.Capacity,
			Available: v.Available,
			Rejected:  v.Rejected,
		}
	}

	response := common.StatsResponse{
		GlobalInUse:     stats.GlobalInUse,
		GlobalCapacity:  stats.GlobalCapacity,
		GlobalAvailable: stats.GlobalAvailable,
		GlobalQueued:    stats.GlobalQueued,
		GlobalRejected:  stats.GlobalRejected,
		Mode:            common.LimitMode(stats.Mode),
		KeyStats:        keyStats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	addr := flag.String("addr", ":8080", "Server address")
	globalLimit := flag.Int("global-limit", 100, "Global concurrency limit")
	keyLimit := flag.Int("key-limit", 10, "Default per-key concurrency limit")
	flag.Parse()

	config := limiter.DefaultConfig()
	config.GlobalLimit = *globalLimit
	config.KeyLimit = *keyLimit

	server := &Server{
		limiter: limiter.NewLimiter(config),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/request", server.handleRequest)
	mux.HandleFunc("/config", server.handleConfig)
	mux.HandleFunc("/mode", server.handleMode)
	mux.HandleFunc("/stats", server.handleStats)

	log.Printf("Server starting on %s (global: %d, key: %d)", *addr, *globalLimit, *keyLimit)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
