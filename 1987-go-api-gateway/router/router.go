package router

import (
	"strings"
	"sync"
)

type RouteMatchType int

const (
	MatchExact RouteMatchType = iota
	MatchParam
	MatchPrefix
)

type BackendRef struct {
	Host string
	Port int
}

type Route struct {
	Path        string
	MatchType   RouteMatchType
	Backend     BackendRef
	Timeout     int
	PathSegments []string
	IsWildcard  bool
}

type MatchResult struct {
	Matched   bool
	Route     *Route
	Params    map[string]string
}

type Router struct {
	mu     sync.RWMutex
	routes []*Route
}

func NewRouter() *Router {
	return &Router{
		routes: make([]*Route, 0),
	}
}

func (r *Router) AddRoute(path string, backend BackendRef, timeout int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	matchType, segments, isWildcard := parsePath(path)

	route := &Route{
		Path:        path,
		MatchType:   matchType,
		Backend:     backend,
		Timeout:     timeout,
		PathSegments: segments,
		IsWildcard:  isWildcard,
	}

	r.routes = append(r.routes, route)
	r.sortRoutes()
}

func (r *Router) RemoveRoute(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, route := range r.routes {
		if route.Path == path {
			r.routes = append(r.routes[:i], r.routes[i+1:]...)
			break
		}
	}
}

func (r *Router) ListRoutes() []*Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make([]*Route, len(r.routes))
	copy(routes, r.routes)
	return routes
}

func (r *Router) Match(requestPath string) *MatchResult {
	r.mu.RLock()
	defer r.mu.RUnlock()

	requestSegments := splitPath(requestPath)

	for _, route := range r.routes {
		if match, params := matchRoute(route, requestSegments); match {
			return &MatchResult{
				Matched: true,
				Route:   route,
				Params:  params,
			}
		}
	}

	return &MatchResult{Matched: false}
}

func (r *Router) sortRoutes() {
	for i := range r.routes {
		for j := i + 1; j < len(r.routes); j++ {
			if compareRoutePriority(r.routes[i], r.routes[j]) > 0 {
				r.routes[i], r.routes[j] = r.routes[j], r.routes[i]
			}
		}
	}
}

func compareRoutePriority(a, b *Route) int {
	if a.MatchType != b.MatchType {
		return int(a.MatchType) - int(b.MatchType)
	}

	if len(a.PathSegments) != len(b.PathSegments) {
		return len(b.PathSegments) - len(a.PathSegments)
	}

	paramCountA, paramCountB := 0, 0
	for _, seg := range a.PathSegments {
		if isParamSegment(seg) {
			paramCountA++
		}
	}
	for _, seg := range b.PathSegments {
		if isParamSegment(seg) {
			paramCountB++
		}
	}

	return paramCountA - paramCountB
}

func parsePath(path string) (RouteMatchType, []string, bool) {
	segments := splitPath(path)
	matchType := MatchExact
	isWildcard := false

	for i, seg := range segments {
		if seg == "*" {
			if i == len(segments)-1 {
				isWildcard = true
				matchType = MatchPrefix
				break
			}
		} else if isParamSegment(seg) {
			matchType = MatchParam
		}
	}

	return matchType, segments, isWildcard
}

func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

func isParamSegment(seg string) bool {
	return len(seg) > 2 && seg[0] == '{' && seg[len(seg)-1] == '}'
}

func matchRoute(route *Route, requestSegments []string) (bool, map[string]string) {
	if route.IsWildcard {
		if len(requestSegments) < len(route.PathSegments)-1 {
			return false, nil
		}

		for i := 0; i < len(route.PathSegments)-1; i++ {
			routeSeg := route.PathSegments[i]
			reqSeg := requestSegments[i]

			if isParamSegment(routeSeg) {
				continue
			}
			if routeSeg != reqSeg {
				return false, nil
			}
		}

		params := make(map[string]string)
		for i := 0; i < len(route.PathSegments)-1; i++ {
			routeSeg := route.PathSegments[i]
			if isParamSegment(routeSeg) {
				paramName := routeSeg[1 : len(routeSeg)-1]
				params[paramName] = requestSegments[i]
			}
		}

		return true, params
	}

	if len(route.PathSegments) != len(requestSegments) {
		return false, nil
	}

	params := make(map[string]string)
	for i := range route.PathSegments {
		routeSeg := route.PathSegments[i]
		reqSeg := requestSegments[i]

		if isParamSegment(routeSeg) {
			paramName := routeSeg[1 : len(routeSeg)-1]
			params[paramName] = reqSeg
			continue
		}

		if routeSeg != reqSeg {
			return false, nil
		}
	}

	return true, params
}
