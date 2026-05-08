package snowflake

import "errors"

var (
	ErrInvalidEpoch        = errors.New("invalid epoch: epoch must be non-negative")
	ErrInvalidMachineID    = errors.New("invalid machine ID: must be between 0 and 1023")
	ErrInvalidMaxWait      = errors.New("invalid max wait: must be non-negative")
	ErrClockTooFarBack     = errors.New("clock moved backwards too far, cannot wait")
	ErrClockBackwards      = errors.New("clock moved backwards")
	ErrMachineIDGenFailed  = errors.New("failed to generate machine ID")
)
