package api

type InsertRequest struct {
	Key int `json:"key"`
}

type InsertResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type DeleteRequest struct {
	Key int `json:"key"`
}

type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type RangeQueryRequest struct {
	Low  int `json:"low"`
	High int `json:"high"`
}

type RangeQueryResponse struct {
	Success bool   `json:"success"`
	Values  []int  `json:"values"`
	Message string `json:"message,omitempty"`
}

type KthLargestRequest struct {
	K int `json:"k"`
}

type KthLargestResponse struct {
	Success bool   `json:"success"`
	Value   int    `json:"value,omitempty"`
	Message string `json:"message,omitempty"`
}

type SizeResponse struct {
	Success bool   `json:"success"`
	Size    int    `json:"size"`
	Message string `json:"message,omitempty"`
}
