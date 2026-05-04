package common

type QueryRequest struct {
	IP string `json:"ip"`
}

type QueryResponse struct {
	Success  bool   `json:"success"`
	IP       string `json:"ip"`
	Country  string `json:"country,omitempty"`
	Province string `json:"province,omitempty"`
	City     string `json:"city,omitempty"`
	Error    string `json:"error,omitempty"`
}

type BatchQueryRequest struct {
	IPs []string `json:"ips"`
}

type BatchQueryResponse struct {
	Success  bool            `json:"success"`
	Results  []IPQueryResult `json:"results,omitempty"`
	Error    string          `json:"error,omitempty"`
}

type IPQueryResult struct {
	IP       string `json:"ip"`
	Country  string `json:"country,omitempty"`
	Province string `json:"province,omitempty"`
	City     string `json:"city,omitempty"`
	Error    string `json:"error,omitempty"`
}

type SameProvinceRequest struct {
	IP1 string `json:"ip1"`
	IP2 string `json:"ip2"`
}

type SameProvinceResponse struct {
	Success bool   `json:"success"`
	Same    bool   `json:"same"`
	Error   string `json:"error,omitempty"`
}

type IPInfoRequest struct {
	IP string `json:"ip"`
}

type IPInfoResponse struct {
	Success   bool   `json:"success"`
	IP        string `json:"ip"`
	IsPrivate bool   `json:"is_private"`
	Type      string `json:"type"`
	Error     string `json:"error,omitempty"`
}
