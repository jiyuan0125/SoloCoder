package api

// AddressInfo 解析后的邮箱地址信息
type AddressInfo struct {
	Original      string `json:"original"`
	Local         string `json:"local"`
	Domain        string `json:"domain"`
	HasAlias      bool   `json:"has_alias"`
	BaseLocal     string `json:"base_local"`
	BaseAddress   string `json:"base_address"`
	IsQuotedLocal bool   `json:"is_quoted_local"`
}

// ParseRequest 单个邮箱解析请求
type ParseRequest struct {
	Email string `json:"email"`
}

// ParseResponse 单个邮箱解析响应
type ParseResponse struct {
	Success bool         `json:"success"`
	Valid   bool         `json:"valid,omitempty"`
	Address *AddressInfo `json:"address,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// BatchParseRequest 批量解析请求
type BatchParseRequest struct {
	Emails string `json:"emails"`
}

// BatchParseResult 单个邮箱的批量解析结果
type BatchParseResult struct {
	Email   string       `json:"email"`
	Valid   bool         `json:"valid"`
	Address *AddressInfo `json:"address,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// BatchParseResponse 批量解析响应
type BatchParseResponse struct {
	Success bool                `json:"success"`
	Total   int                 `json:"total"`
	Valid   int                 `json:"valid"`
	Invalid int                 `json:"invalid"`
	Results []BatchParseResult  `json:"results"`
	Error   string              `json:"error,omitempty"`
}

// ValidateRequest 校验请求
type ValidateRequest struct {
	Email string `json:"email"`
}

// ValidateResponse 校验响应
type ValidateResponse struct {
	Success bool `json:"success"`
	Valid   bool `json:"valid"`
}
