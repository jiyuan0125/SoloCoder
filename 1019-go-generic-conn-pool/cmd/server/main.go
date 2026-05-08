package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"generic-conn-pool/internal/pool"
	"generic-conn-pool/pkg/api"
)

type serverState struct {
	pools       map[string]*pool.Pool
	connToPool  map[string]string
	activeConns map[string]net.Conn
	mu          sync.Mutex
	poolCounter uint64
}

func newServerState() *serverState {
	return &serverState{
		pools:       make(map[string]*pool.Pool),
		connToPool:  make(map[string]string),
		activeConns: make(map[string]net.Conn),
	}
}

func (s *serverState) generatePoolID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.poolCounter++
	return fmt.Sprintf("pool-%d-%s", s.poolCounter, time.Now().Format("20060102150405"))
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func handleCreate(s *serverState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, api.CreatePoolResponse{
				Success: false,
				Message: "method not allowed",
			})
			return
		}

		var req api.CreatePoolRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.CreatePoolResponse{
				Success: false,
				Message: "invalid request body",
			})
			return
		}

		cfg := pool.DefaultConfig()
		cfg.TargetAddress = req.TargetAddress
		if req.MinIdle > 0 {
			cfg.MinIdle = req.MinIdle
		}
		if req.MaxActive > 0 {
			cfg.MaxActive = req.MaxActive
		}
		if req.IdleTimeoutSec > 0 {
			cfg.IdleTimeout = time.Duration(req.IdleTimeoutSec) * time.Second
		}
		if req.MaxLifetimeSec > 0 {
			cfg.MaxLifetime = time.Duration(req.MaxLifetimeSec) * time.Second
		}
		if req.AcquireTimeoutMs > 0 {
			cfg.AcquireTimeout = time.Duration(req.AcquireTimeoutMs) * time.Millisecond
		}
		if req.HealthCheckMs > 0 {
			cfg.HealthCheckTime = time.Duration(req.HealthCheckMs) * time.Millisecond
		}

		if cfg.MinIdle > cfg.MaxActive {
			writeJSON(w, http.StatusBadRequest, api.CreatePoolResponse{
				Success: false,
				Message: "min idle cannot be greater than max active",
			})
			return
		}

		p, err := pool.NewPool(cfg)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, api.CreatePoolResponse{
				Success: false,
				Message: fmt.Sprintf("failed to create pool: %v", err),
			})
			return
		}

		poolID := s.generatePoolID()
		s.mu.Lock()
		s.pools[poolID] = p
		s.mu.Unlock()

		writeJSON(w, http.StatusOK, api.CreatePoolResponse{
			Success: true,
			Message: poolID,
		})
	}
}

func handleAcquire(s *serverState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, api.AcquireConnectionResponse{
				Success: false,
				Message: "method not allowed",
			})
			return
		}

		poolID := r.URL.Query().Get("pool_id")
		if poolID == "" {
			writeJSON(w, http.StatusBadRequest, api.AcquireConnectionResponse{
				Success: false,
				Message: "pool_id is required",
			})
			return
		}

		s.mu.Lock()
		p, ok := s.pools[poolID]
		s.mu.Unlock()

		if !ok {
			writeJSON(w, http.StatusNotFound, api.AcquireConnectionResponse{
				Success: false,
				Message: "pool not found",
			})
			return
		}

		connID, conn, err := p.Acquire()
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, api.AcquireConnectionResponse{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		s.mu.Lock()
		s.connToPool[connID] = poolID
		s.activeConns[connID] = conn
		s.mu.Unlock()

		writeJSON(w, http.StatusOK, api.AcquireConnectionResponse{
			Success: true,
			ConnID:  connID,
		})
	}
}

func handleRelease(s *serverState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, api.ReleaseConnectionResponse{
				Success: false,
				Message: "method not allowed",
			})
			return
		}

		var req api.ReleaseConnectionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.ReleaseConnectionResponse{
				Success: false,
				Message: "invalid request body",
			})
			return
		}

		if req.ConnID == "" {
			writeJSON(w, http.StatusBadRequest, api.ReleaseConnectionResponse{
				Success: false,
				Message: "conn_id is required",
			})
			return
		}

		s.mu.Lock()
		poolID, ok := s.connToPool[req.ConnID]
		if !ok {
			s.mu.Unlock()
			writeJSON(w, http.StatusNotFound, api.ReleaseConnectionResponse{
				Success: false,
				Message: "connection not found",
			})
			return
		}

		p, ok := s.pools[poolID]
		if !ok {
			delete(s.connToPool, req.ConnID)
			delete(s.activeConns, req.ConnID)
			s.mu.Unlock()
			writeJSON(w, http.StatusNotFound, api.ReleaseConnectionResponse{
				Success: false,
				Message: "pool not found",
			})
			return
		}

		delete(s.connToPool, req.ConnID)
		delete(s.activeConns, req.ConnID)
		s.mu.Unlock()

		if err := p.Release(req.ConnID); err != nil {
			writeJSON(w, http.StatusInternalServerError, api.ReleaseConnectionResponse{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, api.ReleaseConnectionResponse{
			Success: true,
		})
	}
}

func handleStats(s *serverState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, api.PoolStatsResponse{
				Success: false,
				Message: "method not allowed",
			})
			return
		}

		poolID := r.URL.Query().Get("pool_id")
		if poolID == "" {
			writeJSON(w, http.StatusBadRequest, api.PoolStatsResponse{
				Success: false,
				Message: "pool_id is required",
			})
			return
		}

		s.mu.Lock()
		p, ok := s.pools[poolID]
		s.mu.Unlock()

		if !ok {
			writeJSON(w, http.StatusNotFound, api.PoolStatsResponse{
				Success: false,
				Message: "pool not found",
			})
			return
		}

		minIdle, maxActive, totalConns, idleConns, activeConns := p.Stats()

		writeJSON(w, http.StatusOK, api.PoolStatsResponse{
			Success:     true,
			MinIdle:     minIdle,
			MaxActive:   maxActive,
			TotalConns:  totalConns,
			IdleConns:   idleConns,
			ActiveConns: activeConns,
		})
	}
}

func main() {
	state := newServerState()

	mux := http.NewServeMux()
	mux.HandleFunc("/pool/create", handleCreate(state))
	mux.HandleFunc("/pool/acquire", handleAcquire(state))
	mux.HandleFunc("/pool/release", handleRelease(state))
	mux.HandleFunc("/pool/stats", handleStats(state))

	addr := ":8080"
	fmt.Printf("Server starting on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
