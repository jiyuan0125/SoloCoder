package core

import "errors"

var (
	ErrInvalidInput   = errors.New("invalid input")
	ErrPointNotFound  = errors.New("point not found")
	ErrPointExists    = errors.New("point already exists")
	ErrNoData         = errors.New("no data available")
)
