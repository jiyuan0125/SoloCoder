package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Dimension string

const (
	DimensionIP     Dimension = "ip"
	DimensionAPIKey Dimension = "apikey"
)

type Rule struct {
	ID           string     `json:"id"`
	Dimension    Dimension  `json:"dimension"`
	WindowSeconds int       `json:"window_seconds"`
	MaxRequests  int        `json:"max_requests"`
	CreatedAt    time.Time  `json:"created_at"`
}

type Counter struct {
	Timestamp time.Time
	Count     int
}

type SlidingWindow struct {
	buckets []Counter
	window  time.Duration
	mu      sync.RWMutex
}

type CoolDownEntry struct {
	Until time.Time
	RuleIDs map[string]struct{}
}

type Service struct {
	rules      map[string]*Rule
	rulesMu    sync.RWMutex
	apiKeys    map[string]struct{}
	revokedKeys map[string]struct{}
	keysMu     sync.RWMutex
	ipCounters map[string]*SlidingWindow
	keyCounters map[string]*SlidingWindow
	countersMu sync.RWMutex
	coolDowns  map[string]*CoolDownEntry
	coolDownMu sync.RWMutex
	ruleCounter int
	cleanupTicker *time.Ticker
}

func newSlidingWindow(windowSeconds int) *SlidingWindow {
	return &SlidingWindow{
		buckets: make([]Counter, 0),
		window:  time.Duration(windowSeconds) * time.Second,
	}
}

func (sw *SlidingWindow) add(now time.Time) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	cutoff := now.Add(-sw.window)
	if len(sw.buckets) > 0 && sw.buckets[len(sw.buckets)-1].Timestamp.Equal(now) {
		sw.buckets[len(sw.buckets)-1].Count++
	} else {
		sw.buckets = append(sw.buckets, Counter{Timestamp: now, Count: 1})
	}
	sw.cleanup(cutoff)
}

func (sw *SlidingWindow) count(now time.Time) int {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	cutoff := now.Add(-sw.window)
	total := 0
	for i := 0; i < len(sw.buckets); i++ {
		if sw.buckets[i].Timestamp.After(cutoff) || sw.buckets[i].Timestamp.Equal(cutoff) {
			total += sw.buckets[i].Count
		}
	}
	return total
}

func (sw *SlidingWindow) cleanup(cutoff time.Time) {
	newBuckets := sw.buckets[:0]
	for _, b := range sw.buckets {
		if b.Timestamp.After(cutoff) || b.Timestamp.Equal(cutoff) {
			newBuckets = append(newBuckets, b)
		}
	}
	sw.buckets = newBuckets
}

func (sw *SlidingWindow) isEmpty() bool {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return len(sw.buckets) == 0
}

func NewService() *Service {
	s := &Service{
		rules:       make(map[string]*Rule),
		apiKeys:     make(map[string]struct{}),
		revokedKeys: make(map[string]struct{}),
		ipCounters:  make(map[string]*SlidingWindow),
		keyCounters: make(map[string]*SlidingWindow),
		coolDowns:   make(map[string]*CoolDownEntry),
	}
	s.cleanupTicker = time.NewTicker(time.Minute)
	go s.cleanupLoop()
	return s
}

func (s *Service) cleanupLoop() {
	for range s.cleanupTicker.C {
		s.Cleanup()
	}
}

func (s *Service) Cleanup() {
	now := time.Now()
	s.countersMu.Lock()
	for key, sw := range s.ipCounters {
		cutoff := now.Add(-sw.window)
		sw.mu.Lock()
		sw.cleanup(cutoff)
		if len(sw.buckets) == 0 {
			delete(s.ipCounters, key)
		}
		sw.mu.Unlock()
	}
	for key, sw := range s.keyCounters {
		cutoff := now.Add(-sw.window)
		sw.mu.Lock()
		sw.cleanup(cutoff)
		if len(sw.buckets) == 0 {
			delete(s.keyCounters, key)
		}
		sw.mu.Unlock()
	}
	s.countersMu.Unlock()

	s.coolDownMu.Lock()
	for key, entry := range s.coolDowns {
		if now.After(entry.Until) {
			delete(s.coolDowns, key)
		}
	}
	s.coolDownMu.Unlock()
}

func generateKey() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *Service) getClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		parts := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(parts[0])
	}
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (s *Service) getAPIKey(r *http.Request) string {
	return r.Header.Get("X-API-Key")
}

func (s *Service) getOrCreateWindow(dimension Dimension, key string, windowSeconds int) *SlidingWindow {
	var counters map[string]*SlidingWindow
	if dimension == DimensionIP {
		counters = s.ipCounters
	} else {
		counters = s.keyCounters
	}

	s.countersMu.RLock()
	sw, exists := counters[key]
	if exists {
		s.countersMu.RUnlock()
		return sw
	}
	s.countersMu.RUnlock()

	s.countersMu.Lock()
	defer s.countersMu.Unlock()
	sw, exists = counters[key]
	if exists {
		return sw
	}
	sw = newSlidingWindow(windowSeconds)
	counters[key] = sw
	return sw
}

func (s *Service) checkAndIncrement(dimension Dimension, key string, rule *Rule, now time.Time) (limited bool, exceeded bool) {
	sw := s.getOrCreateWindow(dimension, key, rule.WindowSeconds)
	current := sw.count(now)
	if current >= rule.MaxRequests {
		return true, true
	}
	sw.add(now)
	return false, false
}

func (s *Service) clearCoolDownsForRule(ruleID string) {
	s.coolDownMu.Lock()
	defer s.coolDownMu.Unlock()
	for key, entry := range s.coolDowns {
		delete(entry.RuleIDs, ruleID)
		if len(entry.RuleIDs) == 0 {
			delete(s.coolDowns, key)
		}
	}
}

func (s *Service) clearAllCoolDowns() {
	s.coolDownMu.Lock()
	defer s.coolDownMu.Unlock()
	s.coolDowns = make(map[string]*CoolDownEntry)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	type ErrorResponse struct {
		Error string `json:"error"`
	}
	writeJSON(w, status, ErrorResponse{Error: message})
}

func (s *Service) CreateRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	type CreateRuleRequest struct {
		Dimension     string `json:"dimension"`
		WindowSeconds int    `json:"window_seconds"`
		MaxRequests   int    `json:"max_requests"`
	}

	var req CreateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	dim := Dimension(strings.ToLower(req.Dimension))
	if dim != DimensionIP && dim != DimensionAPIKey {
		writeError(w, http.StatusBadRequest, "Invalid dimension, must be 'ip' or 'apikey'")
		return
	}
	if req.WindowSeconds <= 0 {
		writeError(w, http.StatusBadRequest, "window_seconds must be positive")
		return
	}
	if req.MaxRequests <= 0 {
		writeError(w, http.StatusBadRequest, "max_requests must be positive")
		return
	}

	s.ruleCounter++
	rule := &Rule{
		ID:           fmt.Sprintf("rule-%d", s.ruleCounter),
		Dimension:    dim,
		WindowSeconds: req.WindowSeconds,
		MaxRequests:  req.MaxRequests,
		CreatedAt:    time.Now(),
	}

	s.rulesMu.Lock()
	s.rules[rule.ID] = rule
	s.rulesMu.Unlock()

	writeJSON(w, http.StatusCreated, rule)
}

func (s *Service) DeleteRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	prefix := "/rules/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		writeError(w, http.StatusNotFound, "Not found")
		return
	}
	ruleID := strings.TrimPrefix(r.URL.Path, prefix)
	if ruleID == "" {
		writeError(w, http.StatusBadRequest, "Rule ID required")
		return
	}

	s.rulesMu.Lock()
	_, exists := s.rules[ruleID]
	if !exists {
		s.rulesMu.Unlock()
		writeError(w, http.StatusNotFound, "Rule not found")
		return
	}
	delete(s.rules, ruleID)
	s.rulesMu.Unlock()

	s.clearCoolDownsForRule(ruleID)

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) CreateKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	key := generateKey()
	s.keysMu.Lock()
	s.apiKeys[key] = struct{}{}
	delete(s.revokedKeys, key)
	s.keysMu.Unlock()

	type KeyResponse struct {
		Key string `json:"key"`
	}
	writeJSON(w, http.StatusCreated, KeyResponse{Key: key})
}

func (s *Service) DeleteKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	prefix := "/keys/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		writeError(w, http.StatusNotFound, "Not found")
		return
	}
	key := strings.TrimPrefix(r.URL.Path, prefix)
	if key == "" {
		writeError(w, http.StatusBadRequest, "Key required")
		return
	}

	s.keysMu.Lock()
	_, exists := s.apiKeys[key]
	if !exists {
		s.keysMu.Unlock()
		writeError(w, http.StatusNotFound, "Key not found")
		return
	}
	delete(s.apiKeys, key)
	s.revokedKeys[key] = struct{}{}
	s.keysMu.Unlock()

	s.countersMu.Lock()
	delete(s.keyCounters, key)
	s.countersMu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) CheckRateLimit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	now := time.Now()
	clientIP := s.getClientIP(r)
	apiKey := s.getAPIKey(r)

	s.coolDownMu.RLock()
	if coolDown, exists := s.coolDowns[clientIP]; exists && now.Before(coolDown.Until) {
		retryAfter := int(coolDown.Until.Sub(now).Seconds()) + 1
		s.coolDownMu.RUnlock()
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	s.coolDownMu.RUnlock()

	if apiKey != "" {
		s.keysMu.RLock()
		_, revoked := s.revokedKeys[apiKey]
		s.keysMu.RUnlock()
		if revoked {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
	}

	s.rulesMu.RLock()
	rulesCopy := make([]*Rule, 0, len(s.rules))
	for _, r := range s.rules {
		rulesCopy = append(rulesCopy, r)
	}
	s.rulesMu.RUnlock()

	ipExceededRules := make([]string, 0)
	keyExceededRules := make([]string, 0)
	shouldLimit := false
	retryAfterSeconds := 0

	for _, rule := range rulesCopy {
		if rule.Dimension == DimensionIP {
			limited, _ := s.checkAndIncrement(DimensionIP, clientIP, rule, now)
			if limited {
				shouldLimit = true
				ipExceededRules = append(ipExceededRules, rule.ID)
				if rule.WindowSeconds > retryAfterSeconds {
					retryAfterSeconds = rule.WindowSeconds
				}
			}
		} else if rule.Dimension == DimensionAPIKey && apiKey != "" {
			limited, _ := s.checkAndIncrement(DimensionAPIKey, apiKey, rule, now)
			if limited {
				shouldLimit = true
				keyExceededRules = append(keyExceededRules, rule.ID)
				if rule.WindowSeconds > retryAfterSeconds {
					retryAfterSeconds = rule.WindowSeconds
				}
			}
		}
	}

	if !shouldLimit {
		w.WriteHeader(http.StatusOK)
		return
	}

	if len(ipExceededRules) >= 2 {
		coolDownUntil := now.Add(time.Duration(retryAfterSeconds) * time.Second)
		s.coolDownMu.Lock()
		if entry, exists := s.coolDowns[clientIP]; exists {
			if coolDownUntil.After(entry.Until) {
				entry.Until = coolDownUntil
			}
			for _, ruleID := range ipExceededRules {
				entry.RuleIDs[ruleID] = struct{}{}
			}
		} else {
			ruleIDs := make(map[string]struct{})
			for _, ruleID := range ipExceededRules {
				ruleIDs[ruleID] = struct{}{}
			}
			s.coolDowns[clientIP] = &CoolDownEntry{
				Until:   coolDownUntil,
				RuleIDs: ruleIDs,
			}
		}
		s.coolDownMu.Unlock()
	}

	w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
	w.WriteHeader(http.StatusTooManyRequests)
}

func (s *Service) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	service := NewService()

	mux := http.NewServeMux()

	mux.HandleFunc("/rules", service.CreateRule)
	mux.HandleFunc("/rules/", service.DeleteRule)
	mux.HandleFunc("/keys", service.CreateKey)
	mux.HandleFunc("/keys/", service.DeleteKey)
	mux.HandleFunc("/check", service.CheckRateLimit)
	mux.HandleFunc("/health", service.Health)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("Rate limiter service starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
