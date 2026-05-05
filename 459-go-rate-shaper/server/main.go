// Package main implements an HTTP server for the rate limiter service.
// It exposes REST API endpoints for rate limiting operations.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"rate-shaper/protocol"
	"rate-shaper/ratelimiter"
)

type RateLimiterServer struct {
	limiter   ratelimiter.Limiter
	whitelist *ratelimiter.WhitelistLimiter
	mu        sync.RWMutex
}

func NewRateLimiterServer(config ratelimiter.Config) (*RateLimiterServer, error) {
	limiter, err := ratelimiter.NewLimiter(config)
	if err != nil {
		return nil, err
	}

	whitelist := ratelimiter.NewWhitelistLimiter(limiter)

	return &RateLimiterServer{
		limiter:   limiter,
		whitelist: whitelist,
	}, nil
}

func (s *RateLimiterServer) AllowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.AllowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	if req.Count <= 0 {
		req.Count = 1
	}

	var result ratelimiter.Result
	s.mu.RLock()
	result = s.whitelist.AllowN(req.Key, req.Count)
	s.mu.RUnlock()

	resetAt := ""
	if !result.ResetAt.IsZero() {
		resetAt = result.ResetAt.Format(time.RFC3339)
	}

	resp := protocol.AllowResponse{
		Allowed:    result.Allowed,
		Limited:    result.Limited,
		Remaining:  result.Remaining,
		WaitTimeMs: result.WaitTime.Milliseconds(),
		ResetAt:    resetAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(s.whitelist.GetConfig().Limit, 10))
	w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
	if !result.ResetAt.IsZero() {
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))
	}
	if result.Limited {
		w.Header().Set("Retry-After", strconv.FormatInt(result.WaitTime.Milliseconds()/1000, 10))
		w.WriteHeader(http.StatusTooManyRequests)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *RateLimiterServer) StatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key query parameter is required", http.StatusBadRequest)
		return
	}

	var stats ratelimiter.Stats
	s.mu.RLock()
	stats = s.whitelist.Stats(key)
	s.mu.RUnlock()

	resetAt := ""
	if !stats.ResetAt.IsZero() {
		resetAt = stats.ResetAt.Format(time.RFC3339)
	}

	resp := protocol.StatsResponse{
		Key:       stats.Key,
		Mode:      ratelimiter.ModeToString(stats.Mode),
		Used:      stats.Used,
		Limit:     stats.Limit,
		Remaining: stats.Remaining,
		ResetAt:   resetAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *RateLimiterServer) ConfigHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getConfigHandler(w, r)
	case http.MethodPut:
		s.putConfigHandler(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *RateLimiterServer) getConfigHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	config := s.whitelist.GetConfig()
	s.mu.RUnlock()

	resp := map[string]interface{}{
		"mode":         ratelimiter.ModeToString(config.Mode),
		"limit":        config.Limit,
		"window_ms":    config.Window.Milliseconds(),
		"bucket_count": config.BucketCount,
		"burst":        config.Burst,
		"rate":         config.Rate,
		"capacity":     config.Capacity,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *RateLimiterServer) putConfigHandler(w http.ResponseWriter, r *http.Request) {
	var req protocol.ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var config ratelimiter.Config
	s.mu.RLock()
	config = s.whitelist.GetConfig()
	s.mu.RUnlock()

	if req.Mode != "" {
		config.Mode = ratelimiter.StringToMode(req.Mode)
	}
	if req.Limit > 0 {
		config.Limit = req.Limit
	}
	if req.WindowMs > 0 {
		config.Window = time.Duration(req.WindowMs) * time.Millisecond
	}
	if req.BucketCount >= 2 {
		config.BucketCount = req.BucketCount
	}
	if req.Burst > 0 {
		config.Burst = req.Burst
	}
	if req.Rate > 0 {
		config.Rate = req.Rate
	}
	if req.Capacity > 0 {
		config.Capacity = req.Capacity
	}

	currentMode := s.whitelist.GetConfig().Mode
	if req.Mode != "" && config.Mode != currentMode {
		newLimiter, err := ratelimiter.NewLimiter(config)
		if err != nil {
			resp := protocol.ConfigResponse{
				Success: false,
				Error:   err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
			return
		}

		s.mu.Lock()
		s.limiter = newLimiter
		s.whitelist = ratelimiter.NewWhitelistLimiter(newLimiter)
		s.mu.Unlock()
	} else {
		s.mu.Lock()
		err := s.whitelist.UpdateConfig(config)
		s.mu.Unlock()

		if err != nil {
			resp := protocol.ConfigResponse{
				Success: false,
				Error:   err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
			return
		}
	}

	resp := protocol.ConfigResponse{
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *RateLimiterServer) WhitelistHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getWhitelistHandler(w, r)
	case http.MethodPost:
		s.addWhitelistHandler(w, r)
	case http.MethodDelete:
		s.removeWhitelistHandler(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *RateLimiterServer) getWhitelistHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	keys := s.whitelist.GetWhitelist()
	s.mu.RUnlock()

	resp := protocol.WhitelistResponse{
		Success: true,
		Keys:    keys,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *RateLimiterServer) addWhitelistHandler(w http.ResponseWriter, r *http.Request) {
	var req protocol.WhitelistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.whitelist.AddToWhitelist(req.Key)
	keys := s.whitelist.GetWhitelist()
	s.mu.Unlock()

	resp := protocol.WhitelistResponse{
		Success: true,
		Keys:    keys,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *RateLimiterServer) removeWhitelistHandler(w http.ResponseWriter, r *http.Request) {
	var req protocol.WhitelistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.whitelist.RemoveFromWhitelist(req.Key)
	keys := s.whitelist.GetWhitelist()
	s.mu.Unlock()

	resp := protocol.WhitelistResponse{
		Success: true,
		Keys:    keys,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *RateLimiterServer) ResetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.ResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	if req.Key == "" {
		s.whitelist.ResetAll()
	} else {
		s.whitelist.Reset(req.Key)
	}
	s.mu.Unlock()

	resp := protocol.ResetResponse{
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func main() {
	config := ratelimiter.DefaultConfig()
	config.Mode = ratelimiter.ModeFixedWindow
	config.Limit = 10
	config.Window = time.Second

	server, err := NewRateLimiterServer(config)
	if err != nil {
		fmt.Printf("Failed to create rate limiter server: %v\n", err)
		return
	}

	http.HandleFunc("/allow", server.AllowHandler)
	http.HandleFunc("/stats", server.StatsHandler)
	http.HandleFunc("/config", server.ConfigHandler)
	http.HandleFunc("/whitelist", server.WhitelistHandler)
	http.HandleFunc("/reset", server.ResetHandler)

	port := ":8080"
	fmt.Printf("Rate limiter server starting on %s\n", port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  POST /allow       - Check if request is allowed\n")
	fmt.Printf("  GET  /stats?key=  - Get rate limiter statistics\n")
	fmt.Printf("  GET  /config      - Get current configuration\n")
	fmt.Printf("  PUT  /config      - Update configuration\n")
	fmt.Printf("  GET  /whitelist   - List whitelisted keys\n")
	fmt.Printf("  POST /whitelist   - Add key to whitelist\n")
	fmt.Printf("  DELETE /whitelist - Remove key from whitelist\n")
	fmt.Printf("  POST /reset       - Reset rate limiter\n")

	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
