package plugin

import "errors"

var (
	ErrPluginNotFound   = errors.New("plugin not found")
	ErrVersionMismatch  = errors.New("plugin API version mismatch")
	ErrTimeout          = errors.New("plugin execution timed out")
	ErrPanic            = errors.New("plugin panicked")
	ErrPluginNotLoaded  = errors.New("plugin not loaded")
)
