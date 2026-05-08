package rbac

func (s *Store) AddMutexRoles(roleA, roleB string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[roleA]; !exists {
		return ErrRoleNotFound
	}
	if _, exists := s.roles[roleB]; !exists {
		return ErrRoleNotFound
	}

	if roleA == roleB {
		return ErrSelfInheritance
	}

	if s.mutexRoles[roleA] == nil {
		s.mutexRoles[roleA] = make(map[string]bool)
	}
	if s.mutexRoles[roleB] == nil {
		s.mutexRoles[roleB] = make(map[string]bool)
	}

	s.mutexRoles[roleA][roleB] = true
	s.mutexRoles[roleB][roleA] = true

	if s.checkExistingUserMutexViolationLocked(roleA, roleB) {
		delete(s.mutexRoles[roleA], roleB)
		delete(s.mutexRoles[roleB], roleA)
		return ErrMutexRoleConflict
	}

	return nil
}

func (s *Store) RemoveMutexRoles(roleA, roleB string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[roleA]; !exists {
		return ErrRoleNotFound
	}
	if _, exists := s.roles[roleB]; !exists {
		return ErrRoleNotFound
	}

	if s.mutexRoles[roleA] != nil {
		delete(s.mutexRoles[roleA], roleB)
	}
	if s.mutexRoles[roleB] != nil {
		delete(s.mutexRoles[roleB], roleA)
	}

	return nil
}

func (s *Store) GetMutexRoles(roleID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.roles[roleID]; !exists {
		return nil, ErrRoleNotFound
	}

	mutexRoles := make([]string, 0, len(s.mutexRoles[roleID]))
	for role := range s.mutexRoles[roleID] {
		mutexRoles = append(mutexRoles, role)
	}
	return mutexRoles, nil
}

func (s *Store) checkExistingUserMutexViolationLocked(roleA, roleB string) bool {
	for _, roleSet := range s.userRoles {
		hasA := false
		hasB := false

		for userRole := range roleSet {
			allRoles := s.getAllAncestorsLocked(userRole)
			allRoles[userRole] = true

			if allRoles[roleA] {
				hasA = true
			}
			if allRoles[roleB] {
				hasB = true
			}
		}

		if hasA && hasB {
			return true
		}
	}

	return false
}

func (s *Store) checkUserMutexViolationLocked(userID string, newRoleID string) bool {
	existingRoles := s.userRoles[userID]
	if existingRoles == nil {
		return false
	}

	newAllRoles := s.getAllAncestorsLocked(newRoleID)
	newAllRoles[newRoleID] = true

	for existingRole := range existingRoles {
		existingAllRoles := s.getAllAncestorsLocked(existingRole)
		existingAllRoles[existingRole] = true

		for r1 := range newAllRoles {
			for r2 := range existingAllRoles {
				if s.mutexRoles[r1][r2] {
					return true
				}
			}
		}
	}

	return false
}
