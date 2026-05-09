package pool

import "errors"

var (
	ErrPoolClosed = errors.New("pool is closed")
	ErrTimeout    = errors.New("task timed out")
)
