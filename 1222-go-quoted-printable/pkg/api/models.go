package api

type EncodeRequest struct {
	Text string `json:"text"`
}

type EncodeResponse struct {
	Encoded string `json:"encoded"`
	Error   string `json:"error,omitempty"`
}

type DecodeRequest struct {
	Encoded string `json:"encoded"`
}

type DecodeResponse struct {
	Decoded string `json:"decoded"`
	Error   string `json:"error,omitempty"`
}
