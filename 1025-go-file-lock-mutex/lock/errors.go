package lock

import "errors"

var (
	ErrLockTimeout      = errors.New("lock acquisition timed out")
	ErrDeadlockDetected = errors.New("deadlock detected: max wait duration exceeded")
	ErrLockNotHeld      = errors.New("lock not held")
	ErrInvalidMode      = errors.New("invalid lock mode")
	ErrLockAlreadyHeld  = errors.New("lock already held")
	ErrProcessDead      = errors.New("previous lock holder process is dead")
)
