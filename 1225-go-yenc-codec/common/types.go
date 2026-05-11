package common

type EncodeRequest struct {
	Data string `json:"data"`
	LineSize *int `json:"line_size,omitempty"`
}

type EncodeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Encoded string `json:"encoded,omitempty"`
}

type DecodeRequest struct {
	Encoded string `json:"encoded"`
}

type DecodeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    string `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
