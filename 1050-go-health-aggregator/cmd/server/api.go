package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/health-aggregator/pkg/common"
	"github.com/health-aggregator/pkg/core"
)

const Version = "1.0.0"

type APIServer struct {
	store     *core.StateStore
	scheduler *core.Scheduler
	router    *http.ServeMux
}

func NewAPIServer(store *core.StateStore, scheduler *core.Scheduler) *APIServer {
	srv := &APIServer{
		store:     store,
		scheduler: scheduler,
		router:    http.NewServeMux(),
	}
	srv.registerRoutes()
	return srv
}

func (s *APIServer) registerRoutes() {
	s.router.HandleFunc("/health", s.handleHealth)
	s.router.HandleFunc("/api/status", s.handleAggregateStatus)
	s.router.HandleFunc("/api/services", s.handleServices)
	s.router.HandleFunc("/api/services/", s.handleServiceByName)
}

func (s *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	agg := s.scheduler.GetAggregator().CalculateAggregate()

	resp := common.ServerHealthResponse{
		Version:   Version,
		Status:    agg.OverallStatus,
		Timestamp: time.Now(),
		Aggregate: *agg,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *APIServer) handleAggregateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	agg := s.scheduler.GetAggregator().CalculateAggregate()
	writeJSON(w, http.StatusOK, agg)
}

func (s *APIServer) handleServices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listServices(w, r)
	case http.MethodPost:
		s.addService(w, r)
	case http.MethodDelete:
		s.removeService(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *APIServer) listServices(w http.ResponseWriter, r *http.Request) {
	names := s.store.GetAllServiceNames()
	services := make([]*common.ServiceStatus, 0, len(names))
	for _, name := range names {
		if status, ok := s.store.GetServiceStatus(name); ok {
			services = append(services, status)
		}
	}
	writeJSON(w, http.StatusOK, services)
}

func (s *APIServer) addService(w http.ResponseWriter, r *http.Request) {
	var req common.AddServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "service name is required")
		return
	}
	if req.ProbeType != common.ProbeTypeHTTP && req.ProbeType != common.ProbeTypeTCP {
		writeError(w, http.StatusBadRequest, "invalid probe type")
		return
	}
	if req.Target == "" {
		writeError(w, http.StatusBadRequest, "target is required")
		return
	}

	if s.store.ServiceExists(req.Name) {
		writeError(w, http.StatusConflict, "service already exists")
		return
	}

	s.store.AddService(&req.ServiceConfig)
	writeJSON(w, http.StatusCreated, common.AddServiceResponse{
		Success: true,
		Message: "service added",
	})
}

func (s *APIServer) removeService(w http.ResponseWriter, r *http.Request) {
	var req common.RemoveServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "service name is required")
		return
	}

	if !s.store.ServiceExists(req.Name) {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}

	s.store.RemoveService(req.Name)
	writeJSON(w, http.StatusOK, common.RemoveServiceResponse{
		Success: true,
		Message: "service removed",
	})
}

func (s *APIServer) handleServiceByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/services/")
	parts := strings.SplitN(path, "/", 2)
	name := parts[0]
	if name == "" {
		writeError(w, http.StatusBadRequest, "service name is required")
		return
	}

	if len(parts) > 1 && parts[1] == "history" {
		s.getServiceHistory(w, r, name)
		return
	}

	s.getServiceStatus(w, r, name)
}

func (s *APIServer) getServiceStatus(w http.ResponseWriter, r *http.Request, name string) {
	status, ok := s.store.GetServiceStatus(name)
	if !ok {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *APIServer) getServiceHistory(w http.ResponseWriter, r *http.Request, name string) {
	history, ok := s.store.GetServiceHistory(name)
	if !ok {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}

	resp := common.ProbeHistory{
		Total:   len(history),
		Results: history,
	}
	writeJSON(w, http.StatusOK, resp)
}
