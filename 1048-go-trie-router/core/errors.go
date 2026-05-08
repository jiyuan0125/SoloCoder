package core

import "errors"

var (
	ErrInvalidPath        = errors.New("invalid path")
	ErrInvalidMethod      = errors.New("invalid method")
	ErrWildcardNotAtEnd   = errors.New("wildcard must be at the end of path")
	ErrParamNameConflict  = errors.New("parameter name conflict")
	ErrParamTypeConflict  = errors.New("parameter type conflict")
	ErrRouteExists        = errors.New("route already exists")
)
