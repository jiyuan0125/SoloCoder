package api

type BuildRequest struct {
	Values []int `json:"values"`
}

type BuildResponse struct {
	Success      bool   `json:"success"`
	VersionCount int    `json:"version_count,omitempty"`
	Message      string `json:"message,omitempty"`
}

type QueryKthRequest struct {
	L int `json:"l"`
	R int `json:"r"`
	K int `json:"k"`
}

type QueryKthResponse struct {
	Success bool   `json:"success"`
	Value   int    `json:"value,omitempty"`
	Message string `json:"message,omitempty"`
}

type StatusResponse struct {
	Ready       bool `json:"ready"`
	ArrayLength int  `json:"array_length,omitempty"`
}
