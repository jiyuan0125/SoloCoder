package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultGroup     = "production"
	DefaultHTTPPort  = 9105
	HeartbeatCheckInterval = 1 * time.Second
)

type GroupConfig struct {
	HeartbeatTimeout time.Duration
}

var defaultGroupConfigs = map[string]GroupConfig{
	"production": {HeartbeatTimeout: 30 * time.Second},
	"canary":     {HeartbeatTimeout: 15 * time.Second},
	"staging":    {HeartbeatTimeout: 30 * time.Second},
}

type Instance struct {
	ID        string            `json:"id"`
	Host      string            `json:"host"`
	Port      int               `json:"port"`
	Tags      []string          `json:"tags"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type ServiceGroup struct {
	GroupName  string              `json:"group"`
	Instances  map[string]*Instance `json:"instances"`
	lastSeen   map[string]time.Time
	mu         sync.RWMutex
}

type Registry struct {
	groups        map[string]*ServiceGroup
	serviceGroups map[string]map[string]*ServiceGroup
	groupConfigs  map[string]GroupConfig
	mu            sync.RWMutex
	lastAlert     map[string]map[string]bool
	lastAlertMu   sync.RWMutex
}

func NewRegistry() *Registry {
	r := &Registry{
		groups:        make(map[string]*ServiceGroup),
		serviceGroups: make(map[string]map[string]*ServiceGroup),
		groupConfigs:  make(map[string]GroupConfig),
		lastAlert:     make(map[string]map[string]bool),
		lastAlertMu:   sync.RWMutex{},
	}
	for name, cfg := range defaultGroupConfigs {
		r.groupConfigs[name] = cfg
	}
	return r
}

func (r *Registry) getOrCreateServiceGroup(serviceName, groupName string) *ServiceGroup {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.serviceGroups[serviceName]; !ok {
		r.serviceGroups[serviceName] = make(map[string]*ServiceGroup)
	}

	if _, ok := r.lastAlert[serviceName]; !ok {
		r.lastAlert[serviceName] = make(map[string]bool)
	}

	sg, ok := r.serviceGroups[serviceName][groupName]
	if !ok {
		sg = &ServiceGroup{
			GroupName: groupName,
			Instances: make(map[string]*Instance),
			lastSeen:  make(map[string]time.Time),
		}
		r.serviceGroups[serviceName][groupName] = sg
		r.groups[groupName] = sg
	}

	return sg
}

func (r *Registry) getServiceGroup(serviceName, groupName string) (*ServiceGroup, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sgs, ok := r.serviceGroups[serviceName]
	if !ok {
		return nil, false
	}

	sg, ok := sgs[groupName]
	return sg, ok
}

func (r *Registry) getGroupConfig(groupName string) GroupConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if cfg, ok := r.groupConfigs[groupName]; ok {
		return cfg
	}
	return GroupConfig{HeartbeatTimeout: 30 * time.Second}
}

func (r *Registry) Register(serviceName string, inst *Instance) error {
	if serviceName == "" {
		return fmt.Errorf("service name is required")
	}
	if inst.ID == "" {
		return fmt.Errorf("instance id is required")
	}
	if inst.Host == "" {
		return fmt.Errorf("instance host is required")
	}
	if inst.Port <= 0 {
		return fmt.Errorf("instance port is required")
	}
	if len(inst.Tags) == 0 {
		return fmt.Errorf("at least one group tag is required (e.g., production, canary)")
	}

	for _, tag := range inst.Tags {
		if tag == "" {
			return fmt.Errorf("group tag cannot be empty")
		}
	}

	now := time.Now()
	inst.CreatedAt = now
	inst.UpdatedAt = now

	for _, groupName := range inst.Tags {
		sg := r.getOrCreateServiceGroup(serviceName, groupName)
		sg.mu.Lock()
		sg.Instances[inst.ID] = inst
		sg.lastSeen[inst.ID] = now
		sg.mu.Unlock()

		r.lastAlertMu.Lock()
		r.lastAlert[serviceName][groupName] = false
		r.lastAlertMu.Unlock()

		log.Printf("Instance %s registered in service %s group %s", inst.ID, serviceName, groupName)
	}

	return nil
}

func (r *Registry) Deregister(serviceName, instanceID string) error {
	if serviceName == "" {
		return fmt.Errorf("service name is required")
	}
	if instanceID == "" {
		return fmt.Errorf("instance id is required")
	}

	r.mu.RLock()
	sgs, ok := r.serviceGroups[serviceName]
	if !ok {
		r.mu.RUnlock()
		return fmt.Errorf("service %s not found", serviceName)
	}

	groups := make([]*ServiceGroup, 0, len(sgs))
	for _, sg := range sgs {
		groups = append(groups, sg)
	}
	r.mu.RUnlock()

	found := false
	for _, sg := range groups {
		sg.mu.Lock()
		if _, exists := sg.Instances[instanceID]; exists {
			delete(sg.Instances, instanceID)
			delete(sg.lastSeen, instanceID)
			log.Printf("Instance %s deregistered from service %s group %s", instanceID, serviceName, sg.GroupName)
			found = true

			if len(sg.Instances) == 0 {
				r.triggerAlert(serviceName, sg.GroupName)
			}
		}
		sg.mu.Unlock()
	}

	if !found {
		return fmt.Errorf("instance %s not found in service %s", instanceID, serviceName)
	}

	return nil
}

func (r *Registry) GetServices() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]string, 0, len(r.serviceGroups))
	for name := range r.serviceGroups {
		services = append(services, name)
	}
	return services
}

func (r *Registry) GetServiceInstances(serviceName, groupName string) []*Instance {
	sg, ok := r.getServiceGroup(serviceName, groupName)
	if !ok {
		return []*Instance{}
	}

	sg.mu.RLock()
	defer sg.mu.RUnlock()

	instances := make([]*Instance, 0, len(sg.Instances))
	for _, inst := range sg.Instances {
		instances = append(instances, inst)
	}
	return instances
}

func (r *Registry) GetServiceGroups(serviceName string) map[string][]*Instance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sgs, ok := r.serviceGroups[serviceName]
	if !ok {
		return make(map[string][]*Instance)
	}

	result := make(map[string][]*Instance)
	for groupName, sg := range sgs {
		sg.mu.RLock()
		instances := make([]*Instance, 0, len(sg.Instances))
		for _, inst := range sg.Instances {
			instances = append(instances, inst)
		}
		sg.mu.RUnlock()
		result[groupName] = instances
	}
	return result
}

func (r *Registry) Heartbeat(serviceName, instanceID string) error {
	if serviceName == "" {
		return fmt.Errorf("service name is required")
	}
	if instanceID == "" {
		return fmt.Errorf("instance id is required")
	}

	r.mu.RLock()
	sgs, ok := r.serviceGroups[serviceName]
	if !ok {
		r.mu.RUnlock()
		return fmt.Errorf("service %s not found", serviceName)
	}

	groups := make([]*ServiceGroup, 0, len(sgs))
	for _, sg := range sgs {
		groups = append(groups, sg)
	}
	r.mu.RUnlock()

	now := time.Now()
	found := false
	for _, sg := range groups {
		sg.mu.Lock()
		if _, exists := sg.Instances[instanceID]; exists {
			sg.lastSeen[instanceID] = now
			sg.Instances[instanceID].UpdatedAt = now
			found = true
		}
		sg.mu.Unlock()
	}

	if !found {
		return fmt.Errorf("instance %s not found in service %s", instanceID, serviceName)
	}

	return nil
}

func (r *Registry) CheckTimeouts() {
	r.mu.RLock()
	services := make(map[string]map[string]*ServiceGroup)
	for name, sgs := range r.serviceGroups {
		services[name] = sgs
	}
	r.mu.RUnlock()

	for serviceName, sgs := range services {
		for groupName, sg := range sgs {
			cfg := r.getGroupConfig(groupName)
			now := time.Now()

			sg.mu.Lock()
			toRemove := make([]string, 0)

			for instID, lastSeen := range sg.lastSeen {
				if now.Sub(lastSeen) > cfg.HeartbeatTimeout {
					toRemove = append(toRemove, instID)
					log.Printf("Instance %s in service %s group %s timed out", instID, serviceName, groupName)
				}
			}

			for _, instID := range toRemove {
				delete(sg.Instances, instID)
				delete(sg.lastSeen, instID)
			}

			if len(sg.Instances) == 0 && len(toRemove) > 0 {
				sg.mu.Unlock()
				r.triggerAlert(serviceName, groupName)
			} else {
				sg.mu.Unlock()
			}
		}
	}
}

func (r *Registry) triggerAlert(serviceName, groupName string) {
	r.lastAlertMu.Lock()
	defer r.lastAlertMu.Unlock()

	if alerts, ok := r.lastAlert[serviceName]; ok {
		if !alerts[groupName] {
			alerts[groupName] = true
			log.Printf("[ALERT] All instances of service %s in group %s are offline!", serviceName, groupName)
		}
	}
}

func (r *Registry) DeregisterUnhealthy(serviceName, instanceID string) error {
	return r.Deregister(serviceName, instanceID)
}

type RegisterRequest struct {
	ID       string            `json:"id"`
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type Server struct {
	registry *Registry
}

func NewServer(r *Registry) *Server {
	return &Server{registry: r}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasPrefix(path, "/services/") {
		s.handleServices(w, r)
		return
	}

	if path == "/services" {
		s.handleListServices(w, r)
		return
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func (s *Server) handleListServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	services := s.registry.GetServices()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"services": services,
	})
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/services/")
	parts := strings.Split(path, "/")

	if len(parts) < 1 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	serviceName := parts[0]

	if len(parts) == 2 {
		if parts[1] == "groups" {
			s.handleServiceGroups(w, r, serviceName)
			return
		}
		if parts[1] == "instances" {
			s.handleServiceInstances(w, r, serviceName)
			return
		}
	}

	if len(parts) == 3 {
		if parts[1] == "instances" {
			s.handleServiceInstance(w, r, serviceName, parts[2])
			return
		}
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func (s *Server) handleServiceGroups(w http.ResponseWriter, r *http.Request, serviceName string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	groups := s.registry.GetServiceGroups(serviceName)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": serviceName,
		"groups":  groups,
	})
}

func (s *Server) handleServiceInstances(w http.ResponseWriter, r *http.Request, serviceName string) {
	if r.Method == http.MethodGet {
		groupName := r.URL.Query().Get("group")
		if groupName == "" {
			groupName = DefaultGroup
		}

		instances := s.registry.GetServiceInstances(serviceName, groupName)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service":   serviceName,
			"group":     groupName,
			"instances": instances,
		})
		return
	}

	if r.Method == http.MethodPost {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		inst := &Instance{
			ID:       req.ID,
			Host:     req.Host,
			Port:     req.Port,
			Tags:     req.Tags,
			Metadata: req.Metadata,
		}

		if err := s.registry.Register(serviceName, inst); err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "registered",
			"service":  serviceName,
			"instance": inst,
		})
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleServiceInstance(w http.ResponseWriter, r *http.Request, serviceName, instanceID string) {
	if r.Method == http.MethodDelete {
		if err := s.registry.Deregister(serviceName, instanceID); err != nil {
			http.Error(w, "not found: "+err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "deregistered",
			"service":  serviceName,
			"instance": instanceID,
		})
		return
	}

	if r.Method == http.MethodPost {
		action := r.URL.Query().Get("action")
		if action == "heartbeat" {
			if err := s.registry.Heartbeat(serviceName, instanceID); err != nil {
				http.Error(w, "not found: "+err.Error(), http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":   "heartbeat_updated",
				"service":  serviceName,
				"instance": instanceID,
			})
			return
		}

		if action == "force_deregister" {
			if err := s.registry.DeregisterUnhealthy(serviceName, instanceID); err != nil {
				http.Error(w, "not found: "+err.Error(), http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":   "force_deregistered",
				"service":  serviceName,
				"instance": instanceID,
			})
			return
		}

		http.Error(w, "bad request: invalid action", http.StatusBadRequest)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func getPort() int {
	if portStr := os.Getenv("PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
			return port
		}
	}
	return DefaultHTTPPort
}

func main() {
	registry := NewRegistry()
	server := NewServer(registry)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ticker := time.NewTicker(HeartbeatCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				registry.CheckTimeouts()
			}
		}
	}()

	port := getPort()
	addr := fmt.Sprintf(":%d", port)

	log.Printf("Service registry starting on %s", addr)
	log.Printf("Default group: %s", DefaultGroup)
	log.Printf("Heartbeat timeouts:")
	for name, cfg := range defaultGroupConfigs {
		log.Printf("  %s: %v", name, cfg.HeartbeatTimeout)
	}

	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
