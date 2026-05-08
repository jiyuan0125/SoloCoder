package main

import "errors"

var (
	errInvalidIntervalType = errors.New("invalid interval type")
	errInvalidOperation    = errors.New("invalid operation")
	errInvalidTimezone     = errors.New("invalid timezone")
	errMissingPoint        = errors.New("missing point for query operation")
)
