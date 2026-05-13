package types

type QRCodeRequest struct {
	Content      string `json:"content" form:"content"`
	Format       string `json:"format" form:"format"`
	Size         int    `json:"size" form:"size"`
	Foreground   string `json:"foreground" form:"foreground"`
	Background   string `json:"background" form:"background"`
	ErrorLevel   string `json:"error_level" form:"error_level"`
	LogoFile     []byte `json:"-"`
	LogoFilename string `json:"-"`
	LogoFormat   string `json:"-"`
	Border       int    `json:"border" form:"border"`
	Title        string `json:"title" form:"title"`
	ResourceID   int64  `json:"resource_id" form:"resource_id"`
}

type BatchResult struct {
	LineNumber   int    `json:"line_number"`
	Content      string `json:"content"`
	Filename     string `json:"filename"`
	Error        string `json:"error,omitempty"`
}

type BatchResponse struct {
	ZipFile []byte `json:"-"`
	Results []BatchResult `json:"results"`
	Total   int `json:"total"`
	Success int `json:"success"`
	Skipped []int `json:"skipped"`
}

type Resource struct {
	ID          int64  `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

type ResourceSummary struct {
	ResourceID    int64  `json:"resource_id"`
	ResourceName  string `json:"resource_name"`
	ResourceType  string `json:"resource_type"`
	TotalOps      int64  `json:"total_operations"`
	LastOpTime    string `json:"last_operation_time"`
}

type Operation struct {
	ID           int64  `json:"id"`
	ResourceID   int64  `json:"resource_id"`
	ResourceName string `json:"resource_name,omitempty"`
	OpType       string `json:"op_type"`
	Format       string `json:"format"`
	Count        int    `json:"count"`
	CreatedAt    string `json:"created_at"`
}
