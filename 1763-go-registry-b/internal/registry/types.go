package registry

import (
	"time"
)

type ServiceInstance struct {
	ID          string            `json:"id"`
	ServiceName string            `json:"service_name"`
	IP          string            `json:"ip"`
	Port        int               `json:"port"`
	Metadata    map[string]string `json:"metadata"`
	Status      string            `json:"status"`
	LastHeartbeat time.Time       `json:"last_heartbeat"`
	RegisteredAt time.Time        `json:"registered_at"`
}

type ServiceInstanceInfo struct {
	ID          string            `json:"id"`
	IP          string            `json:"ip"`
	Port        int               `json:"port"`
	Metadata    map[string]string `json:"metadata"`
	Status      string            `json:"status"`
	LastHeartbeat time.Time       `json:"last_heartbeat"`
}

type ServiceInfo struct {
	ServiceName string              `json:"service_name"`
	Instances   []ServiceInstanceInfo `json:"instances"`
}

type StatusChange struct {
	ServiceName  string          `json:"service_name"`
	InstanceID   string          `json:"instance_id"`
	FromStatus   string          `json:"from_status"`
	ToStatus     string          `json:"to_status"`
	IP           string          `json:"ip"`
	Port         int             `json:"port"`
	Timestamp    time.Time       `json:"timestamp"`
}

const (
	StatusHealthy = "healthy"
	StatusUnhealthy = "unhealthy"
)
