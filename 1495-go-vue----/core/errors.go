package core

import "errors"

var (
	ErrNotFound             = errors.New("resource not found")
	ErrInsufficientStock    = errors.New("insufficient chemical stock")
	ErrStaffOverloaded      = errors.New("staff has exceeded daily operation limit (max 4)")
	ErrInvalidData          = errors.New("invalid data")
	ErrContractExpired      = errors.New("contract has expired")
	ErrDuplicateResource    = errors.New("resource already exists")
)
