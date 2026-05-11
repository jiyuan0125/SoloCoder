package api

type CreateRequest struct {
	Width int `json:"width"`
	Depth int `json:"depth"`
	K     int `json:"k"`
}

type CreateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type AddRequest struct {
	Items []string `json:"items"`
}

type AddResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type TopKItem struct {
	Element   string `json:"element"`
	Freq      uint64 `json:"freq"`
	Uncertain bool   `json:"uncertain,omitempty"`
}

type TopKResponse struct {
	Success bool       `json:"success"`
	Items   []TopKItem `json:"items,omitempty"`
	Message string     `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
