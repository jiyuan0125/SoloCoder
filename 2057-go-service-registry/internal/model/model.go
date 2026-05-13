package model

import (
	"time"
)

type ServiceInstance struct {
	ID          string
	ServiceName string
	Address     string
	Port        int
	Version     string
	Weight      int
	Environment string
	Metadata    map[string]string
	Status      InstanceStatus
	LastHeartbeat time.Time
	CreatedAt   time.Time
}

type InstanceStatus string

const (
	StatusHealthy   InstanceStatus = "healthy"
	StatusUnhealthy InstanceStatus = "unhealthy"
)

type ServiceEvent struct {
	Type        EventType
	Instance    *ServiceInstance
	Timestamp   time.Time
}

type EventType string

const (
	EventRegistered   EventType = "registered"
	EventUnregistered EventType = "unregistered"
	EventHealthy    EventType = "healthy"
	EventUnhealthy  EventType = "unhealthy"
)

type ResourceType string

const (
	ResourceService  ResourceType = "service"
	ResourceInstance ResourceType = "instance"
)

type Resource struct {
	ID   string
	Type ResourceType
	Name string
}

type ResourceAssociation struct {
	ResourceID   string
	TargetID   string
	TargetType ResourceType
	Operation  string
	Timestamp  time.Time
}

type LoadBalanceStrategy string

const (
	StrategyRoundRobin     LoadBalanceStrategy = "round_robin"
	StrategyWeightedRandom LoadBalanceStrategy = "weighted_random"
)

const (
	UnhealthyThreshold = 15 * time.Second
	TimeoutThreshold    = 30 * time.Second
)
