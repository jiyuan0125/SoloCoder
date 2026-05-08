package pkg

import (
	"encoding/base64"
	"encoding/json"
)

type Mode string

const (
	ModeStandard Mode = "standard"
	ModeURLSafe  Mode = "urlsafe"
	ModeMIME     Mode = "mime"
)

type EncodeRequest struct {
	Mode          Mode   `json:"mode"`
	Data          string `json:"data"`
	Padding       bool   `json:"padding"`
	MIMELineWidth int    `json:"mime_line_width,omitempty"`
}

type EncodeResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

type DecodeRequest struct {
	Mode   Mode   `json:"mode"`
	Data   string `json:"data"`
}

type DecodeResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

func BytesToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func Base64ToBytes(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

func MarshalJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func UnmarshalJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}
