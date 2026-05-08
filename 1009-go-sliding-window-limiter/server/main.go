package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/example/sliding-window-limiter/common"
	"github.com/example/sliding-window-limiter/limiter"
)

type Server struct {
	managers sync.Map
	mux      *http.ServeMux
}

func NewServer() *Server {
	s := &Server{
		mux: http.NewServeMux(),
	}
	s.setupRoutes()
	return s
}

func (s *Server) getOrCreateManager(endpoint string) *limiter.LimiterManager {
	if m, ok := s.managers.Load(endpoint); ok {
		return m.(*limiter.LimiterManager)
	}
	m := limiter.NewLimiterManager()
	actual, _ := s.managers.LoadOrStore(endpoint, m)
	return actual.(*limiter.LimiterManager)
}

func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/api/test", s.handleTest)
	s.mux.HandleFunc("/admin/rules", s.handleRules)
	s.mux.HandleFunc("/admin/status", s.handleStatus)
	s.mux.HandleFunc("/admin/reset", s.handleReset)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (s *Server) handleTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.TestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Endpoint == "" {
		writeError(w, http.StatusBadRequest, "Endpoint is required")
		return
	}

	manager := s.getOrCreateManager(req.Endpoint)
	result, ruleName, allowed := manager.Allow()

	if !allowed {
		retryAfterSeconds := int(result.RetryAfter / time.Second)
		if result.RetryAfter%time.Second > 0 {
			retryAfterSeconds++
		}

		w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
		writeJSON(w, http.StatusTooManyRequests, common.RateLimitResponse{
			Allowed:    false,
			Remaining:  result.Remaining,
			Limit:      result.Limit,
			RetryAfter: result.RetryAfter.Milliseconds(),
			RuleName:   ruleName,
			Message:    "Rate limit exceeded",
		})
		return
	}

	writeJSON(w, http.StatusOK, common.RateLimitResponse{
		Allowed:   true,
		Remaining: result.Remaining,
		Limit:     result.Limit,
	})
}

func (s *Server) handleRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleAddRule(w, r)
	case http.MethodDelete:
		s.handleRemoveRule(w, r)
	case http.MethodGet:
		s.handleListRules(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func dtoToRule(dto common.LimitRuleDTO) limiter.LimitRule {
	return limiter.LimitRule{
		Name:     dto.Name,
		Limit:    dto.Limit,
		Window:   time.Duration(dto.WindowMs) * time.Millisecond,
		GridSize: time.Duration(dto.GridMs) * time.Millisecond,
		Mode:     limiter.LimiterMode(dto.Mode),
		Priority: dto.Priority,
	}
}

func ruleToDTO(rule limiter.LimitRule) common.LimitRuleDTO {
	return common.LimitRuleDTO{
		Name:     rule.Name,
		Limit:    rule.Limit,
		WindowMs: rule.Window.Milliseconds(),
		GridMs:   rule.GridSize.Milliseconds(),
		Mode:     string(rule.Mode),
		Priority: rule.Priority,
	}
}

func (s *Server) handleAddRule(w http.ResponseWriter, r *http.Request) {
	var req common.AddRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Endpoint == "" {
		writeError(w, http.StatusBadRequest, "Endpoint is required")
		return
	}

	if req.Rule.Name == "" || req.Rule.Limit <= 0 || req.Rule.WindowMs <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid rule: name, limit and window are required")
		return
	}

	manager := s.getOrCreateManager(req.Endpoint)
	manager.AddRule(dtoToRule(req.Rule))

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleRemoveRule(w http.ResponseWriter, r *http.Request) {
	var req common.RemoveRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Endpoint == "" || req.RuleName == "" {
		writeError(w, http.StatusBadRequest, "Endpoint and rule_name are required")
		return
	}

	manager := s.getOrCreateManager(req.Endpoint)
	if manager.RemoveRule(req.RuleName) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	} else {
		writeError(w, http.StatusNotFound, "Rule not found")
	}
}

func (s *Server) handleListRules(w http.ResponseWriter, r *http.Request) {
	endpoint := r.URL.Query().Get("endpoint")
	if endpoint == "" {
		writeError(w, http.StatusBadRequest, "Endpoint parameter is required")
		return
	}

	manager := s.getOrCreateManager(endpoint)
	rules := manager.GetRules()

	dtos := make([]common.LimitRuleDTO, len(rules))
	for i, rule := range rules {
		dtos[i] = ruleToDTO(rule)
	}

	writeJSON(w, http.StatusOK, dtos)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	endpoints := make(map[string]common.EndpointStatsDTO)

	s.managers.Range(func(key, value interface{}) bool {
		endpoint := key.(string)
		manager := value.(*limiter.LimiterManager)

		rules := manager.GetRules()
		ruleDTOs := make([]common.LimitRuleDTO, len(rules))
		for i, rule := range rules {
			ruleDTOs[i] = ruleToDTO(rule)
		}

		endpoints[endpoint] = common.EndpointStatsDTO{
			Endpoint: endpoint,
			Rules:    ruleDTOs,
			Stats:    manager.Stats(),
		}
		return true
	})

	writeJSON(w, http.StatusOK, common.AdminStatusResponse{
		Endpoints: endpoints,
		Timestamp: time.Now(),
	})
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	endpoint := r.URL.Query().Get("endpoint")
	if endpoint == "" {
		s.managers.Range(func(key, value interface{}) bool {
			manager := value.(*limiter.LimiterManager)
			manager.Reset()
			return true
		})
	} else {
		manager := s.getOrCreateManager(endpoint)
		manager.Reset()
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func main() {
	server := NewServer()
	port := ":8080"

	fmt.Printf("Server starting on %s...\n", port)
	if err := http.ListenAndServe(port, server.mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
