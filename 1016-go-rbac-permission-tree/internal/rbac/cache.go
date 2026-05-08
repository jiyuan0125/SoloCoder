package rbac

import "sync"

type userPermissionCache struct {
	mu sync.RWMutex
	permissions map[string]Scope
}

type PermissionCache struct {
	mu        sync.RWMutex
	userCache map[string]*userPermissionCache
}

func NewPermissionCache() *PermissionCache {
	return &PermissionCache{
		userCache: make(map[string]*userPermissionCache),
	}
}

func (c *PermissionCache) Get(userID, permKey string) (Scope, bool) {
	c.mu.RLock()
	uc, exists := c.userCache[userID]
	c.mu.RUnlock()

	if !exists {
		return "", false
	}

	uc.mu.RLock()
	defer uc.mu.RUnlock()
	scope, hasPerm := uc.permissions[permKey]
	return scope, hasPerm
}

func (c *PermissionCache) Set(userID string, permissions map[string]Scope) {
	uc := &userPermissionCache{
		permissions: permissions,
	}

	c.mu.Lock()
	c.userCache[userID] = uc
	c.mu.Unlock()
}

func (c *PermissionCache) InvalidateUser(userID string) {
	c.mu.Lock()
	delete(c.userCache, userID)
	c.mu.Unlock()
}

func (c *PermissionCache) InvalidateRole(roleID string) {
}

func (c *PermissionCache) Clear() {
	c.mu.Lock()
	c.userCache = make(map[string]*userPermissionCache)
	c.mu.Unlock()
}
