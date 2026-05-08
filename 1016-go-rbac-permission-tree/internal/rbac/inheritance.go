package rbac

func (s *Store) AddParent(childID, parentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[childID]; !exists {
		return ErrRoleNotFound
	}
	if _, exists := s.roles[parentID]; !exists {
		return ErrRoleNotFound
	}

	if childID == parentID {
		return ErrSelfInheritance
	}

	if s.parents[childID][parentID] {
		return ErrInheritanceExists
	}

	if s.wouldCreateCycleLocked(childID, parentID) {
		return ErrInheritanceCycle
	}

	if s.wouldViolateMutexLockInheritance(childID, parentID) {
		return ErrInheritanceConflict
	}

	if s.parents[parentID] == nil {
		s.parents[parentID] = make(map[string]bool)
	}
	if s.children[childID] == nil {
		s.children[childID] = make(map[string]bool)
	}

	s.parents[childID][parentID] = true
	s.children[parentID][childID] = true

	affectedUsers := s.getUsersWithRoleLocked(childID)
	for child := range s.getAllChildrenLocked(childID) {
		affectedUsers = append(affectedUsers, s.getUsersWithRoleLocked(child)...)
	}

	for _, userID := range affectedUsers {
		s.cache.InvalidateUser(userID)
	}

	return nil
}

func (s *Store) RemoveParent(childID, parentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[childID]; !exists {
		return ErrRoleNotFound
	}
	if _, exists := s.roles[parentID]; !exists {
		return ErrRoleNotFound
	}

	if !s.parents[childID][parentID] {
		return ErrRoleNotFound
	}

	delete(s.parents[childID], parentID)
	delete(s.children[parentID], childID)

	affectedUsers := s.getUsersWithRoleLocked(childID)
	for child := range s.getAllChildrenLocked(childID) {
		affectedUsers = append(affectedUsers, s.getUsersWithRoleLocked(child)...)
	}

	for _, userID := range affectedUsers {
		s.cache.InvalidateUser(userID)
	}

	return nil
}

func (s *Store) GetRoleParents(roleID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.roles[roleID]; !exists {
		return nil, ErrRoleNotFound
	}

	parents := make([]string, 0, len(s.parents[roleID]))
	for parent := range s.parents[roleID] {
		parents = append(parents, parent)
	}
	return parents, nil
}

func (s *Store) GetRoleChildren(roleID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.roles[roleID]; !exists {
		return nil, ErrRoleNotFound
	}

	children := make([]string, 0, len(s.children[roleID]))
	for child := range s.children[roleID] {
		children = append(children, child)
	}
	return children, nil
}

func (s *Store) wouldCreateCycleLocked(startID, parentID string) bool {
	visited := make(map[string]bool)
	stack := []string{parentID}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[current] {
			continue
		}
		visited[current] = true

		if current == startID {
			return true
		}

		for ancestor := range s.parents[current] {
			if !visited[ancestor] {
				stack = append(stack, ancestor)
			}
		}
	}

	return false
}

func (s *Store) wouldViolateMutexLockInheritance(childID, parentID string) bool {
	childAllRoles := s.getAllAncestorsLocked(childID)
	childAllRoles[childID] = true

	parentAllRoles := s.getAllAncestorsLocked(parentID)
	parentAllRoles[parentID] = true

	for roleA := range childAllRoles {
		for roleB := range parentAllRoles {
			if s.mutexRoles[roleA][roleB] {
				return true
			}
		}
	}

	for _, userRoles := range s.userRoles {
		userHasChild := false
		for r := range userRoles {
			if r == childID || s.getAllAncestorsLocked(r)[childID] {
				userHasChild = true
				break
			}
		}

		if !userHasChild {
			continue
		}

		userRoleSet := make(map[string]bool)
		for r := range userRoles {
			userRoleSet[r] = true
			for a := range s.getAllAncestorsLocked(r) {
				userRoleSet[a] = true
			}
		}

		for parentRole := range parentAllRoles {
			for userRole := range userRoleSet {
				if s.mutexRoles[parentRole][userRole] {
					return true
				}
			}
		}
	}

	return false
}

func (s *Store) getAllAncestorsLocked(roleID string) map[string]bool {
	ancestors := make(map[string]bool)
	stack := []string{}

	for parent := range s.parents[roleID] {
		stack = append(stack, parent)
	}

	visited := make(map[string]bool)
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[current] {
			continue
		}
		visited[current] = true
		ancestors[current] = true

		for parent := range s.parents[current] {
			if !visited[parent] {
				stack = append(stack, parent)
			}
		}
	}

	return ancestors
}

func (s *Store) getAllAncestorsIncludingSelf(roleID string) map[string]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := s.getAllAncestorsLocked(roleID)
	result[roleID] = true
	return result
}
