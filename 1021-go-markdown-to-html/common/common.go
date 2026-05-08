package common

type ConvertRequest struct {
	Markdown string `json:"markdown"`
}

type ConvertResponse struct {
	HTML    string `json:"html"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
