package api

type EncodeRequest struct {
	Domain string `json:"domain"`
}

type EncodeResponse struct {
	Domain string `json:"domain"`
	Error  string `json:"error,omitempty"`
}

type DecodeRequest struct {
	Domain string `json:"domain"`
}

type DecodeResponse struct {
	Domain string `json:"domain"`
	Error  string `json:"error,omitempty"`
}
