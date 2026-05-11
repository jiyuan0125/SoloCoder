package core

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrDuplicateDeviceCode = errors.New("duplicate device code")
	ErrDuplicateActiveDevice = errors.New("duplicate active device at same location and type")
	ErrInvalidDeviceType  = errors.New("invalid device type")
	ErrInvalidDeviceStatus = errors.New("invalid device status")
	ErrInvalidFrequency   = errors.New("invalid inspection frequency")
	ErrInvalidDrillType   = errors.New("invalid drill type")
	ErrTaskNotPending     = errors.New("task is not pending")
	ErrInvalidPointOrder  = errors.New("invalid point check order, must follow route order")
	ErrPlanAlreadyCompleted = errors.New("drill plan already completed")
	ErrPointAlreadyChecked = errors.New("point already checked")
	ErrTaskAlreadyCompleted = errors.New("task already completed")
)
