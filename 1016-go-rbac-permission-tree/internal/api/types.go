package api

import "time"

type Scope string

const (
	ScopeSelf       Scope = "self"
	ScopeDepartment Scope = "department"
	ScopeAll        Scope = "all"
)

type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Scope    Scope  `json:"scope"`
}

type CreateRoleRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AddPermissionRequest struct {
	Permissions []Permission `json:"permissions"`
}

type RemovePermissionRequest struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type AddParentRequest struct {
	ParentID string `json:"parent_id"`
}

type RemoveParentRequest struct {
	ParentID string `json:"parent_id"`
}

type AddRoleToUserRequest struct {
	RoleID string `json:"role_id"`
}

type RemoveRoleFromUserRequest struct {
	RoleID string `json:"role_id"`
}

type CheckPermissionRequest struct {
	UserID   string `json:"user_id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type CheckPermissionResponse struct {
	Allowed bool  `json:"allowed"`
	Scope   Scope `json:"scope,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ListRolesResponse struct {
	Roles []Role `json:"roles"`
}

type GetRolePermissionsResponse struct {
	Permissions []Permission `json:"permissions"`
}

type GetRoleParentsResponse struct {
	Parents []string `json:"parents"`
}

type GetUserRolesResponse struct {
	Roles []string `json:"roles"`
}
