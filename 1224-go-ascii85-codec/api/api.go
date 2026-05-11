package api

type EncodeRequest struct {
	Data string `json:"data"`
	Mode string `json:"mode,omitempty"`
}

type EncodeResponse struct {
	Encoded string `json:"encoded"`
	Mode    string `json:"mode"`
}

type DecodeRequest struct {
	Encoded string `json:"encoded"`
}

type DecodeResponse struct {
	Decoded string `json:"decoded"`
	Mode    string `json:"mode"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
