package common

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewResponse(code int, data interface{}) Response {
	return Response{
		Code:    code,
		Message: CodeMessage(code),
		Data:    data,
	}
}

func NewSuccessResponse(data interface{}) Response {
	return NewResponse(CodeSuccess, data)
}

func NewErrorResponse(code int, msg string) Response {
	if msg == "" {
		msg = CodeMessage(code)
	}
	return Response{
		Code:    code,
		Message: msg,
		Data:    nil,
	}
}
