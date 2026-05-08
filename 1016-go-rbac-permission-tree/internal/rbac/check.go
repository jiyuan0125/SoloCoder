package rbac

type CheckResult struct {
	Allowed bool
	Scope   Scope
}

func (s *Store) CheckPermission(userID, resource, action string) CheckResult {
	permKey := resource + ":" + action

	if scope, cached := s.cache.Get(userID, permKey); cached {
		return CheckResult{Allowed: true, Scope: scope}
	}

	permissions := s.computeUserPermissions(userID)
	s.cache.Set(userID, permissions)

	scope, hasPerm := permissions[permKey]
	if hasPerm {
		return CheckResult{Allowed: true, Scope: scope}
	}

	return CheckResult{Allowed: false}
}

func (s *Store) computeUserPermissions(userID string) map[string]Scope {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userRoles := s.userRoles[userID]
	if len(userRoles) == 0 {
		return make(map[string]Scope)
	}

	allRoles := make(map[string]bool)
	for roleID := range userRoles {
		allRoles[roleID] = true
		for ancestor := range s.getAllAncestorsLocked(roleID) {
			allRoles[ancestor] = true
		}
	}

	permissions := make(map[string]Scope)
	for roleID := range allRoles {
		for permKey, perm := range s.rolePermissions[roleID] {
			if existing, hasPerm := permissions[permKey]; hasPerm {
				permissions[permKey] = existing.Max(perm.Scope)
			} else {
				permissions[permKey] = perm.Scope
			}
		}
	}

	return permissions
}

func (s *Store) GetUserAllPermissions(userID string) map[string]Scope {
	permissions := s.computeUserPermissions(userID)
	return permissions
}
