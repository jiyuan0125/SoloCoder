package main

import (
	"strings"
	"sync"
)

type Route struct {
	ID       string
	Path     string
	Backends []string
}

type RouteManager struct {
	mu    sync.RWMutex
	routes map[string]*Route
}

func NewRouteManager() *RouteManager {
	return &RouteManager{
		routes: make(map[string]*Route),
	}
}

func (rm *RouteManager) Add(path string, backends []string) *Route {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	id := GenerateKeyID()
	route := &Route{
		ID:       id,
		Path:     path,
		Backends: backends,
	}
	rm.routes[id] = route
	return route
}

func (rm *RouteManager) Delete(id string) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, ok := rm.routes[id]; ok {
		delete(rm.routes, id)
		return true
	}
	return false
}

func (rm *RouteManager) Get(id string) *Route {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return rm.routes[id]
}

func (rm *RouteManager) Match(path string) *Route {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var best *Route
	bestLen := 0

	for _, r := range rm.routes {
		if strings.HasPrefix(path, r.Path) && len(r.Path) > bestLen {
			best = r
			bestLen = len(r.Path)
		}
	}

	return best
}

func (rm *RouteManager) List() []*Route {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make([]*Route, 0, len(rm.routes))
	for _, r := range rm.routes {
		result = append(result, r)
	}
	return result
}
