package common

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorCode int

const (
	CodeSuccess           ErrorCode = 0
	CodeFormatNotSupport  ErrorCode = 2001
	CodeDecodeFailed      ErrorCode = 2002
	CodeEncodeFailed      ErrorCode = 2003
	CodeAdapterExists     ErrorCode = 2004
	CodeAdapterNotFound   ErrorCode = 2005
	CodeInvalidRequest    ErrorCode = 2006
	CodeBusinessError     ErrorCode = 2010
)

func (e ErrorCode) String() string {
	switch e {
	case CodeSuccess:
		return "success"
	case CodeFormatNotSupport:
		return "format not support"
	case CodeDecodeFailed:
		return "decode failed"
	case CodeEncodeFailed:
		return "encode failed"
	case CodeAdapterExists:
		return "adapter already exists"
	case CodeAdapterNotFound:
		return "adapter not found"
	case CodeInvalidRequest:
		return "invalid request"
	case CodeBusinessError:
		return "business error"
	default:
		return "unknown error"
	}
}

func Success(data interface{}) Response {
	return Response{
		Code:    int(CodeSuccess),
		Message: CodeSuccess.String(),
		Data:    data,
	}
}

func Error(code ErrorCode, message string) Response {
	if message == "" {
		message = code.String()
	}
	return Response{
		Code:    int(code),
		Message: message,
	}
}
