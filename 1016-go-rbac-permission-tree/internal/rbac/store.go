package rbac

import (
	"sync"
	"time"
)

type Store struct {
	mu sync.RWMutex

	roles          map[string]*Role
	rolePermissions map[string]map[string]Permission

	parents  map[string]map[string]bool
	children map[string]map[string]bool

	mutexRoles map[string]map[string]bool

	userRoles map[string]map[string]bool

	cache *PermissionCache
}

func NewStore() *Store {
	return &Store{
		roles:           make(map[string]*Role),
		rolePermissions: make(map[string]map[string]Permission),
		parents:         make(map[string]map[string]bool),
		children:        make(map[string]map[string]bool),
		mutexRoles:      make(map[string]map[string]bool),
		userRoles:       make(map[string]map[string]bool),
		cache:           NewPermissionCache(),
	}
}

func (s *Store) CreateRole(id, name, description string) (*Role, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[id]; exists {
		return nil, ErrRoleAlreadyExists
	}

	role := &Role{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.roles[id] = role
	s.rolePermissions[id] = make(map[string]Permission)
	s.parents[id] = make(map[string]bool)
	s.children[id] = make(map[string]bool)

	return role, nil
}

func (s *Store) GetRole(id string) (*Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	role, exists := s.roles[id]
	if !exists {
		return nil, ErrRoleNotFound
	}
	return role, nil
}

func (s *Store) ListRoles() []*Role {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roles := make([]*Role, 0, len(s.roles))
	for _, role := range s.roles {
		roles = append(roles, role)
	}
	return roles
}

func (s *Store) UpdateRole(id, name, description string) (*Role, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	role, exists := s.roles[id]
	if !exists {
		return nil, ErrRoleNotFound
	}

	role.Name = name
	role.Description = description
	role.UpdatedAt = time.Now()

	return role, nil
}

func (s *Store) DeleteRole(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[id]; !exists {
		return ErrRoleNotFound
	}

	affectedUsers := s.getUsersWithRoleLocked(id)

	for parent := range s.parents[id] {
		delete(s.children[parent], id)
	}
	for child := range s.children[id] {
		delete(s.parents[child], id)
	}

	delete(s.parents, id)
	delete(s.children, id)

	for otherID := range s.mutexRoles {
		delete(s.mutexRoles[otherID], id)
	}
	delete(s.mutexRoles, id)

	for userID := range s.userRoles {
		delete(s.userRoles[userID], id)
	}

	delete(s.rolePermissions, id)
	delete(s.roles, id)

	s.cache.InvalidateRole(id)
	for _, userID := range affectedUsers {
		s.cache.InvalidateUser(userID)
	}

	return nil
}

func (s *Store) getUsersWithRoleLocked(roleID string) []string {
	var users []string
	for userID, roles := range s.userRoles {
		if roles[roleID] {
			users = append(users, userID)
		}
	}
	return users
}
