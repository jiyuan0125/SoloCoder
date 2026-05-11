package api

type NextRequest struct {
	BizType string `json:"biz_type"`
}

type NextResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Serial  string `json:"serial,omitempty"`
}

type BatchRequest struct {
	BizType string `json:"biz_type"`
	Count   int    `json:"count"`
}

type BatchResponse struct {
	Success bool     `json:"success"`
	Error   string   `json:"error,omitempty"`
	Start   string   `json:"start,omitempty"`
	End     string   `json:"end,omitempty"`
	Serials []string `json:"serials,omitempty"`
}

type StatusRequest struct {
	BizType string `json:"biz_type"`
}

type StatusResponse struct {
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
	BizType     string `json:"biz_type,omitempty"`
	Date        string `json:"date,omitempty"`
	CurrentMax  int64  `json:"current_max,omitempty"`
	TotalAllocated int64 `json:"total_allocated,omitempty"`
}

type CheckRequest struct {
	BizType string `json:"biz_type"`
}

type CheckResponse struct {
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	HasGap     bool   `json:"has_gap,omitempty"`
	Gaps       []Gap  `json:"gaps,omitempty"`
	CurrentMax int64  `json:"current_max,omitempty"`
	TotalAllocated int64 `json:"total_allocated,omitempty"`
}

type Gap struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type ResetRequest struct {
	BizType string `json:"biz_type"`
	Confirm bool   `json:"confirm"`
}

type ResetResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Date    string `json:"date,omitempty"`
}
