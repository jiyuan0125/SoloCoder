package main

import (
	"net/http"
	"sort"
)

type Router struct {
	routes []*routeEntry
}

type routeEntry struct {
	prefix   string
	backends []*Backend
}

func NewRouter(configRoutes []RouteConfig) *Router {
	r := &Router{
		routes: make([]*routeEntry, 0, len(configRoutes)),
	}

	for _, rc := range configRoutes {
		backends := make([]*Backend, 0, len(rc.Backends))
		for _, bc := range rc.Backends {
			weight := bc.Weight
			if weight <= 0 {
				weight = 1
			}
			backends = append(backends, &Backend{
				Addr:          bc.Addr,
				Weight:        weight,
				CurrentWeight: 0,
				Breaker:       NewBreaker(),
			})
		}

		r.routes = append(r.routes, &routeEntry{
			prefix:   rc.PathPrefix,
			backends: backends,
		})
	}

	sort.Slice(r.routes, func(i, j int) bool {
		return len(r.routes[i].prefix) > len(r.routes[j].prefix)
	})

	return r
}

func (r *Router) Match(req *http.Request) ([]*Backend, string) {
	path := req.URL.Path
	for _, route := range r.routes {
		if len(path) >= len(route.prefix) && path[:len(route.prefix)] == route.prefix {
			return route.backends, route.prefix
		}
	}
	return nil, ""
}
