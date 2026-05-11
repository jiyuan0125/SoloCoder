package idgen

import "errors"

var (
	ErrClockBackward        = errors.New("clock moved backward, exceeds maximum tolerable delay")
	ErrInvalidNodeID        = errors.New("invalid node ID")
	ErrNodeNotRegistered    = errors.New("node not registered")
	ErrNodeAlreadyExists    = errors.New("node name already exists")
	ErrNoAvailableNodeID    = errors.New("no available node ID")
	ErrInvalidBatchCount    = errors.New("invalid batch count, must be between 1 and 1000")
	ErrBatchGenerationFail  = errors.New("batch generation failed, no IDs returned")
)
