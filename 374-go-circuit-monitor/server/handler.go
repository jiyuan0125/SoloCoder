package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"circuit-monitor/circuitbreaker"
	"circuit-monitor/common"
)

type Handler struct {
	registry *circuitbreaker.Registry
}

func NewHandler(registry *circuitbreaker.Registry) *Handler {
	return &Handler{registry: registry}
}

func (h *Handler) ListCircuits(w http.ResponseWriter, r *http.Request) {
	names := h.registry.List()
	response := common.ListCircuitsResponse{
		Circuits: names,
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) GetCircuit(w http.ResponseWriter, r *http.Request) {
	name := getCircuitName(r)
	if name == "" {
		writeError(w, http.StatusBadRequest, "circuit name is required")
		return
	}

	cb, exists := h.registry.Get(name)
	if !exists {
		writeError(w, http.StatusNotFound, "circuit not found")
		return
	}

	stats := cb.GetStatistics()
	state := cb.State()

	var openAt *time.Time
	var lastStateChange *time.Time

	if cb.State() == circuitbreaker.StateOpen {
		t := cb.GetOpenAt()
		if !t.IsZero() {
			openAt = &t
		}
	}

	lsc := cb.GetLastStateChange()
	if !lsc.IsZero() {
		lastStateChange = &lsc
	}

	response := common.GetCircuitResponse{
		CircuitInfo: common.CircuitInfo{
			Name:            name,
			State:           common.CircuitState(state),
			TotalRequests:   stats.TotalRequests,
			SuccessCount:    stats.SuccessCount,
			FailureCount:    stats.FailureCount,
			RecentFailures:  stats.RecentFailures,
			RecentSuccesses: stats.RecentSuccesses,
			OpenAt:          openAt,
			LastStateChange: lastStateChange,
		},
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) CreateCircuit(w http.ResponseWriter, r *http.Request) {
	var req common.CreateCircuitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "circuit name is required")
		return
	}

	var err error
	if req.Config != nil {
		config := circuitbreaker.Config{
			WindowSize:          req.Config.WindowSize,
			FailureThreshold:    req.Config.FailureThreshold,
			OpenTimeout:         req.Config.OpenTimeout,
			HalfOpenMaxRequests: req.Config.HalfOpenMaxRequests,
			SuccessThreshold:    req.Config.SuccessThreshold,
		}
		if config.WindowSize == 0 {
			config.WindowSize = circuitbreaker.DefaultWindowSize
		}
		if config.FailureThreshold == 0 {
			config.FailureThreshold = circuitbreaker.DefaultFailureThreshold
		}
		if config.OpenTimeout == 0 {
			config.OpenTimeout = circuitbreaker.DefaultOpenTimeout
		}
		if config.HalfOpenMaxRequests == 0 {
			config.HalfOpenMaxRequests = circuitbreaker.DefaultHalfOpenMaxRequests
		}
		if config.SuccessThreshold == 0 {
			config.SuccessThreshold = circuitbreaker.DefaultSuccessThreshold
		}
		_, err = h.registry.Register(req.Name, config)
	} else {
		_, err = h.registry.Register(req.Name)
	}
	if err != nil {
		if strings.Contains(err.Error(), "already registered") {
			writeError(w, http.StatusConflict, err.Error())
		} else {
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response := common.CreateCircuitResponse{Success: true}
	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) ResetCircuit(w http.ResponseWriter, r *http.Request) {
	name := getCircuitName(r)
	if name == "" {
		writeError(w, http.StatusBadRequest, "circuit name is required")
		return
	}

	cb, exists := h.registry.Get(name)
	if !exists {
		writeError(w, http.StatusNotFound, "circuit not found")
		return
	}

	cb.Reset()

	response := common.ResetCircuitResponse{Success: true}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) ForceState(w http.ResponseWriter, r *http.Request) {
	name := getCircuitName(r)
	if name == "" {
		writeError(w, http.StatusBadRequest, "circuit name is required")
		return
	}

	var req common.ForceStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cb, exists := h.registry.Get(name)
	if !exists {
		writeError(w, http.StatusNotFound, "circuit not found")
		return
	}

	var state circuitbreaker.State
	switch req.State {
	case common.StateClosed:
		state = circuitbreaker.StateClosed
	case common.StateOpen:
		state = circuitbreaker.StateOpen
	case common.StateHalfOpen:
		state = circuitbreaker.StateHalfOpen
	default:
		writeError(w, http.StatusBadRequest, "invalid state")
		return
	}

	cb.ForceState(state)

	response := common.ForceStateResponse{Success: true}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	name := getCircuitName(r)
	if name == "" {
		writeError(w, http.StatusBadRequest, "circuit name is required")
		return
	}

	cb, exists := h.registry.Get(name)
	if !exists {
		writeError(w, http.StatusNotFound, "circuit not found")
		return
	}

	cfg := cb.GetConfig()
	response := common.GetConfigResponse{
		Config: common.CircuitConfig{
			WindowSize:          cfg.WindowSize,
			FailureThreshold:    cfg.FailureThreshold,
			OpenTimeout:         cfg.OpenTimeout,
			HalfOpenMaxRequests: cfg.HalfOpenMaxRequests,
			SuccessThreshold:    cfg.SuccessThreshold,
		},
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) SaveState(w http.ResponseWriter, r *http.Request) {
	name := getCircuitName(r)
	if name == "" {
		writeError(w, http.StatusBadRequest, "circuit name is required")
		return
	}

	cb, exists := h.registry.Get(name)
	if !exists {
		writeError(w, http.StatusNotFound, "circuit not found")
		return
	}

	path := name + ".json"
	if err := cb.SaveToFile(path); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save state: "+err.Error())
		return
	}

	response := common.SaveStateResponse{Success: true}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) LoadState(w http.ResponseWriter, r *http.Request) {
	name := getCircuitName(r)
	if name == "" {
		writeError(w, http.StatusBadRequest, "circuit name is required")
		return
	}

	cb, exists := h.registry.Get(name)
	if !exists {
		writeError(w, http.StatusNotFound, "circuit not found")
		return
	}

	path := name + ".json"
	if err := cb.LoadFromFile(path); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load state: "+err.Error())
		return
	}

	response := common.LoadStateResponse{Success: true}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func getCircuitName(r *http.Request) string {
	return r.PathValue("name")
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	response := common.ErrorResponse{
		Error: message,
	}
	writeJSON(w, status, response)
}
