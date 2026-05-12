package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	manager *ComponentManager
	checker *Checker
}

func NewHandler(manager *ComponentManager, checker *Checker) *Handler {
	return &Handler{
		manager: manager,
		checker: checker,
	}
}

type RegisterRequest struct {
	Name        string `json:"name"`
	CheckType   string `json:"checkType"`
	Address     string `json:"address"`
	Timeout     string `json:"timeout"`
	Interval    string `json:"interval"`
	CallbackURL string `json:"callbackUrl"`
}

type HealthResponse struct {
	Healthy bool            `json:"healthy"`
	Status  []ComponentInfo `json:"status"`
}

type ComponentInfo struct {
	Name        string        `json:"name"`
	Healthy     bool          `json:"healthy"`
	LastChecked time.Time     `json:"lastChecked"`
	History     []CheckResult `json:"history"`
}

func (h *Handler) RegisterComponent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}
	
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.CheckType == "" {
		http.Error(w, "checkType is required", http.StatusBadRequest)
		return
	}
	if req.Address == "" {
		http.Error(w, "address is required", http.StatusBadRequest)
		return
	}
	
	var checkType CheckType
	switch strings.ToLower(req.CheckType) {
	case "http":
		checkType = CheckTypeHTTP
	case "tcp":
		checkType = CheckTypeTCP
	default:
		http.Error(w, "invalid checkType, must be 'http' or 'tcp'", http.StatusBadRequest)
		return
	}
	
	config := ComponentConfig{
		Name:        req.Name,
		CheckType:   checkType,
		Address:     req.Address,
		CallbackURL: req.CallbackURL,
	}
	
	if req.Timeout != "" {
		timeout, err := time.ParseDuration(req.Timeout)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid timeout: %v", err), http.StatusBadRequest)
			return
		}
		config.Timeout = timeout
	}
	
	if req.Interval != "" {
		interval, err := time.ParseDuration(req.Interval)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid interval: %v", err), http.StatusBadRequest)
			return
		}
		config.Interval = interval
	}
	
	registered := h.manager.Register(config)
	if !registered {
		http.Error(w, "component already exists", http.StatusConflict)
		return
	}
	
	h.checker.StartComponent(req.Name)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "component registered",
		"name":    req.Name,
	})
}

func (h *Handler) UnregisterComponent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	name := strings.TrimPrefix(r.URL.Path, "/components/")
	if name == "" {
		http.Error(w, "component name is required", http.StatusBadRequest)
		return
	}
	
	h.checker.StopComponent(name)
	unregistered := h.manager.Unregister(name)
	
	if !unregistered {
		http.Error(w, "component not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "component unregistered",
		"name":    name,
	})
}

func (h *Handler) GetComponents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	components := h.manager.GetAll()
	statuses := make([]ComponentInfo, 0, len(components))
	for _, comp := range components {
		statuses = append(statuses, ComponentInfo{
			Name:        comp.Name,
			Healthy:     comp.Healthy,
			LastChecked: comp.LastChecked,
			History:     comp.History,
		})
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statuses)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	healthy := h.manager.AllHealthy()
	components := h.manager.GetAll()
	statuses := make([]ComponentInfo, 0, len(components))
	for _, comp := range components {
		statuses = append(statuses, ComponentInfo{
			Name:        comp.Name,
			Healthy:     comp.Healthy,
			LastChecked: comp.LastChecked,
			History:     comp.History,
		})
	}
	
	response := HealthResponse{
		Healthy: healthy,
		Status:  statuses,
	}
	
	w.Header().Set("Content-Type", "application/json")
	if healthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(response)
}
