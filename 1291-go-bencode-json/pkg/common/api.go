package common

type DecodeRequest struct {
	Filename    string `json:"filename,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

type DecodeResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type EncodeRequest struct {
	Data     interface{} `json:"data"`
	Filename string      `json:"filename,omitempty"`
}

type EncodeResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Size    int64  `json:"size,omitempty"`
}

const (
	DefaultPort = "8100"
	EnvPort     = "BENCODE_PORT"
	FlagPort    = "port"
)
