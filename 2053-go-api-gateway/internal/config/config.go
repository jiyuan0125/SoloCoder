package config

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/example/apigateway/internal/model"
	"github.com/example/apigateway/internal/storage"
)

type ConfigManager struct {
	store           *storage.SQLiteStorage
	currentRoutes   atomic.Pointer[[]*model.RouteRule]
	routesMu        sync.RWMutex
	onRouteChange   func()
}

var configMgr *ConfigManager
var configOnce sync.Once

func GetConfigManager() *ConfigManager {
	configOnce.Do(func() {
		mgr := &ConfigManager{
			store: storage.GetStorage(),
		}
		mgr.reloadRoutes()
		configMgr = mgr
	})
	return configMgr
}

func (m *ConfigManager) SetOnRouteChange(fn func()) {
	m.onRouteChange = fn
}

func (m *ConfigManager) reloadRoutes() {
	routes, err := m.store.ListRoutes()
	if err != nil {
		return
	}
	m.currentRoutes.Store(&routes)
	if m.onRouteChange != nil {
		m.onRouteChange()
	}
}

func (m *ConfigManager) GetRoutes() []*model.RouteRule {
	ptr := m.currentRoutes.Load()
	if ptr == nil {
		return nil
	}
	return *ptr
}

func (m *ConfigManager) MatchRoute(path string, headers map[string]string) (*model.RouteRule, bool) {
	routes := m.GetRoutes()
	if routes == nil {
		return nil, false
	}

	var bestMatch *model.RouteRule
	bestMatchLen := -1

	for _, route := range routes {
		if !route.Healthy {
			continue
		}

		if !m.matchHeaders(route.Headers, headers) {
			continue
		}

		matched := false
		matchLen := 0

		switch route.MatchType {
		case model.MatchTypeExact:
			if path == route.Path {
				matched = true
				matchLen = len(route.Path)
			}
		case model.MatchTypePrefix:
			if strings.HasPrefix(path, route.Path) {
				matched = true
				matchLen = len(route.Path)
			}
		}

		if matched && matchLen > bestMatchLen {
			bestMatch = route
			bestMatchLen = matchLen
		}
	}

	return bestMatch, bestMatch != nil
}

func (m *ConfigManager) matchHeaders(routeHeaders, reqHeaders map[string]string) bool {
	if routeHeaders == nil || len(routeHeaders) == 0 {
		return true
	}

	for k, v := range routeHeaders {
		if reqHeaders[k] != v {
			return false
		}
	}
	return true
}

func (m *ConfigManager) CheckBackendHealth(targetURL string) bool {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	healthPath := parsed.Scheme + "://" + parsed.Host + "/health"

	req, err := http.NewRequestWithContext(ctx, "GET", healthPath, nil)
	if err != nil {
		return false
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode < 500
}

func (m *ConfigManager) CreateRouteChange(change *model.RouteChange) (int64, error) {
	id, err := m.store.CreateRouteChange(change)
	return id, err
}

func (m *ConfigManager) SubmitForApproval(changeID int64, applicant string) error {
	change, err := m.store.GetRouteChange(changeID)
	if err != nil {
		return err
	}
	if change == nil {
		return ErrNotFound
	}

	change.Status = model.StatusApproval1
	change.Applicant = applicant
	return m.store.UpdateRouteChange(change)
}

func (m *ConfigManager) Approval1(changeID int64, approver string, approved bool, reason string) error {
	change, err := m.store.GetRouteChange(changeID)
	if err != nil {
		return err
	}
	if change == nil {
		return ErrNotFound
	}
	if change.Status != model.StatusApproval1 {
		return ErrInvalidStatus
	}

	if !approved {
		change.Status = model.StatusRejected
		change.Reason = reason
		return m.store.UpdateRouteChange(change)
	}

	change.Status = model.StatusApproval2
	change.Approver1 = approver
	return m.store.UpdateRouteChange(change)
}

func (m *ConfigManager) Approval2(changeID int64, approver string, approved bool, reason string) error {
	change, err := m.store.GetRouteChange(changeID)
	if err != nil {
		return err
	}
	if change == nil {
		return ErrNotFound
	}
	if change.Status != model.StatusApproval2 {
		return ErrInvalidStatus
	}

	if !approved {
		change.Status = model.StatusRejected
		change.Reason = reason
		return m.store.UpdateRouteChange(change)
	}

	change.Status = model.StatusApproved
	change.Approver2 = approver
	return m.store.UpdateRouteChange(change)
}

func (m *ConfigManager) FinalApproveAndApply(changeID int64, approver string) error {
	change, err := m.store.GetRouteChange(changeID)
	if err != nil {
		return err
	}
	if change == nil {
		return ErrNotFound
	}
	if change.Status != model.StatusApproved {
		return ErrInvalidStatus
	}

	var newRule model.RouteRule
	if change.NewConfig != "" {
		if err := json.Unmarshal([]byte(change.NewConfig), &newRule); err != nil {
			return err
		}
	}

	var oldRule model.RouteRule
	if change.OldConfig != "" {
		json.Unmarshal([]byte(change.OldConfig), &oldRule)
	}

	m.routesMu.Lock()
	defer m.routesMu.Unlock()

	switch change.Operation {
	case "create":
		newRule.Healthy = m.CheckBackendHealth(newRule.TargetURL)
		id, err := m.store.CreateRoute(&newRule)
		if err != nil {
			return err
		}
		newRule.ID = id

	case "update":
		existing, err := m.store.GetRoute(change.RouteID)
		if err != nil {
			return err
		}
		if existing == nil {
			return ErrNotFound
		}
		newRule.ID = change.RouteID
		newRule.Healthy = m.CheckBackendHealth(newRule.TargetURL)
		if err := m.store.UpdateRoute(&newRule); err != nil {
			return err
		}

	case "delete":
		if err := m.store.DeleteRoute(change.RouteID); err != nil {
			return err
		}
	}

	now := time.Now()
	change.AppliedAt = &now
	change.FinalApprover = approver
	change.Status = model.StatusCompleted
	if err := m.store.UpdateRouteChange(change); err != nil {
		return err
	}

	m.reloadRoutes()
	return nil
}

func (m *ConfigManager) ListRoutes() ([]*model.RouteRule, error) {
	return m.store.ListRoutes()
}

func (m *ConfigManager) GetRoute(id int64) (*model.RouteRule, error) {
	return m.store.GetRoute(id)
}

func (m *ConfigManager) ListRouteChanges() ([]*model.RouteChange, error) {
	return m.store.ListRouteChanges()
}

func (m *ConfigManager) GetRouteChange(id int64) (*model.RouteChange, error) {
	return m.store.GetRouteChange(id)
}

func (m *ConfigManager) UpdateRouteChange(change *model.RouteChange) error {
	return m.store.UpdateRouteChange(change)
}

func (m *ConfigManager) CreateAPIKey(key *model.APIKey) (int64, error) {
	return m.store.CreateAPIKey(key)
}

func (m *ConfigManager) GetAPIKey(keyStr string) (*model.APIKey, error) {
	return m.store.GetAPIKey(keyStr)
}

func (m *ConfigManager) ListAPIKeys() ([]*model.APIKey, error) {
	return m.store.ListAPIKeys()
}

func (m *ConfigManager) DeleteAPIKey(id int64) error {
	return m.store.DeleteAPIKey(id)
}

func (m *ConfigManager) CreateJWTSecret(secret *model.JWTSecret) (int64, error) {
	return m.store.CreateJWTSecret(secret)
}

func (m *ConfigManager) GetJWTSecret() (*model.JWTSecret, error) {
	return m.store.GetJWTSecret()
}

var (
	ErrNotFound      = &ConfigError{Message: "not found"}
	ErrInvalidStatus = &ConfigError{Message: "invalid status"}
)

type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}
