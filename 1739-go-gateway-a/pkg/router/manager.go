package router

import (
	"fmt"
	"net/http/httputil"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gateway/pkg/models"
	"gateway/pkg/plugins"
)

type RouteEntry struct {
	models.Route
	Plugins  []plugins.Plugin
	Stats    *models.RouteStats
	Proxy    *httputil.ReverseProxy
}

type Manager struct {
	sync.RWMutex
	routes    map[string]*RouteEntry
	versions  map[string][]models.RouteVersion
	patterns  []*patternEntry
}

type patternEntry struct {
	id       string
	path     string
	methods  map[string]struct{}
	priority int
}

func NewManager() *Manager {
	return &Manager{
		routes:   make(map[string]*RouteEntry),
		versions: make(map[string][]models.RouteVersion),
	}
}

func (m *Manager) Create(r *models.Route) (*RouteEntry, error) {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if len(r.Methods) == 0 {
		r.Methods = []string{"GET"}
	}
	if r.Plugins == nil {
		r.Plugins = []models.PluginConfig{}
	}
	now := time.Now()
	r.CreatedAt = now
	r.UpdatedAt = now
	r.Version = 1

	m.Lock()
	defer m.Unlock()

	if _, exists := m.routes[r.ID]; exists {
		return nil, fmt.Errorf("route %s already exists", r.ID)
	}

	entry, err := m.buildEntry(r)
	if err != nil {
		return nil, err
	}

	m.routes[r.ID] = entry
	m.versions[r.ID] = []models.RouteVersion{
		{Route: *r, VersionTime: now},
	}
	m.rebuildPatterns()
	return entry, nil
}

func (m *Manager) Update(id string, r *models.Route) (*RouteEntry, error) {
	m.Lock()
	defer m.Unlock()

	old, exists := m.routes[id]
	if !exists {
		return nil, fmt.Errorf("route %s not found", id)
	}

	newRoute := old.Route
	if r.Path != "" {
		newRoute.Path = r.Path
	}
	if len(r.Methods) > 0 {
		newRoute.Methods = r.Methods
	}
	if r.Upstream != "" {
		newRoute.Upstream = r.Upstream
	}
	if r.Plugins != nil {
		newRoute.Plugins = r.Plugins
	}
	newRoute.Enabled = r.Enabled
	newRoute.UpdatedAt = time.Now()
	newRoute.Version = old.Version + 1

	stats := old.Stats
	oldVersion := old.Route
	m.versions[id] = append(m.versions[id], models.RouteVersion{
		Route:       oldVersion,
		VersionTime: time.Now(),
	})

	newEntry, err := m.buildEntry(&newRoute)
	if err != nil {
		return nil, err
	}
	newEntry.Stats = stats
	m.routes[id] = newEntry
	m.rebuildPatterns()
	return newEntry, nil
}

func (m *Manager) Delete(id string) error {
	m.Lock()
	defer m.Unlock()

	if _, exists := m.routes[id]; !exists {
		return fmt.Errorf("route %s not found", id)
	}
	delete(m.routes, id)
	delete(m.versions, id)
	m.rebuildPatterns()
	return nil
}

func (m *Manager) Get(id string) (*RouteEntry, bool) {
	m.RLock()
	defer m.RUnlock()
	entry, ok := m.routes[id]
	return entry, ok
}

func (m *Manager) List() []*RouteEntry {
	m.RLock()
	defer m.RUnlock()
	list := make([]*RouteEntry, 0, len(m.routes))
	for _, e := range m.routes {
		list = append(list, e)
	}
	return list
}

func (m *Manager) Versions(id string) []models.RouteVersion {
	m.RLock()
	defer m.RUnlock()
	vers := append([]models.RouteVersion{}, m.versions[id]...)
	sort.Slice(vers, func(i, j int) bool {
		return vers[i].Version > vers[j].Version
	})
	return vers
}

func (m *Manager) Rollback(id string, version int) (*RouteEntry, error) {
	m.Lock()
	defer m.Unlock()

	current, exists := m.routes[id]
	if !exists {
		return nil, fmt.Errorf("route %s not found", id)
	}

	versions := m.versions[id]
	var target *models.RouteVersion
	for i := range versions {
		if versions[i].Version == version {
			target = &versions[i]
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("version %d not found for route %s", version, id)
	}

	m.versions[id] = append(m.versions[id], models.RouteVersion{
		Route:       current.Route,
		VersionTime: time.Now(),
	})

	newRoute := target.Route
	newRoute.Version = current.Version + 1
	newRoute.UpdatedAt = time.Now()

	stats := current.Stats
	newEntry, err := m.buildEntry(&newRoute)
	if err != nil {
		return nil, err
	}
	newEntry.Stats = stats
	m.routes[id] = newEntry
	m.rebuildPatterns()
	return newEntry, nil
}

func (m *Manager) Match(method, path string) (*RouteEntry, map[string]string) {
	m.RLock()
	defer m.RUnlock()

	for _, p := range m.patterns {
		if _, methodOk := p.methods[method]; !methodOk {
			continue
		}
		if params, matched := matchPattern(p.path, path); matched {
			entry := m.routes[p.id]
			if entry != nil && entry.Enabled {
				return entry, params
			}
		}
	}
	return nil, nil
}

func (m *Manager) buildEntry(r *models.Route) (*RouteEntry, error) {
	entry := &RouteEntry{Route: *r}

	if r.Enabled {
		entry.Plugins = plugins.BuildPlugins(r.Plugins)

		u, err := url.Parse(r.Upstream)
		if err != nil {
			return nil, fmt.Errorf("invalid upstream url: %w", err)
		}
		entry.Proxy = httputil.NewSingleHostReverseProxy(u)
	}

	if entry.Stats == nil {
		entry.Stats = &models.RouteStats{StartTime: time.Now()}
	}
	return entry, nil
}

func (m *Manager) rebuildPatterns() {
	patterns := make([]*patternEntry, 0, len(m.routes))
	for id, entry := range m.routes {
		methods := make(map[string]struct{})
		for _, m := range entry.Methods {
			methods[strings.ToUpper(m)] = struct{}{}
		}
		patterns = append(patterns, &patternEntry{
			id:       id,
			path:     entry.Path,
			methods:  methods,
			priority: patternPriority(entry.Path),
		})
	}

	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].priority < patterns[j].priority
	})
	m.patterns = patterns
}

func patternPriority(path string) int {
	if strings.Contains(path, "*") {
		return 100
	}
	if strings.Contains(path, ":") {
		return 50
	}
	return 0
}

func matchPattern(pattern, path string) (map[string]string, bool) {
	patternParts := splitPath(pattern)
	pathParts := splitPath(path)

	params := make(map[string]string)
	wildcard := false
	var wildcardStart int

	for i, p := range patternParts {
		if p == "*" {
			wildcard = true
			wildcardStart = i
			break
		}
		if i >= len(pathParts) {
			return nil, false
		}
		if strings.HasPrefix(p, ":") {
			params[p[1:]] = pathParts[i]
		} else if p != pathParts[i] {
			return nil, false
		}
	}

	if wildcard {
		if len(pathParts) >= wildcardStart {
			params["*"] = strings.Join(pathParts[wildcardStart:], "/")
			return params, true
		}
		return nil, false
	}

	if len(patternParts) != len(pathParts) {
		return nil, false
	}
	return params, true
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return []string{}
	}
	return strings.Split(p, "/")
}
