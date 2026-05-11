package common

type EncodeRequest struct {
	Text string `json:"text"`
}

type EncodeResponse struct {
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}

type DecodeRequest struct {
	Text string `json:"text"`
}

type DecodeResponse struct {
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}

type BatchRequest struct {
	Operation string   `json:"operation"`
	Texts     []string `json:"texts"`
}

type BatchResponse struct {
	Results []string `json:"results,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}
