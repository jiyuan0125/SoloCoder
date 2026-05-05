package common

type ErrorCode int

const (
	ErrCodeSuccess            ErrorCode = 0
	ErrCodeInvalidRequest     ErrorCode = 1001
	ErrCodeTicketNotFound     ErrorCode = 1002
	ErrCodeHandlerNotFound    ErrorCode = 1003
	ErrCodeInvalidCategory    ErrorCode = 1004
	ErrCodeInvalidPriority    ErrorCode = 1005
	ErrCodeInvalidStatus      ErrorCode = 1006
	ErrCodeReassignReasonRequired ErrorCode = 1007
	ErrCodeTicketNotClosed    ErrorCode = 1008
	ErrCodeInvalidRating      ErrorCode = 1009
	ErrCodeParentTicketExists ErrorCode = 1010
	ErrCodeCircularDependency ErrorCode = 1011
	ErrCodePermissionDenied   ErrorCode = 1012
	ErrCodeInternalError      ErrorCode = 2001
)

var ErrorMessages = map[ErrorCode]string{
	ErrCodeSuccess:            "success",
	ErrCodeInvalidRequest:     "invalid request",
	ErrCodeTicketNotFound:     "ticket not found",
	ErrCodeHandlerNotFound:    "handler not found",
	ErrCodeInvalidCategory:    "invalid category",
	ErrCodeInvalidPriority:    "invalid priority",
	ErrCodeInvalidStatus:      "invalid status",
	ErrCodeReassignReasonRequired: "reassign reason is required",
	ErrCodeTicketNotClosed:    "ticket is not closed",
	ErrCodeInvalidRating:      "rating must be between 1 and 5",
	ErrCodeParentTicketExists: "ticket already has a parent",
	ErrCodeCircularDependency: "circular dependency detected",
	ErrCodePermissionDenied:   "permission denied",
	ErrCodeInternalError:      "internal error",
}

func GetErrorMessage(code ErrorCode) string {
	if msg, ok := ErrorMessages[code]; ok {
		return msg
	}
	return "unknown error"
}

func (c ErrorCode) String() string {
	return GetErrorMessage(c)
}
