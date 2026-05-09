package rangehandler

import (
	"errors"
	"net/http"
)

var (
	ErrInvalidRangeFormat = errors.New("invalid range header format")
	ErrInvalidRangeUnit   = errors.New("invalid range unit, only bytes is supported")
	ErrRangeUnsatisfiable = errors.New("requested range is not satisfiable")
)

type Range struct {
	Start int64
	End   int64
}

type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}
