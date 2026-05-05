package common

import "fmt"

const (
	ErrCodeSuccess            = 0
	ErrCodeInvalidRequest     = 400
	ErrCodeUnauthorized       = 401
	ErrCodeNotFound           = 404
	ErrCodeConflict           = 409
	ErrCodeInternalError      = 500
	ErrCodeTemplateNotFound   = 1001
	ErrCodeTemplateNameExists = 1002
	ErrCodeTemplateInvalid    = 1003
	ErrCodeTaskNotFound       = 2001
	ErrCodeTaskNotCompleted   = 2002
	ErrCodeTaskFailed         = 2003
	ErrCodeFileExpired        = 3001
	ErrCodeFileNotFound       = 3002
	ErrCodeUserQueueFull      = 4001
)

var errMessages = map[int]string{
	ErrCodeSuccess:            "success",
	ErrCodeInvalidRequest:     "invalid request",
	ErrCodeUnauthorized:       "unauthorized",
	ErrCodeNotFound:           "resource not found",
	ErrCodeConflict:           "resource conflict",
	ErrCodeInternalError:      "internal server error",
	ErrCodeTemplateNotFound:   "template not found",
	ErrCodeTemplateNameExists: "template name already exists",
	ErrCodeTemplateInvalid:    "invalid template configuration",
	ErrCodeTaskNotFound:       "export task not found",
	ErrCodeTaskNotCompleted:   "export task not completed yet",
	ErrCodeTaskFailed:         "export task failed",
	ErrCodeFileExpired:        "export file has expired",
	ErrCodeFileNotFound:       "export file not found",
	ErrCodeUserQueueFull:      "user has too many pending tasks",
}

func ErrorMessage(code int) string {
	if msg, ok := errMessages[code]; ok {
		return msg
	}
	return "unknown error"
}

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func NewAppError(code int, detail ...string) *AppError {
	msg := ErrorMessage(code)
	appErr := &AppError{
		Code:    code,
		Message: msg,
	}
	if len(detail) > 0 && detail[0] != "" {
		appErr.Detail = detail[0]
	}
	return appErr
}

func NewInvalidRequestError(detail string) *AppError {
	return NewAppError(ErrCodeInvalidRequest, detail)
}

func NewNotFoundError(detail string) *AppError {
	return NewAppError(ErrCodeNotFound, detail)
}

func NewInternalError(detail string) *AppError {
	return NewAppError(ErrCodeInternalError, detail)
}
