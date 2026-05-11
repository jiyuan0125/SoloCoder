package common

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type IDResponse struct {
	ID string `json:"id"`
}
