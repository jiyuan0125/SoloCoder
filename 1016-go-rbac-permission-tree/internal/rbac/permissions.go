package rbac

func (s *Store) AddPermissionToRole(roleID string, perm Permission) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[roleID]; !exists {
		return ErrRoleNotFound
	}

	existing, hasPerm := s.rolePermissions[roleID][perm.Key()]
	if hasPerm {
		if existing.Scope.Compare(perm.Scope) >= 0 {
			return nil
		}
	}

	s.rolePermissions[roleID][perm.Key()] = perm

	affectedUsers := s.getUsersWithRoleLocked(roleID)
	for child := range s.getAllChildrenLocked(roleID) {
		affectedUsers = append(affectedUsers, s.getUsersWithRoleLocked(child)...)
	}

	for _, userID := range affectedUsers {
		s.cache.InvalidateUser(userID)
	}

	return nil
}

func (s *Store) RemovePermissionFromRole(roleID, resource, action string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[roleID]; !exists {
		return ErrRoleNotFound
	}

	permKey := resource + ":" + action
	if _, hasPerm := s.rolePermissions[roleID][permKey]; !hasPerm {
		return ErrPermissionNotFound
	}

	delete(s.rolePermissions[roleID], permKey)

	affectedUsers := s.getUsersWithRoleLocked(roleID)
	for child := range s.getAllChildrenLocked(roleID) {
		affectedUsers = append(affectedUsers, s.getUsersWithRoleLocked(child)...)
	}

	for _, userID := range affectedUsers {
		s.cache.InvalidateUser(userID)
	}

	return nil
}

func (s *Store) GetRolePermissions(roleID string) ([]Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.roles[roleID]; !exists {
		return nil, ErrRoleNotFound
	}

	perms := make([]Permission, 0, len(s.rolePermissions[roleID]))
	for _, perm := range s.rolePermissions[roleID] {
		perms = append(perms, perm)
	}
	return perms, nil
}

func (s *Store) getAllChildrenLocked(roleID string) map[string]bool {
	children := make(map[string]bool)
	queue := []string{}

	for child := range s.children[roleID] {
		queue = append(queue, child)
	}

	visited := make(map[string]bool)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true
		children[current] = true

		for child := range s.children[current] {
			if !visited[child] {
				queue = append(queue, child)
			}
		}
	}

	return children
}
