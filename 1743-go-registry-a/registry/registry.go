package registry

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	HeartbeatTimeout    = 15 * time.Second
	UnhealthyTimeout    = 30 * time.Second
	HealthCheckInterval = 5 * time.Second
)

type ServiceStatus string

const (
	StatusHealthy   ServiceStatus = "healthy"
	StatusUnhealthy ServiceStatus = "unhealthy"
)

type ServiceInstance struct {
	IP       string        `json:"ip"`
	Port     int           `json:"port"`
	Name     string        `json:"name"`
	Version  string        `json:"version"`
	Weight   int           `json:"weight"`
	Status   ServiceStatus `json:"status"`
	LastHeartbeat time.Time `json:"-"`
	UnhealthySince time.Time `json:"-"`
}

type WatchEvent struct {
	Action    string           `json:"action"`
	Service   string           `json:"service"`
	Instances []ServiceInstance `json:"instances"`
}

type Registry struct {
	mu      sync.RWMutex
	services map[string]map[string]*ServiceInstance
	watchers map[string]map[chan WatchEvent]struct{}
}

func New() *Registry {
	r := &Registry{
		services: make(map[string]map[string]*ServiceInstance),
		watchers: make(map[string]map[chan WatchEvent]struct{}),
	}
	go r.healthCheckLoop()
	return r
}

func (r *Registry) Register(inst ServiceInstance) error {
	if inst.Name == "" || inst.IP == "" || inst.Port == 0 {
		return &APIError{Code: http.StatusBadRequest, Message: "name, ip, port are required"}
	}

	key := inst.IP + ":" + strconv.Itoa(inst.Port)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.services[inst.Name]; !ok {
		r.services[inst.Name] = make(map[string]*ServiceInstance)
	}

	existing, ok := r.services[inst.Name][key]
	if ok {
		existing.IP = inst.IP
		existing.Port = inst.Port
		existing.Name = inst.Name
		existing.Version = inst.Version
		existing.Weight = inst.Weight
		existing.Status = StatusHealthy
		existing.LastHeartbeat = time.Now()
		existing.UnhealthySince = time.Time{}
	} else {
		inst.Status = StatusHealthy
		inst.LastHeartbeat = time.Now()
		r.services[inst.Name][key] = &inst
	}

	r.notifyWatchers(inst.Name, "register")
	return nil
}

func (r *Registry) Heartbeat(name, ip string, port int) error {
	if name == "" || ip == "" || port == 0 {
		return &APIError{Code: http.StatusBadRequest, Message: "name, ip, port are required"}
	}

	key := ip + ":" + strconv.Itoa(port)

	r.mu.Lock()
	defer r.mu.Unlock()

	instances, ok := r.services[name]
	if !ok {
		return &APIError{Code: http.StatusNotFound, Message: "service not found"}
	}

	inst, ok := instances[key]
	if !ok {
		return &APIError{Code: http.StatusNotFound, Message: "instance not found, please register first"}
	}

	inst.LastHeartbeat = time.Now()
	wasUnhealthy := inst.Status == StatusUnhealthy
	inst.Status = StatusHealthy
	inst.UnhealthySince = time.Time{}

	if wasUnhealthy {
		r.notifyWatchers(name, "update")
	}

	return nil
}

func (r *Registry) Discover(name, version string) ([]ServiceInstance, error) {
	if name == "" {
		return nil, &APIError{Code: http.StatusBadRequest, Message: "name is required"}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	instances, ok := r.services[name]
	if !ok {
		return []ServiceInstance{}, nil
	}

	result := make([]ServiceInstance, 0)
	for _, inst := range instances {
		if inst.Status == StatusHealthy {
			if version == "" || inst.Version == version {
				result = append(result, *inst)
			}
		}
	}

	return result, nil
}

func (r *Registry) Watch(name string) (chan WatchEvent, func()) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.watchers[name]; !ok {
		r.watchers[name] = make(map[chan WatchEvent]struct{})
	}

	ch := make(chan WatchEvent, 10)
	r.watchers[name][ch] = struct{}{}

	go func() {
		r.mu.RLock()
		instances, ok := r.services[name]
		r.mu.RUnlock()

		if ok {
			healthy := make([]ServiceInstance, 0)
			for _, inst := range instances {
				if inst.Status == StatusHealthy {
					healthy = append(healthy, *inst)
				}
			}
			select {
			case ch <- WatchEvent{Action: "init", Service: name, Instances: healthy}:
			default:
			}
		}
	}()

	unsubscribe := func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if watchers, ok := r.watchers[name]; ok {
			delete(watchers, ch)
			close(ch)
		}
	}

	return ch, unsubscribe
}

func (r *Registry) healthCheckLoop() {
	ticker := time.NewTicker(HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		r.checkHealth()
	}
}

func (r *Registry) checkHealth() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	servicesToNotify := make(map[string]struct{})

	for serviceName, instances := range r.services {
		for key, inst := range instances {
			sinceLastHeartbeat := now.Sub(inst.LastHeartbeat)

			if inst.Status == StatusHealthy && sinceLastHeartbeat > HeartbeatTimeout {
				inst.Status = StatusUnhealthy
				inst.UnhealthySince = now
				servicesToNotify[serviceName] = struct{}{}
			} else if inst.Status == StatusUnhealthy {
				if !inst.UnhealthySince.IsZero() && now.Sub(inst.UnhealthySince) > UnhealthyTimeout {
					delete(instances, key)
					if len(instances) == 0 {
						delete(r.services, serviceName)
					}
					servicesToNotify[serviceName] = struct{}{}
				}
			}
		}
	}

	for serviceName := range servicesToNotify {
		r.notifyWatchersNoLock(serviceName, "update")
	}
}

func (r *Registry) notifyWatchers(serviceName, action string) {
	r.notifyWatchersNoLock(serviceName, action)
}

func (r *Registry) notifyWatchersNoLock(serviceName, action string) {
	watchers, ok := r.watchers[serviceName]
	if !ok {
		return
	}

	instances, _ := r.services[serviceName]
	healthy := make([]ServiceInstance, 0)
	for _, inst := range instances {
		if inst.Status == StatusHealthy {
			healthy = append(healthy, *inst)
		}
	}

	event := WatchEvent{
		Action:    action,
		Service:   serviceName,
		Instances: healthy,
	}

	for ch := range watchers {
		select {
		case ch <- event:
		default:
		}
	}
}

type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		json.NewEncoder(w).Encode(v)
	}
}

func WriteError(w http.ResponseWriter, err error) {
	if apiErr, ok := err.(*APIError); ok {
		WriteJSON(w, apiErr.Code, map[string]string{"error": apiErr.Message})
		return
	}
	WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}
