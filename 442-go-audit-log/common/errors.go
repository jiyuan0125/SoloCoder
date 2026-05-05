package common

import "errors"

var (
	ErrLogNotFound          = errors.New("audit log not found")
	ErrUserLocked           = errors.New("user is locked")
	ErrInvalidOperationType = errors.New("invalid operation type")
	ErrMissingUserID        = errors.New("user_id is required")
	ErrMissingOperationType = errors.New("operation_type is required")
	ErrNeedApproval         = errors.New("this operation requires approval")
	ErrStorageCapacity      = errors.New("storage capacity exceeded")
	ErrExportNotApproved    = errors.New("export operation not approved")
)
