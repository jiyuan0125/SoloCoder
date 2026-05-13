package models

import "time"

type Role struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Permission struct {
	ID          int64     `json:"id"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
}

type RolePermission struct {
	RoleID       int64  `json:"role_id"`
	PermissionID int64  `json:"permission_id"`
	IsAllowed    bool   `json:"is_allowed"`
	Resource     string `json:"resource,omitempty"`
	Action       string `json:"action,omitempty"`
}

type UserRole struct {
	UserID    string    `json:"user_id"`
	RoleID    int64     `json:"role_id"`
	RoleName  string    `json:"role_name,omitempty"`
	AssignedAt time.Time `json:"assigned_at"`
}

type RoleInheritance struct {
	RoleID       int64 `json:"role_id"`
	InheritsFrom int64 `json:"inherits_from"`
}

type EffectivePermission struct {
	PermissionID int64  `json:"permission_id"`
	Resource     string `json:"resource"`
	Action       string `json:"action"`
	IsAllowed    bool   `json:"is_allowed"`
	SourceRoleID int64  `json:"source_role_id"`
	SourceRole   string `json:"source_role"`
}

type PermissionCheckResult struct {
	HasPermission bool   `json:"has_permission"`
	IsAllowed     bool   `json:"is_allowed"`
	SourceRoleID  int64  `json:"source_role_id,omitempty"`
	SourceRole    string `json:"source_role,omitempty"`
}
