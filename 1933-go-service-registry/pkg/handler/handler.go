package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"registry/pkg/healthcheck"
	"registry/pkg/model"
	"registry/pkg/store"
)

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return hex.EncodeToString(b[:4]) + "-" +
		hex.EncodeToString(b[4:6]) + "-" +
		hex.EncodeToString(b[6:8]) + "-" +
		hex.EncodeToString(b[8:10]) + "-" +
		hex.EncodeToString(b[10:])
}

type Handler struct {
	store   *store.InstanceStore
	checker *healthcheck.HealthChecker
}

func NewHandler(s *store.InstanceStore, c *healthcheck.HealthChecker) *Handler {
	return &Handler{store: s, checker: c}
}

type RegisterRequest struct {
	ID          string              `json:"id"`
	Address     string              `json:"address"`
	Port        int                 `json:"port"`
	Weight      int                 `json:"weight"`
	HealthCheck HealthCheckConfigReq `json:"health_check"`
}

type HealthCheckConfigReq struct {
	Path     string `json:"path"`
	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`
}

type InstanceResponse struct {
	ID            string    `json:"id"`
	Address       string    `json:"address"`
	Port          int       `json:"port"`
	Weight        int       `json:"weight"`
	Status        string    `json:"status"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}

	switch {
	case len(parts) == 4 && parts[1] == "services" && parts[3] == "instances":
		if r.Method == http.MethodPost {
			h.registerInstance(w, r, parts[2])
			return
		}
		if r.Method == http.MethodGet {
			h.listInstances(w, r, parts[2])
			return
		}

	case len(parts) == 4 && parts[1] == "instances" && parts[3] == "heartbeat":
		if r.Method == http.MethodPut {
			h.heartbeat(w, r, parts[2])
			return
		}

	case len(parts) == 4 && parts[1] == "instances" && parts[3] == "promote":
		if r.Method == http.MethodPost {
			h.promote(w, r, parts[2])
			return
		}

	case len(parts) == 3 && parts[1] == "instances":
		if r.Method == http.MethodGet {
			h.getInstance(w, r, parts[2])
			return
		}
		if r.Method == http.MethodDelete {
			h.offlineInstance(w, r, parts[2])
			return
		}
	}

	http.NotFound(w, r)
}

func (h *Handler) registerInstance(w http.ResponseWriter, r *http.Request, serviceName string) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Address == "" || req.Port <= 0 {
		http.Error(w, "address and port are required", http.StatusBadRequest)
		return
	}

	interval, err := time.ParseDuration(req.HealthCheck.Interval)
	if err != nil {
		http.Error(w, "invalid health check interval", http.StatusBadRequest)
		return
	}
	timeout, err := time.ParseDuration(req.HealthCheck.Timeout)
	if err != nil {
		http.Error(w, "invalid health check timeout", http.StatusBadRequest)
		return
	}

	instanceID := req.ID
	if instanceID == "" {
		instanceID = generateUUID()
	}

	if _, exists := h.store.Get(instanceID); exists {
		http.Error(w, "instance id already exists", http.StatusConflict)
		return
	}

	hc := model.HealthCheckConfig{
		Path:     req.HealthCheck.Path,
		Interval: interval,
		Timeout:  timeout,
	}

	weight := req.Weight
	if weight <= 0 {
		weight = 1
	}

	inst := model.NewInstance(instanceID, serviceName, req.Address, req.Port, weight, hc)
	h.store.Add(inst)
	h.checker.Start(inst)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": instanceID})
}

func (h *Handler) listInstances(w http.ResponseWriter, r *http.Request, serviceName string) {
	includeCanary := false
	if param := r.URL.Query().Get("include_canary"); param != "" {
		includeCanary, _ = strconv.ParseBool(param)
	}

	instances := h.store.ListByService(serviceName, includeCanary)
	response := make([]InstanceResponse, 0, len(instances))
	for _, inst := range instances {
		response = append(response, InstanceResponse{
			ID:            inst.ID,
			Address:       inst.Address,
			Port:          inst.Port,
			Weight:        inst.Weight,
			Status:        string(inst.GetStatus()),
			LastHeartbeat: inst.LastHeartbeat,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) heartbeat(w http.ResponseWriter, r *http.Request, id string) {
	inst, ok := h.store.Get(id)
	if !ok {
		http.Error(w, "instance not found", http.StatusNotFound)
		return
	}

	inst.UpdateHeartbeat()
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) promote(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.store.CanPromote(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) getInstance(w http.ResponseWriter, r *http.Request, id string) {
	inst, ok := h.store.Get(id)
	if !ok {
		http.Error(w, "instance not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inst)
}

func (h *Handler) offlineInstance(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.store.StartDraining(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
