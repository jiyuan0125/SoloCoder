package common

type Granularity string

const (
	GranularityChar Granularity = "char"
	GranularityWord Granularity = "word"
)

type TrainRequest struct {
	Text        string      `json:"text"`
	Granularity Granularity `json:"granularity"`
	Order       int         `json:"order"`
}

type TrainResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type GenerateRequest struct {
	Text        string      `json:"text"`
	Granularity Granularity `json:"granularity"`
	Order       int         `json:"order"`
	Length      int         `json:"length"`
	Seed        int64       `json:"seed"`
}

type GenerateResponse struct {
	Success bool   `json:"success"`
	Text    string `json:"text,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
