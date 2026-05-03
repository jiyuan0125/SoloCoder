package main

import (
	"sync"
)

type Backend struct {
	Addr          string
	Weight        int
	CurrentWeight int
	Breaker       *Breaker
}

type Balancer struct {
	mu sync.Mutex
}

func NewBalancer() *Balancer {
	return &Balancer{}
}

func (b *Balancer) Select(backends []*Backend) *Backend {
	b.mu.Lock()
	defer b.mu.Unlock()

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
		backend.CurrentWeight += backend.Weight
		if backend.CurrentWeight > maxWeight {
			maxWeight = backend.CurrentWeight
			selected = backend
		}
	}

	if selected != nil {
		selected.CurrentWeight -= totalWeight
	}

	return selected
}
