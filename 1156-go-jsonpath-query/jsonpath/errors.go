package jsonpath

import "errors"

var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrUnterminatedString = errors.New("unterminated string")
	ErrUnexpectedToken    = errors.New("unexpected token")
	ErrInvalidPath        = errors.New("invalid jsonpath")
)
