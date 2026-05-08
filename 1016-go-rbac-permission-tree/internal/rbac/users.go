package rbac

func (s *Store) AddRoleToUser(userID, roleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[roleID]; !exists {
		return ErrRoleNotFound
	}

	if s.userRoles[userID] == nil {
		s.userRoles[userID] = make(map[string]bool)
	}

	if s.userRoles[userID][roleID] {
		return nil
	}

	if s.checkUserMutexViolationLocked(userID, roleID) {
		return ErrMutexRoleConflict
	}

	s.userRoles[userID][roleID] = true

	s.cache.InvalidateUser(userID)

	return nil
}

func (s *Store) RemoveRoleFromUser(userID, roleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.userRoles[userID] == nil || !s.userRoles[userID][roleID] {
		return ErrRoleNotFound
	}

	delete(s.userRoles[userID], roleID)

	s.cache.InvalidateUser(userID)

	return nil
}

func (s *Store) GetUserRoles(userID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roles := make([]string, 0, len(s.userRoles[userID]))
	for role := range s.userRoles[userID] {
		roles = append(roles, role)
	}
	return roles
}
