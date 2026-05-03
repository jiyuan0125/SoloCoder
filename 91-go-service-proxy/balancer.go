package main

import (
	"sync"
)

type Backend struct {
	Addr          string
	Weight        int
	CurrentWeight int
	Breaker       *Breaker
	mu            sync.Mutex
}

type Balancer struct{}

func NewBalancer() *Balancer {
	return &Balancer{}
}

func (b *Balancer) Select(backends []*Backend) *Backend {
	var availableBackends []*Backend
	totalWeight := 0

	for _, backend := range backends {
		if backend.Breaker.Allow() {
			availableBackends = append(availableBackends, backend)
			totalWeight += backend.Weight
		}
	}

	if len(availableBackends) == 0 {
		return nil
	}

	if totalWeight == 0 {
		return availableBackends[0]
	}

	var selected *Backend
	maxWeight := -1

	for _, backend := range availableBackends {
		backend.mu.Lock()
		backend.CurrentWeight += backend.Weight
		if backend.CurrentWeight > maxWeight {
			maxWeight = backend.CurrentWeight
			selected = backend
		}
		backend.mu.Unlock()
	}

	if selected != nil {
		selected.mu.Lock()
		selected.CurrentWeight -= totalWeight
		selected.mu.Unlock()
	}

	return selected
}
