package common

import "errors"

var (
	ErrInvalidLuggageTag   = errors.New("invalid luggage tag format, must be 10 alphanumeric characters")
	ErrLuggageAlreadyExists = errors.New("luggage tag already has an active record")
	ErrLuggageNotFound     = errors.New("luggage tag not found")
	ErrInvalidStage        = errors.New("invalid stage")
	ErrFlightNotFound      = errors.New("flight not found")
	ErrInvalidFlightNumber = errors.New("invalid flight number")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrInvalidRole         = errors.New("invalid role")
	ErrOperationFailed     = errors.New("operation failed")
	ErrStageAlreadyScanned = errors.New("stage already scanned")
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewErrorResponse(code int, message string) ErrorResponse {
	return ErrorResponse{
		Code:    code,
		Message: message,
	}
}
