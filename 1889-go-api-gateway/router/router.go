package router

import (
	"strings"
	"sync"
)

type AuthStrategy string

const (
	AuthNone   AuthStrategy = "none"
	AuthAPIKey AuthStrategy = "apikey"
	AuthJWT    AuthStrategy = "jwt"
)

type Route struct {
	Path        string
	BackendURL  string
	Auth        AuthStrategy
	MatchType   string
}

type Table struct {
	mu           sync.RWMutex
	exact        map[string]*Route
	prefixes     map[string]*Route
	backendHosts map[string]bool
}

func NewTable() *Table {
	return &Table{
		exact:        make(map[string]*Route),
		prefixes:     make(map[string]*Route),
		backendHosts: make(map[string]bool),
	}
}

func (t *Table) Add(path string, backend string, auth AuthStrategy, isPrefix bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	route := &Route{
		Path:       path,
		BackendURL: backend,
		Auth:       auth,
	}
	if isPrefix {
		route.MatchType = "prefix"
		t.prefixes[path] = route
	} else {
		route.MatchType = "exact"
		t.exact[path] = route
	}
	if u := parseHost(backend); u != "" {
		t.backendHosts[u] = true
	}
}

func (t *Table) Match(path string) *Route {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if r, ok := t.exact[path]; ok {
		return r
	}
	var best *Route
	for p, r := range t.prefixes {
		if strings.HasPrefix(path, p) {
			if best == nil || len(p) > len(best.Path) {
				best = r
			}
		}
	}
	return best
}

func (t *Table) IsBackendURL(url string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	host := parseHost(url)
	return t.backendHosts[host]
}

func parseHost(raw string) string {
	parts := strings.SplitN(raw, "://", 2)
	if len(parts) != 2 {
		return ""
	}
	hostPart := parts[1]
	if idx := strings.Index(hostPart, "/"); idx >= 0 {
		hostPart = hostPart[:idx]
	}
	return parts[0] + "://" + hostPart
}
