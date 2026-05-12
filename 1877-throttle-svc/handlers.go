package main

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	ruleStore  *RuleStore
	statsStore *StatsStore
}

func NewHandler(rs *RuleStore, ss *StatsStore) *Handler {
	return &Handler{
		ruleStore:  rs,
		statsStore: ss,
	}
}

func (h *Handler) HandleRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listRules(w, r)
	case http.MethodPost:
		h.createRule(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) listRules(w http.ResponseWriter, _ *http.Request) {
	rules := h.ruleStore.GetAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}

func (h *Handler) createRule(w http.ResponseWriter, r *http.Request) {
	var req RuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Dimension != DimensionIP && req.Dimension != DimensionAPIKey {
		http.Error(w, "Invalid dimension, must be 'ip' or 'apikey'", http.StatusBadRequest)
		return
	}

	if req.Mode != ModeFixed && req.Mode != ModeSliding {
		http.Error(w, "Invalid mode, must be 'fixed' or 'sliding'", http.StatusBadRequest)
		return
	}

	if req.WindowSeconds <= 0 {
		http.Error(w, "window_seconds must be positive", http.StatusBadRequest)
		return
	}

	if req.MaxRequests <= 0 {
		http.Error(w, "max_requests must be positive", http.StatusBadRequest)
		return
	}

	rule := h.ruleStore.Create(req)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

func (h *Handler) HandleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := h.statsStore.GetAllStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *Handler) HandleCheck(w http.ResponseWriter, r *http.Request) {
	ipRules := h.ruleStore.GetByDimension(DimensionIP)
	apiKeyRules := h.ruleStore.GetByDimension(DimensionAPIKey)

	clientIP := getClientIP(r)
	for _, rule := range ipRules {
		key := StatsKey{RuleID: rule.ID, Key: clientIP}
		result := h.statsStore.CheckAndRecord(key, rule)
		if !result.Allowed {
			w.Header().Set("Retry-After", intToSeconds(result.RetryAfter))
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
	}

	apiKey := getAPIKey(r)
	if apiKey != "" {
		for _, rule := range apiKeyRules {
			key := StatsKey{RuleID: rule.ID, Key: apiKey}
			result := h.statsStore.CheckAndRecord(key, rule)
			if !result.Allowed {
				w.Header().Set("Retry-After", intToSeconds(result.RetryAfter))
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
