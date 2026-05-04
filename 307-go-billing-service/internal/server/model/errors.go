package model

import "errors"

var (
	ErrCustomerNotFound = errors.New("customer not found")
	ErrPlanNotFound     = errors.New("plan not found")
	ErrBillNotFound     = errors.New("bill not found")
	ErrUsageNotFound    = errors.New("usage record not found")
	ErrBillAlreadyExists = errors.New("bill already exists for this customer and month")
	ErrInvalidDate      = errors.New("invalid date")
)
