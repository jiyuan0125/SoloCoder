package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultLeaseDuration = 30 * time.Second
	gracePeriod          = 5 * time.Second
	maxFailedHeartbeats  = 3
	randomIDLength       = 6
)

type InstanceStatus string

const (
	StatusHealthy   InstanceStatus = "healthy"
	StatusUnhealthy InstanceStatus = "unhealthy"
)

type Instance struct {
	ID             string            `json:"id"`
	ServiceName    string            `json:"service_name"`
	IP             string            `json:"ip"`
	Port           int               `json:"port"`
	Tags           map[string]string `json:"tags"`
	Status         InstanceStatus    `json:"status"`
	LeaseExpiresAt time.Time         `json:"lease_expires_at"`
	FailedCount    int               `json:"-"`
	LastHeartbeat  time.Time         `json:"-"`
}

type ServiceStats struct {
	TotalCount     int `json:"total_count"`
	HealthyCount   int `json:"healthy_count"`
	UnhealthyCount int `json:"unhealthy_count"`
}

type Registry struct {
	mu        sync.RWMutex
	instances map[string]*Instance
	stats     map[string]*ServiceStats
	subsMu    sync.RWMutex
	subs      map[string][]chan<- string
}

func NewRegistry() *Registry {
	r := &Registry{
		instances: make(map[string]*Instance),
		stats:     make(map[string]*ServiceStats),
		subs:      make(map[string][]chan<- string),
	}
	go r.cleanupLoop()
	return r
}

func generateInstanceID(serviceName string) string {
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	randomBytes := make([]byte, randomIDLength)
	rand.Read(randomBytes)
	randomStr := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%s-%s-%s", serviceName, timestamp, randomStr)
}

func tagsToString(tags map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range tags {
		switch val := v.(type) {
		case string:
			result[k] = val
		case float64:
			if val == math.Trunc(val) {
				result[k] = strconv.FormatInt(int64(val), 10)
			} else {
				result[k] = strconv.FormatFloat(val, 'f', -1, 64)
			}
		default:
			result[k] = fmt.Sprintf("%v", val)
		}
	}
	return result
}

func (r *Registry) Register(serviceName, ip string, port int, tags map[string]string) (string, time.Duration) {
	instanceID := generateInstanceID(serviceName)
	instance := &Instance{
		ID:             instanceID,
		ServiceName:    serviceName,
		IP:             ip,
		Port:           port,
		Tags:           tags,
		Status:         StatusHealthy,
		LeaseExpiresAt: time.Now().Add(defaultLeaseDuration),
		FailedCount:    0,
		LastHeartbeat:  time.Now(),
	}

	r.mu.Lock()
	r.instances[instanceID] = instance
	r.updateStatsLocked(serviceName, true)
	r.mu.Unlock()

	return instanceID, defaultLeaseDuration
}

func (r *Registry) Heartbeat(instanceID string) (string, time.Time, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance, exists := r.instances[instanceID]
	if !exists {
		return "", time.Time{}, false
	}

	now := time.Now()
	instance.LastHeartbeat = now

	if instance.Status == StatusUnhealthy {
		instance.Status = StatusHealthy
		instance.FailedCount = 0
		r.updateStatsLocked(instance.ServiceName, true)
	}

	instance.LeaseExpiresAt = now.Add(defaultLeaseDuration)
	return instanceID, instance.LeaseExpiresAt, true
}

func (r *Registry) updateStatsLocked(serviceName string, added bool) {
	stats, exists := r.stats[serviceName]
	if !exists {
		stats = &ServiceStats{}
		r.stats[serviceName] = stats
	}

	if added {
		stats.TotalCount++
		stats.HealthyCount++
	}
}

func (r *Registry) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		r.checkExpiredInstances()
	}
}

func (r *Registry) checkExpiredInstances() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	var toRemove []string

	for id, instance := range r.instances {
		if now.After(instance.LeaseExpiresAt) {
			instance.FailedCount++

			if instance.FailedCount < maxFailedHeartbeats {
				if instance.Status == StatusHealthy {
					instance.Status = StatusUnhealthy
					stats := r.stats[instance.ServiceName]
					stats.HealthyCount--
					stats.UnhealthyCount++
				}
				instance.LeaseExpiresAt = now.Add(defaultLeaseDuration)
			} else {
				toRemove = append(toRemove, id)
			}
		} else if now.After(instance.LeaseExpiresAt.Add(-gracePeriod)) {
			if instance.Status == StatusHealthy {
				instance.Status = StatusUnhealthy
				stats := r.stats[instance.ServiceName]
				stats.HealthyCount--
				stats.UnhealthyCount++
			}
		}
	}

	for _, id := range toRemove {
		instance := r.instances[id]
		stats := r.stats[instance.ServiceName]
		stats.TotalCount--
		if instance.Status == StatusHealthy {
			stats.HealthyCount--
		} else {
			stats.UnhealthyCount--
		}
		if stats.TotalCount == 0 {
			delete(r.stats, instance.ServiceName)
		}
		delete(r.instances, id)
		r.notifySubscribers(id)
	}
}

func (r *Registry) notifySubscribers(instanceID string) {
	r.subsMu.RLock()
	defer r.subsMu.RUnlock()

	for _, ch := range r.subs[instanceID] {
		select {
		case ch <- instanceID:
		default:
		}
	}
}

func (r *Registry) Query(serviceName string, tagFilters map[string]string, includeUnhealthy bool) []*Instance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*Instance
	for _, instance := range r.instances {
		if instance.ServiceName != serviceName {
			continue
		}
		if !includeUnhealthy && instance.Status != StatusHealthy {
			continue
		}

		match := true
		for k, v := range tagFilters {
			if instanceVal, exists := instance.Tags[k]; !exists || instanceVal != v {
				match = false
				break
			}
		}
		if match {
			results = append(results, copyInstance(instance))
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})

	return results
}

func copyInstance(src *Instance) *Instance {
	tags := make(map[string]string)
	for k, v := range src.Tags {
		tags[k] = v
	}
	return &Instance{
		ID:             src.ID,
		ServiceName:    src.ServiceName,
		IP:             src.IP,
		Port:           src.Port,
		Tags:           tags,
		Status:         src.Status,
		LeaseExpiresAt: src.LeaseExpiresAt,
	}
}

func (r *Registry) GetStats() map[string]*ServiceStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*ServiceStats)
	for name, stats := range r.stats {
		result[name] = &ServiceStats{
			TotalCount:     stats.TotalCount,
			HealthyCount:   stats.HealthyCount,
			UnhealthyCount: stats.UnhealthyCount,
		}
	}
	return result
}

type RegisterRequest struct {
	ServiceName string                 `json:"service_name"`
	IP          string                 `json:"ip"`
	Port        int                    `json:"port"`
	Tags        map[string]interface{} `json:"tags"`
}

type RegisterResponse struct {
	InstanceID    string `json:"instance_id"`
	LeaseDuration string `json:"lease_duration"`
}

type HeartbeatResponse struct {
	InstanceID     string `json:"instance_id"`
	NextHeartbeat  string `json:"next_heartbeat"`
}

type QueryResponse struct {
	Instances []*Instance `json:"instances"`
}

type StatsResponse struct {
	Services map[string]*ServiceStats `json:"services"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func main() {
	registry := NewRegistry()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8909"
	}

	http.HandleFunc("/register", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var body RegisterRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if body.ServiceName == "" || body.IP == "" || body.Port <= 0 || body.Port > 65535 {
			writeError(w, http.StatusBadRequest, "invalid service_name, ip or port")
			return
		}

		tags := tagsToString(body.Tags)
		instanceID, leaseDuration := registry.Register(body.ServiceName, body.IP, body.Port, tags)

		writeJSON(w, http.StatusOK, RegisterResponse{
			InstanceID:    instanceID,
			LeaseDuration: leaseDuration.String(),
		})
	})

	http.HandleFunc("/heartbeat", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost && req.Method != http.MethodPut {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		instanceID := req.URL.Query().Get("instance_id")
		if instanceID == "" {
			writeError(w, http.StatusBadRequest, "missing instance_id")
			return
		}

		id, nextHeartbeat, ok := registry.Heartbeat(instanceID)
		if !ok {
			writeError(w, http.StatusNotFound, "instance not found")
			return
		}

		writeJSON(w, http.StatusOK, HeartbeatResponse{
			InstanceID:    id,
			NextHeartbeat: nextHeartbeat.Format(time.RFC3339),
		})
	})

	http.HandleFunc("/query", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		serviceName := req.URL.Query().Get("service_name")
		if serviceName == "" {
			writeError(w, http.StatusBadRequest, "missing service_name")
			return
		}

		includeUnhealthy := req.URL.Query().Get("include_unhealthy") == "true"

		tagFilters := make(map[string]string)
		for k, v := range req.URL.Query() {
			if strings.HasPrefix(k, "tag_") && len(v) > 0 {
				tagFilters[k[4:]] = v[0]
			}
		}

		instances := registry.Query(serviceName, tagFilters, includeUnhealthy)
		writeJSON(w, http.StatusOK, QueryResponse{Instances: instances})
	})

	http.HandleFunc("/stats", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		stats := registry.GetStats()
		writeJSON(w, http.StatusOK, StatsResponse{Services: stats})
	})

	fmt.Printf("Service registry starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
