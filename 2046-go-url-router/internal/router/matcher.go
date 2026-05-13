package router

import (
	"net/url"
	"sort"
	"strings"
	"sync"

	"router/internal/model"
)

type Matcher struct {
	mu     sync.RWMutex
	routes []*model.Route
}

func NewMatcher() *Matcher {
	return &Matcher{}
}

func (m *Matcher) Update(routes []*model.Route) {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]*model.Route, len(routes))
	copy(copied, routes)
	sort.SliceStable(copied, func(i, j int) bool {
		return copied[i].Priority() > copied[j].Priority()
	})
	m.routes = copied
}

func (m *Matcher) Match(method, rawPath string) (*model.Route, bool, bool) {
	decoded, err := url.PathUnescape(rawPath)
	if err != nil {
		decoded = rawPath
	}
	decoded = cleanPath(decoded)

	m.mu.RLock()
	defer m.mu.RUnlock()

	var pathMatch bool
	var matchedRoute *model.Route

	for _, r := range m.routes {
		if matchesPath(r, decoded) {
			if r.Method == method {
				return r, true, true
			}
			pathMatch = true
			if matchedRoute == nil {
				matchedRoute = r
			}
		}
	}

	if pathMatch {
		return matchedRoute, true, false
	}
	return nil, false, false
}

func matchesPath(r *model.Route, decodedPath string) bool {
	if !r.IsWildcard() {
		return r.Path == decodedPath
	}
	prefix := r.WildcardPrefix()
	return strings.HasPrefix(decodedPath, prefix)
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	return p
}
