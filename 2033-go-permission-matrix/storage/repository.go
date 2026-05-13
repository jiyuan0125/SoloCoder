package storage

import (
	"database/sql"
	"permission-matrix/models"
	"time"
)

type RoleRepo struct{}

func NewRoleRepo() *RoleRepo {
	return &RoleRepo{}
}

func (r *RoleRepo) Create(name, description string) (*models.Role, error) {
	result, err := GetDB().Exec(
		"INSERT INTO roles (name, description) VALUES (?, ?)",
		name, description,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *RoleRepo) GetByID(id int64) (*models.Role, error) {
	role := &models.Role{}
	err := GetDB().QueryRow(
		"SELECT id, name, description, created_at FROM roles WHERE id = ?",
		id,
	).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleRepo) GetByName(name string) (*models.Role, error) {
	role := &models.Role{}
	err := GetDB().QueryRow(
		"SELECT id, name, description, created_at FROM roles WHERE name = ?",
		name,
	).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleRepo) List() ([]models.Role, error) {
	rows, err := GetDB().Query(
		"SELECT id, name, description, created_at FROM roles ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *RoleRepo) Update(id int64, description string) (*models.Role, error) {
	_, err := GetDB().Exec(
		"UPDATE roles SET description = ? WHERE id = ?",
		description, id,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(id)
}

func (r *RoleRepo) Delete(id int64) error {
	_, err := GetDB().Exec("DELETE FROM roles WHERE id = ?", id)
	return err
}

func (r *RoleRepo) Exists(id int64) (bool, error) {
	var count int
	err := GetDB().QueryRow("SELECT COUNT(*) FROM roles WHERE id = ?", id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

type PermissionRepo struct{}

func NewPermissionRepo() *PermissionRepo {
	return &PermissionRepo{}
}

func (p *PermissionRepo) Create(resource, action, description string) (*models.Permission, error) {
	result, err := GetDB().Exec(
		"INSERT INTO permissions (resource, action, description) VALUES (?, ?, ?)",
		resource, action, description,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return p.GetByID(id)
}

func (p *PermissionRepo) GetByID(id int64) (*models.Permission, error) {
	perm := &models.Permission{}
	err := GetDB().QueryRow(
		"SELECT id, resource, action, description FROM permissions WHERE id = ?",
		id,
	).Scan(&perm.ID, &perm.Resource, &perm.Action, &perm.Description)
	if err != nil {
		return nil, err
	}
	return perm, nil
}

func (p *PermissionRepo) GetByResourceAction(resource, action string) (*models.Permission, error) {
	perm := &models.Permission{}
	err := GetDB().QueryRow(
		"SELECT id, resource, action, description FROM permissions WHERE resource = ? AND action = ?",
		resource, action,
	).Scan(&perm.ID, &perm.Resource, &perm.Action, &perm.Description)
	if err != nil {
		return nil, err
	}
	return perm, nil
}

func (p *PermissionRepo) ListByResource(resource string) ([]models.Permission, error) {
	query := "SELECT id, resource, action, description FROM permissions"
	var args []interface{}
	if resource != "" {
		query += " WHERE resource = ?"
		args = append(args, resource)
	}
	query += " ORDER BY resource, action"

	rows, err := GetDB().Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.Permission
	for rows.Next() {
		var perm models.Permission
		if err := rows.Scan(&perm.ID, &perm.Resource, &perm.Action, &perm.Description); err != nil {
			return nil, err
		}
		perms = append(perms, perm)
	}
	return perms, nil
}

func (p *PermissionRepo) Delete(id int64) error {
	_, err := GetDB().Exec("DELETE FROM permissions WHERE id = ?", id)
	return err
}

func (p *PermissionRepo) Exists(id int64) (bool, error) {
	var count int
	err := GetDB().QueryRow("SELECT COUNT(*) FROM permissions WHERE id = ?", id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

type RolePermissionRepo struct{}

func NewRolePermissionRepo() *RolePermissionRepo {
	return &RolePermissionRepo{}
}

func (rp *RolePermissionRepo) Assign(roleID, permissionID int64, isAllowed bool) error {
	_, err := GetDB().Exec(
		"INSERT OR REPLACE INTO role_permissions (role_id, permission_id, is_allowed) VALUES (?, ?, ?)",
		roleID, permissionID, isAllowed,
	)
	return err
}

func (rp *RolePermissionRepo) Remove(roleID, permissionID int64) error {
	_, err := GetDB().Exec(
		"DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?",
		roleID, permissionID,
	)
	return err
}

func (rp *RolePermissionRepo) GetByRole(roleID int64) ([]models.RolePermission, error) {
	rows, err := GetDB().Query(`
		SELECT rp.role_id, rp.permission_id, rp.is_allowed, p.resource, p.action
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		WHERE rp.role_id = ?
	`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.RolePermission
	for rows.Next() {
		var p models.RolePermission
		if err := rows.Scan(&p.RoleID, &p.PermissionID, &p.IsAllowed, &p.Resource, &p.Action); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

type RoleInheritanceRepo struct{}

func NewRoleInheritanceRepo() *RoleInheritanceRepo {
	return &RoleInheritanceRepo{}
}

func (ri *RoleInheritanceRepo) Add(roleID, inheritsFrom int64) error {
	_, err := GetDB().Exec(
		"INSERT OR IGNORE INTO role_inheritances (role_id, inherits_from) VALUES (?, ?)",
		roleID, inheritsFrom,
	)
	return err
}

func (ri *RoleInheritanceRepo) Remove(roleID, inheritsFrom int64) error {
	_, err := GetDB().Exec(
		"DELETE FROM role_inheritances WHERE role_id = ? AND inherits_from = ?",
		roleID, inheritsFrom,
	)
	return err
}

func (ri *RoleInheritanceRepo) GetParents(roleID int64) ([]int64, error) {
	rows, err := GetDB().Query(
		"SELECT inherits_from FROM role_inheritances WHERE role_id = ?",
		roleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parents []int64
	for rows.Next() {
		var parentID int64
		if err := rows.Scan(&parentID); err != nil {
			return nil, err
		}
		parents = append(parents, parentID)
	}
	return parents, nil
}

func (ri *RoleInheritanceRepo) GetAll() ([]models.RoleInheritance, error) {
	rows, err := GetDB().Query("SELECT role_id, inherits_from FROM role_inheritances")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inheritances []models.RoleInheritance
	for rows.Next() {
		var inh models.RoleInheritance
		if err := rows.Scan(&inh.RoleID, &inh.InheritsFrom); err != nil {
			return nil, err
		}
		inheritances = append(inheritances, inh)
	}
	return inheritances, nil
}

type UserRoleRepo struct{}

func NewUserRoleRepo() *UserRoleRepo {
	return &UserRoleRepo{}
}

func (ur *UserRoleRepo) Assign(userID string, roleID int64) error {
	_, err := GetDB().Exec(
		"INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)",
		userID, roleID,
	)
	return err
}

func (ur *UserRoleRepo) Remove(userID string, roleID int64) error {
	_, err := GetDB().Exec(
		"DELETE FROM user_roles WHERE user_id = ? AND role_id = ?",
		userID, roleID,
	)
	return err
}

func (ur *UserRoleRepo) GetByUser(userID string) ([]models.UserRole, error) {
	rows, err := GetDB().Query(`
		SELECT ur.user_id, ur.role_id, ur.assigned_at, r.name
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userRoles []models.UserRole
	for rows.Next() {
		var ur models.UserRole
		if err := rows.Scan(&ur.UserID, &ur.RoleID, &ur.AssignedAt, &ur.RoleName); err != nil {
			return nil, err
		}
		userRoles = append(userRoles, ur)
	}
	return userRoles, nil
}

func (ur *UserRoleRepo) GetRoleIDsByUser(userID string) ([]int64, error) {
	rows, err := GetDB().Query(
		"SELECT role_id FROM user_roles WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roleIDs []int64
	for rows.Next() {
		var roleID int64
		if err := rows.Scan(&roleID); err != nil {
			return nil, err
		}
		roleIDs = append(roleIDs, roleID)
	}
	return roleIDs, nil
}

func IsNotFound(err error) bool {
	return err == sql.ErrNoRows
}

func NullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func NullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}
