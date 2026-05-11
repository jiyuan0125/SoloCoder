package api

type SubmitRequest struct {
	Values []float64 `json:"values"`
}

type SubmitResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type QueryResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    []ElementFreq `json:"data"`
}

type ElementFreq struct {
	Element float64 `json:"element"`
	Freq    int     `json:"freq"`
}

type AdjustKRequest struct {
	K int `json:"k"`
}

type AdjustKResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ClearResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
