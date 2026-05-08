package common

type GenerateRequest struct {
	Count int `json:"count,omitempty"`
}

type ParseRequest struct {
	ID    int64 `json:"id"`
	Epoch int64 `json:"epoch,omitempty"`
}
