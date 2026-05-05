package common

type ErrorCode int

const (
	ErrCodeSuccess ErrorCode = 0

	ErrCodeInvalidRequest     ErrorCode = 1001
	ErrCodeInvalidEventName   ErrorCode = 1002
	ErrCodeInvalidTimestamp   ErrorCode = 1003
	ErrCodePropertiesTooLarge ErrorCode = 1004
	ErrCodeMissingUserID      ErrorCode = 1005
	ErrCodeMissingDeviceID    ErrorCode = 1006

	ErrCodeInternalError ErrorCode = 2001
	ErrCodeStorageError  ErrorCode = 2002

	ErrCodeDuplicateEvent ErrorCode = 3001
)

func (e ErrorCode) String() string {
	switch e {
	case ErrCodeSuccess:
		return "success"
	case ErrCodeInvalidRequest:
		return "invalid_request"
	case ErrCodeInvalidEventName:
		return "invalid_event_name"
	case ErrCodeInvalidTimestamp:
		return "invalid_timestamp"
	case ErrCodePropertiesTooLarge:
		return "properties_too_large"
	case ErrCodeMissingUserID:
		return "missing_user_id"
	case ErrCodeMissingDeviceID:
		return "missing_device_id"
	case ErrCodeInternalError:
		return "internal_error"
	case ErrCodeStorageError:
		return "storage_error"
	case ErrCodeDuplicateEvent:
		return "duplicate_event"
	default:
		return "unknown_error"
	}
}

type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func NewAPIError(code ErrorCode, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

func (e *APIError) Error() string {
	return e.Message
}
