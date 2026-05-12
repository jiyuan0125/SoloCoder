package registry

import "errors"

var (
	ErrServiceNotFound  = errors.New("service not found")
	ErrInstanceNotFound = errors.New("instance not found")
)
