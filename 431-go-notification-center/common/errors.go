package common

import "fmt"

type ErrorCode int

const (
	ErrCodeSuccess           ErrorCode = 0
	ErrCodeInvalidRequest    ErrorCode = 1001
	ErrCodeNotFound          ErrorCode = 1002
	ErrCodeDuplicate         ErrorCode = 1003
	ErrCodeInternal          ErrorCode = 1004
	ErrCodeExceedLimit       ErrorCode = 1005
	ErrCodeInvalidTemplate   ErrorCode = 1006
	ErrCodeSendFailed        ErrorCode = 1007
)

type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

func NewAPIError(code ErrorCode, message string) APIError {
	return APIError{
		Code:    code,
		Message: message,
	}
}

func ErrInvalidRequest(message string) APIError {
	return NewAPIError(ErrCodeInvalidRequest, message)
}

func ErrNotFound(message string) APIError {
	return NewAPIError(ErrCodeNotFound, message)
}

func ErrDuplicate(message string) APIError {
	return NewAPIError(ErrCodeDuplicate, message)
}

func ErrInternal(message string) APIError {
	return NewAPIError(ErrCodeInternal, message)
}

func ErrExceedLimit(message string) APIError {
	return NewAPIError(ErrCodeExceedLimit, message)
}

func ErrInvalidTemplate(message string) APIError {
	return NewAPIError(ErrCodeInvalidTemplate, message)
}

func ErrSendFailed(message string) APIError {
	return NewAPIError(ErrCodeSendFailed, message)
}
