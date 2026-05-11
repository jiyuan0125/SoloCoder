package api

type EncodeRequest struct {
	Data     string `json:"data"`
	Filename string `json:"filename"`
	Mode     int    `json:"mode"`
}

type EncodeResponse struct {
	Encoded string `json:"encoded"`
}

type DecodeRequest struct {
	Encoded string `json:"encoded"`
}

type DecodeResponse struct {
	Data     string `json:"data"`
	Filename string `json:"filename"`
	Mode     int    `json:"mode"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
