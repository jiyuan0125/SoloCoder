package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"serviceregistry/internal/discovery"
	"serviceregistry/internal/events"
	"serviceregistry/internal/heartbeat"
	"serviceregistry/internal/model"
	"serviceregistry/internal/registry"
	"serviceregistry/internal/resource"
	"serviceregistry/internal/storage"
)

type Server struct {
	httpServer *http.Server
	registry   *registry.Registry
	discovery  *discovery.Discovery
	heartbeat  *heartbeat.Checker
	eventBus   *events.EventBus
	resource   *resource.Manager
	storage    *storage.SQLiteStorage
	mux        *http.ServeMux
}

type RegisterRequest struct {
	ServiceName string            `json:"service_name"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Version     string            `json:"version"`
	Weight      int               `json:"weight"`
	Environment string            `json:"environment"`
	Metadata    map[string]string `json:"metadata"`
	ResourceID  string            `json:"resource_id,omitempty"`
}

type DiscoverRequest struct {
	Strategy  string `json:"strategy"`
	Version   string `json:"version"`
	Environment string `json:"environment"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func New(dbPath string, port int) (*Server, error) {
	store, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	eventBus := events.NewEventBus()
	reg := registry.NewRegistry(store, eventBus)
	hb := heartbeat.NewCheckerWithStorage(store, eventBus)
	disco := discovery.NewDiscovery(reg)
	resMgr := resource.NewManager(store)

	if err := reg.Restore(); err != nil {
		store.Close()
		return nil, fmt.Errorf("failed to restore registry: %w", err)
	}

	if err := hb.CheckAll(); err != nil {
		store.Close()
		return nil, fmt.Errorf("failed to check instances: %w", err)
	}

	s := &Server{
		registry:  reg,
		discovery: disco,
		heartbeat: hb,
		eventBus:  eventBus,
		resource:  resMgr,
		storage:   store,
		mux:       http.NewServeMux(),
	}

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.mux,
	}

	s.routes()

	return s, nil
}

func (s *Server) Start() error {
	s.heartbeat.Start()
	fmt.Printf("Service Registry starting on %s...\n", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.heartbeat.Stop()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}
	return s.storage.Close()
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/v1/register", s.handleRegister)
	s.mux.HandleFunc("/api/v1/register/", s.handleRegisterWithID)
	s.mux.HandleFunc("/api/v1/unregister/", s.handleUnregister)
	s.mux.HandleFunc("/api/v1/heartbeat/", s.handleHeartbeat)
	s.mux.HandleFunc("/api/v1/discover/", s.handleDiscover)
	s.mux.HandleFunc("/api/v1/services/", s.handleServices)
	s.mux.HandleFunc("/api/v1/subscribe/", s.handleSubscribe)
	s.mux.HandleFunc("/api/v1/resources/", s.handleResources)
	s.mux.HandleFunc("/api/v1/resources", s.handleCreateResource)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Port <= 0 || req.Port > 65535 || req.Version == "" {
		writeError(w, http.StatusBadRequest, "invalid request: missing or invalid port/version")
		return
	}

	inst := &model.ServiceInstance{
		ServiceName: req.ServiceName,
		Address:     req.Address,
		Port:        req.Port,
		Version:     req.Version,
		Weight:      req.Weight,
		Environment: req.Environment,
		Metadata:    req.Metadata,
	}

	var res *model.Resource
	if req.ResourceID != "" {
		res = &model.Resource{
			ID:   req.ResourceID,
			Type: model.ResourceService,
			Name: req.ServiceName,
		}
	}

	if err := s.registry.Register(inst, res); err != nil {
		if err == registry.ErrInstanceExists {
			writeError(w, http.StatusConflict, "instance already registered")
			return
		}
		if err == registry.ErrInvalidRequest {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, http.StatusCreated, map[string]interface{}{
		"id":       inst.ID,
		"status":   inst.Status,
		"registered_at": inst.CreatedAt,
	})
}

func (s *Server) handleRegisterWithID(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (s *Server) handleUnregister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/unregister/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "instance id is required")
		return
	}

	var req struct {
		ResourceID string `json:"resource_id,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var res *model.Resource
	if req.ResourceID != "" {
		inst, err := s.registry.GetInstance(id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if inst != nil {
			res = &model.Resource{
				ID:   req.ResourceID,
				Type: model.ResourceService,
				Name: inst.ServiceName,
			}
		}
	}

	if err := s.registry.Unregister(id, res); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, nil)
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/heartbeat/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "instance id is required")
		return
	}

	inst, err := s.registry.GetInstance(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if inst == nil {
		writeError(w, http.StatusNotFound, "instance not found")
		return
	}

	if err := s.heartbeat.RecordHeartbeat(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if inst.Status != model.StatusHealthy {
		s.storage.UpdateInstanceStatus(id, model.StatusHealthy)
		s.eventBus.Publish(&model.ServiceEvent{
			Type:      model.EventHealthy,
			Instance:  inst,
			Timestamp: time.Now(),
		})
	}

	writeSuccess(w, http.StatusOK, nil)
}

func (s *Server) handleDiscover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	serviceName := strings.TrimPrefix(r.URL.Path, "/api/v1/discover/")
	if serviceName == "" {
		writeError(w, http.StatusBadRequest, "service name is required")
		return
	}

	version := r.URL.Query().Get("version")
	env := r.URL.Query().Get("environment")
	strategy := r.URL.Query().Get("strategy")

	if strategy == "select" || strategy == "round_robin" || strategy == "weighted_random" {
		var strat model.LoadBalanceStrategy
		switch strategy {
		case "round_robin":
			strat = model.StrategyRoundRobin
		case "weighted_random":
			strat = model.StrategyWeightedRandom
		default:
			strat = model.StrategyRoundRobin
		}

		inst := s.discovery.SelectInstanceWithFilter(serviceName, strat, version, env)
		if inst == nil {
			writeSuccess(w, http.StatusOK, []interface{}{})
			return
		}
		writeSuccess(w, http.StatusOK, inst)
		return
	}

	instances := s.discovery.ListHealthyInstances(serviceName, version, env)
	if instances == nil {
		instances = []*model.ServiceInstance{}
	}
	writeSuccess(w, http.StatusOK, instances)
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	serviceName := strings.TrimPrefix(r.URL.Path, "/api/v1/services/")
	if serviceName == "" {
		writeError(w, http.StatusBadRequest, "service name is required")
		return
	}

	instances := s.registry.GetInstances(serviceName)
	if len(instances) == 0 {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}

	writeSuccess(w, http.StatusOK, instances)
}

func (s *Server) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	serviceName := strings.TrimPrefix(r.URL.Path, "/api/v1/subscribe/")
	if serviceName == "" {
		writeError(w, http.StatusBadRequest, "service name is required")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	sub := events.NewChannelSubscriber(10)
	s.eventBus.Subscribe(serviceName, sub)
	defer s.eventBus.Unsubscribe(serviceName, sub)
	defer sub.Close()

	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-sub.Events():
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, string(data))
			flusher.Flush()
		}
	}
}

func (s *Server) handleCreateResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.ID == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "id and name are required")
		return
	}

	var rType model.ResourceType
	switch req.Type {
	case "service":
		rType = model.ResourceService
	case "instance":
		rType = model.ResourceInstance
	default:
		writeError(w, http.StatusBadRequest, "invalid resource type")
		return
	}

	if err := s.resource.Create(req.ID, rType, req.Name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, http.StatusCreated, nil)
}

func (s *Server) handleResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	resourceID := strings.TrimPrefix(r.URL.Path, "/api/v1/resources/")
	if resourceID == "" {
		writeError(w, http.StatusBadRequest, "resource id is required")
		return
	}

	res, err := s.resource.Get(resourceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if res == nil {
		writeError(w, http.StatusNotFound, "resource not found")
		return
	}

	summary, err := s.resource.GetSummary(resourceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"resource": res,
		"operations": summary,
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Message: message,
	})
}

func writeSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    data,
	})
}

var _ = sync.Mutex{}
