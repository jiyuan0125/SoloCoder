package api

type EncodeRequest struct {
	Mode      string `json:"mode"`
	Input     string `json:"input"`
	Component string `json:"component,omitempty"`
}

type EncodeResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

type DecodeRequest struct {
	Mode  string `json:"mode"`
	Input string `json:"input"`
}

type DecodeResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}
