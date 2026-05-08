package rbac

import "errors"

var (
	ErrRoleNotFound        = errors.New("role not found")
	ErrRoleAlreadyExists   = errors.New("role already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrPermissionNotFound  = errors.New("permission not found")
	ErrInheritanceCycle    = errors.New("role inheritance would create a cycle")
	ErrInheritanceConflict = errors.New("role inheritance would violate constraints")
	ErrMutexRoleConflict   = errors.New("user cannot have mutually exclusive roles")
	ErrSelfInheritance     = errors.New("role cannot inherit from itself")
	ErrInheritanceExists   = errors.New("inheritance relationship already exists")
)
