package protocol

type GenerateRequest struct {
	Prefix string `json:"prefix,omitempty"`
}

type GenerateResponse struct {
	Success bool   `json:"success"`
	OrderNo string `json:"order_no,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ParseRequest struct {
	OrderNo string `json:"order_no"`
}

type ParseResponse struct {
	Success   bool   `json:"success"`
	Prefix    string `json:"prefix,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	SerialNum int    `json:"serial_num,omitempty"`
	RandomNum int    `json:"random_num,omitempty"`
	Error     string `json:"error,omitempty"`
}

type SetPrefixRequest struct {
	Prefix string `json:"prefix"`
}

type SetPrefixResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type GetPrefixResponse struct {
	Success bool   `json:"success"`
	Prefix  string `json:"prefix,omitempty"`
	Error   string `json:"error,omitempty"`
}
