package models

type PushRequest struct {
	Value float64 `json:"value"`
}

type PushResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type MedianResponse struct {
	Success bool    `json:"success"`
	Median  *float64 `json:"median,omitempty"`
	Message string  `json:"message,omitempty"`
}

type SetWindowSizeRequest struct {
	Size int `json:"size"`
}

type SetWindowSizeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ResetResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type StatusResponse struct {
	Success    bool    `json:"success"`
	WindowSize int     `json:"window_size"`
	DataCount  int     `json:"data_count"`
	Median     *float64 `json:"median,omitempty"`
}
