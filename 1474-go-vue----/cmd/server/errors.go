package main

import "errors"

var (
	errInvalidFenceType = errors.New("invalid fence type, must be recommended, forbidden, or dispatch")
	errInvalidFenceID   = errors.New("invalid fence id")
	errInvalidBikeID    = errors.New("invalid bike id")
	errInvalidTaskID    = errors.New("invalid task id")
	errInvalidAction    = errors.New("invalid action")
)
