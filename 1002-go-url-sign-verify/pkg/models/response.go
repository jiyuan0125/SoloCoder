package models

type UserInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
	Email string `json:"email"`
}

type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Debug   *DebugInfo  `json:"debug,omitempty"`
}

type DebugInfo struct {
	SignString string `json:"sign_string"`
	Signature  string `json:"signature"`
	ClientSign string `json:"client_sign,omitempty"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func NewSuccessResponse(data interface{}) *ApiResponse {
	return &ApiResponse{
		Success: true,
		Data:    data,
	}
}

func NewErrorResponse(err string) *ErrorResponse {
	return &ErrorResponse{
		Success: false,
		Error:   err,
	}
}

func NewDebugResponse(data interface{}, debugInfo *DebugInfo) *ApiResponse {
	return &ApiResponse{
		Success: true,
		Data:    data,
		Debug:   debugInfo,
	}
}
