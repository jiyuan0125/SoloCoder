package service

import (
	"errors"
	"permission-matrix/models"
	"permission-matrix/storage"
)

var ErrCircularInheritance = errors.New("检测到循环继承")
var ErrRoleNotFound = errors.New("角色不存在")

type PermissionService struct {
	roleRepo        *storage.RoleRepo
	permRepo        *storage.PermissionRepo
	rolePermRepo    *storage.RolePermissionRepo
	inheritanceRepo *storage.RoleInheritanceRepo
	userRoleRepo    *storage.UserRoleRepo
}

func NewPermissionService() *PermissionService {
	return &PermissionService{
		roleRepo:        storage.NewRoleRepo(),
		permRepo:        storage.NewPermissionRepo(),
		rolePermRepo:    storage.NewRolePermissionRepo(),
		inheritanceRepo: storage.NewRoleInheritanceRepo(),
		userRoleRepo:    storage.NewUserRoleRepo(),
	}
}

func (s *PermissionService) CreateRole(name, description string) (*models.Role, error) {
	return s.roleRepo.Create(name, description)
}

func (s *PermissionService) GetRole(id int64) (*models.Role, error) {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		if storage.IsNotFound(err) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return role, nil
}

func (s *PermissionService) ListRoles() ([]models.Role, error) {
	return s.roleRepo.List()
}

func (s *PermissionService) DeleteRole(id int64) error {
	return s.roleRepo.Delete(id)
}

func (s *PermissionService) CreatePermission(resource, action, description string) (*models.Permission, error) {
	perm, err := s.permRepo.GetByResourceAction(resource, action)
	if err == nil {
		return perm, nil
	}
	if !storage.IsNotFound(err) {
		return nil, err
	}
	return s.permRepo.Create(resource, action, description)
}

func (s *PermissionService) ListPermissions(resource string) ([]models.Permission, error) {
	return s.permRepo.ListByResource(resource)
}

func (s *PermissionService) GetPermission(id int64) (*models.Permission, error) {
	return s.permRepo.GetByID(id)
}

func (s *PermissionService) AssignPermissionToRole(roleID, permissionID int64, isAllowed bool) error {
	exists, err := s.roleRepo.Exists(roleID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrRoleNotFound
	}

	exists, err = s.permRepo.Exists(permissionID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("权限不存在")
	}

	return s.rolePermRepo.Assign(roleID, permissionID, isAllowed)
}

func (s *PermissionService) RemovePermissionFromRole(roleID, permissionID int64) error {
	exists, err := s.roleRepo.Exists(roleID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrRoleNotFound
	}

	return s.rolePermRepo.Remove(roleID, permissionID)
}

func (s *PermissionService) AddRoleInheritance(roleID, inheritsFrom int64) error {
	if roleID == inheritsFrom {
		return ErrCircularInheritance
	}

	exists, err := s.roleRepo.Exists(roleID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrRoleNotFound
	}

	exists, err = s.roleRepo.Exists(inheritsFrom)
	if err != nil {
		return err
	}
	if !exists {
		return ErrRoleNotFound
	}

	if err := s.checkCircularInheritance(roleID, inheritsFrom); err != nil {
		return err
	}

	return s.inheritanceRepo.Add(roleID, inheritsFrom)
}

func (s *PermissionService) RemoveRoleInheritance(roleID, inheritsFrom int64) error {
	return s.inheritanceRepo.Remove(roleID, inheritsFrom)
}

func (s *PermissionService) checkCircularInheritance(roleID, newParent int64) error {
	visited := make(map[int64]bool)
	queue := []int64{newParent}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == roleID {
			return ErrCircularInheritance
		}

		if visited[current] {
			continue
		}
		visited[current] = true

		parents, err := s.inheritanceRepo.GetParents(current)
		if err != nil {
			return err
		}

		for _, parent := range parents {
			if parent == roleID {
				return ErrCircularInheritance
			}
			if !visited[parent] {
				queue = append(queue, parent)
			}
		}
	}

	return nil
}

func (s *PermissionService) AssignRoleToUser(userID string, roleID int64) error {
	exists, err := s.roleRepo.Exists(roleID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrRoleNotFound
	}

	return s.userRoleRepo.Assign(userID, roleID)
}

func (s *PermissionService) RemoveRoleFromUser(userID string, roleID int64) error {
	return s.userRoleRepo.Remove(userID, roleID)
}

func (s *PermissionService) GetUserRoles(userID string) ([]models.UserRole, error) {
	return s.userRoleRepo.GetByUser(userID)
}

func (s *PermissionService) GetAllInheritedRoles(roleID int64) ([]int64, error) {
	visited := make(map[int64]bool)
	result := []int64{}
	queue := []int64{roleID}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true
		result = append(result, current)

		parents, err := s.inheritanceRepo.GetParents(current)
		if err != nil {
			return nil, err
		}

		for _, parent := range parents {
			if !visited[parent] {
				queue = append(queue, parent)
			}
		}
	}

	return result, nil
}

type permSource struct {
	roleID    int64
	roleName  string
	isAllowed bool
}

func (s *PermissionService) GetEffectivePermissions(userID string) ([]models.EffectivePermission, error) {
	roleIDs, err := s.userRoleRepo.GetRoleIDsByUser(userID)
	if err != nil {
		return nil, err
	}

	if len(roleIDs) == 0 {
		return []models.EffectivePermission{}, nil
	}

	permMap := make(map[string]permSource)

	for _, directRoleID := range roleIDs {
		allRoles, err := s.GetAllInheritedRoles(directRoleID)
		if err != nil {
			return nil, err
		}

		for _, roleID := range allRoles {
			role, err := s.roleRepo.GetByID(roleID)
			if err != nil {
				if storage.IsNotFound(err) {
					continue
				}
				return nil, err
			}

			perms, err := s.rolePermRepo.GetByRole(roleID)
			if err != nil {
				return nil, err
			}

			for _, rp := range perms {
				key := rp.Resource + ":" + rp.Action
				existing, exists := permMap[key]

				if !exists {
					permMap[key] = permSource{
						roleID:    roleID,
						roleName:  role.Name,
						isAllowed: rp.IsAllowed,
					}
				} else {
					if existing.isAllowed && !rp.IsAllowed {
						permMap[key] = permSource{
							roleID:    roleID,
							roleName:  role.Name,
							isAllowed: rp.IsAllowed,
						}
					}
				}
			}
		}
	}

	result := make([]models.EffectivePermission, 0, len(permMap))
	for key, src := range permMap {
		perm, err := s.permRepo.GetByResourceAction(
			key[:len(key)-len(key[stringsLastIndex(key, ":")+1:])-1],
			key[stringsLastIndex(key, ":")+1:],
		)
		if err != nil {
			continue
		}

		result = append(result, models.EffectivePermission{
			PermissionID: perm.ID,
			Resource:     perm.Resource,
			Action:       perm.Action,
			IsAllowed:    src.isAllowed,
			SourceRoleID: src.roleID,
			SourceRole:   src.roleName,
		})
	}

	return result, nil
}

func (s *PermissionService) CheckUserPermission(userID, resource, action string) (*models.PermissionCheckResult, error) {
	roleIDs, err := s.userRoleRepo.GetRoleIDsByUser(userID)
	if err != nil {
		return nil, err
	}

	if len(roleIDs) == 0 {
		return &models.PermissionCheckResult{
			HasPermission: false,
			IsAllowed:     false,
		}, nil
	}

	perm, err := s.permRepo.GetByResourceAction(resource, action)
	if err != nil {
		if storage.IsNotFound(err) {
			return &models.PermissionCheckResult{
				HasPermission: false,
				IsAllowed:     false,
			}, nil
		}
		return nil, err
	}

	result := &models.PermissionCheckResult{
		HasPermission: false,
		IsAllowed:     false,
	}

	for _, directRoleID := range roleIDs {
		allRoles, err := s.GetAllInheritedRoles(directRoleID)
		if err != nil {
			return nil, err
		}

		for _, roleID := range allRoles {
			role, err := s.roleRepo.GetByID(roleID)
			if err != nil {
				if storage.IsNotFound(err) {
					continue
				}
				return nil, err
			}

			perms, err := s.rolePermRepo.GetByRole(roleID)
			if err != nil {
				return nil, err
			}

			for _, rp := range perms {
				if rp.PermissionID == perm.ID {
					if !result.HasPermission {
						result.HasPermission = true
						result.IsAllowed = rp.IsAllowed
						result.SourceRoleID = roleID
						result.SourceRole = role.Name
					} else if result.IsAllowed && !rp.IsAllowed {
						result.IsAllowed = false
						result.SourceRoleID = roleID
						result.SourceRole = role.Name
					}
				}
			}
		}
	}

	return result, nil
}

func stringsLastIndex(s, substr string) int {
	for i := len(s) - len(substr); i >= 0; i-- {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func (s *PermissionService) GetRolePermissions(roleID int64) ([]models.RolePermission, error) {
	exists, err := s.roleRepo.Exists(roleID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrRoleNotFound
	}
	return s.rolePermRepo.GetByRole(roleID)
}

func (s *PermissionService) GetRoleParents(roleID int64) ([]int64, error) {
	exists, err := s.roleRepo.Exists(roleID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrRoleNotFound
	}
	return s.inheritanceRepo.GetParents(roleID)
}
