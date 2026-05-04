package common

type DetectRequest struct {
	Data []byte `json:"data,omitempty"`
	Path string `json:"path,omitempty"`
}

type DetectResponse struct {
	MimeType string `json:"mime_type"`
	Error    string `json:"error,omitempty"`
}
