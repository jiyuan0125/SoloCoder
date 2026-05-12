package registry

import (
	"sync"
	"time"
)

const (
	defaultCheckInterval = 5 * time.Second
	defaultTTL           = 15 * time.Second
)

type Registry struct {
	services      map[string]map[string]*ServiceInstance
	lock          sync.RWMutex
	checkInterval time.Duration
	ttl           time.Duration
	subscribers   map[chan StatusChange]bool
	subscriberMu  sync.RWMutex
}

func New() *Registry {
	r := &Registry{
		services:      make(map[string]map[string]*ServiceInstance),
		checkInterval: defaultCheckInterval,
		ttl:           defaultTTL,
		subscribers:   make(map[chan StatusChange]bool),
	}
	go r.healthCheckLoop()
	return r
}

func (r *Registry) Register(instance *ServiceInstance) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	now := time.Now()
	instance.Status = StatusHealthy
	instance.LastHeartbeat = now
	instance.RegisteredAt = now

	if _, exists := r.services[instance.ServiceName]; !exists {
		r.services[instance.ServiceName] = make(map[string]*ServiceInstance)
	}

	oldInstance, oldExists := r.services[instance.ServiceName][instance.ID]
	r.services[instance.ServiceName][instance.ID] = instance

	if !oldExists {
		r.notifyChange(instance.ServiceName, instance.ID, "", StatusHealthy, instance.IP, instance.Port)
	} else if oldInstance.Status != StatusHealthy {
		r.notifyChange(instance.ServiceName, instance.ID, oldInstance.Status, StatusHealthy, instance.IP, instance.Port)
	}

	return nil
}

func (r *Registry) Heartbeat(serviceName, instanceID string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	instances, exists := r.services[serviceName]
	if !exists {
		return ErrServiceNotFound
	}

	instance, exists := instances[instanceID]
	if !exists {
		return ErrInstanceNotFound
	}

	oldStatus := instance.Status
	instance.LastHeartbeat = time.Now()

	if oldStatus != StatusHealthy {
		instance.Status = StatusHealthy
		r.notifyChange(serviceName, instanceID, oldStatus, StatusHealthy, instance.IP, instance.Port)
	}

	return nil
}

func (r *Registry) Unregister(serviceName, instanceID string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	instances, exists := r.services[serviceName]
	if !exists {
		return ErrServiceNotFound
	}

	instance, exists := instances[instanceID]
	if !exists {
		return ErrInstanceNotFound
	}

	oldStatus := instance.Status
	delete(instances, instanceID)

	if len(instances) == 0 {
		delete(r.services, serviceName)
	}

	if oldStatus == StatusHealthy {
		r.notifyChange(serviceName, instanceID, StatusHealthy, "removed", instance.IP, instance.Port)
	}

	return nil
}

func (r *Registry) GetService(serviceName string) ([]ServiceInstance, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	instances, exists := r.services[serviceName]
	if !exists {
		return nil, ErrServiceNotFound
	}

	result := make([]ServiceInstance, 0, len(instances))
	for _, inst := range instances {
		if inst.Status == StatusHealthy {
			result = append(result, *inst)
		}
	}

	return result, nil
}

func (r *Registry) GetAllServices() []ServiceInfo {
	r.lock.RLock()
	defer r.lock.RUnlock()

	result := make([]ServiceInfo, 0, len(r.services))
	for serviceName, instances := range r.services {
		info := ServiceInfo{
			ServiceName: serviceName,
			Instances:   make([]ServiceInstanceInfo, 0, len(instances)),
		}
		for _, inst := range instances {
			info.Instances = append(info.Instances, ServiceInstanceInfo{
				ID:            inst.ID,
				IP:            inst.IP,
				Port:          inst.Port,
				Metadata:      inst.Metadata,
				Status:        inst.Status,
				LastHeartbeat: inst.LastHeartbeat,
			})
		}
		result = append(result, info)
	}

	return result
}

func (r *Registry) Subscribe() chan StatusChange {
	ch := make(chan StatusChange, 100)
	r.subscriberMu.Lock()
	r.subscribers[ch] = true
	r.subscriberMu.Unlock()
	return ch
}

func (r *Registry) Unsubscribe(ch chan StatusChange) {
	r.subscriberMu.Lock()
	delete(r.subscribers, ch)
	close(ch)
	r.subscriberMu.Unlock()
}

func (r *Registry) healthCheckLoop() {
	ticker := time.NewTicker(r.checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		r.checkHealth()
	}
}

func (r *Registry) checkHealth() {
	r.lock.Lock()
	defer r.lock.Unlock()

	now := time.Now()
	for serviceName, instances := range r.services {
		for instanceID, instance := range instances {
			if now.Sub(instance.LastHeartbeat) > r.ttl {
				if instance.Status == StatusHealthy {
					oldStatus := instance.Status
					instance.Status = StatusUnhealthy
					r.notifyChange(serviceName, instanceID, oldStatus, StatusUnhealthy, instance.IP, instance.Port)
				}
			}
		}
	}
}

func (r *Registry) notifyChange(serviceName, instanceID, fromStatus, toStatus, ip string, port int) {
	change := StatusChange{
		ServiceName: serviceName,
		InstanceID:  instanceID,
		FromStatus:  fromStatus,
		ToStatus:    toStatus,
		IP:          ip,
		Port:        port,
		Timestamp:   time.Now(),
	}

	r.subscriberMu.RLock()
	defer r.subscriberMu.RUnlock()

	for ch := range r.subscribers {
		select {
		case ch <- change:
		default:
		}
	}
}
