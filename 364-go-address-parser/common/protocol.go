package common

type ParseRequest struct {
	Address string `json:"address"`
}

type ParseResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message,omitempty"`
	Province string `json:"province,omitempty"`
	City     string `json:"city,omitempty"`
	District string `json:"district,omitempty"`
	Detail   string `json:"detail,omitempty"`
}
