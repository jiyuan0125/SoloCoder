package protocol

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type FAQDetailResponse struct {
	FAQ
	CategoryName string `json:"category_name"`
}

type SearchResultItem struct {
	FAQ
	CategoryName  string  `json:"category_name"`
	MatchScore    float64 `json:"match_score"`
}

type SearchFAQResponse struct {
	Items      []SearchResultItem `json:"items"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

type BatchImportResult struct {
	SuccessCount int                    `json:"success_count"`
	FailedCount  int                    `json:"failed_count"`
	FailedItems  []BatchImportFailedItem `json:"failed_items,omitempty"`
}

type BatchImportFailedItem struct {
	Index int    `json:"index"`
	Item  string `json:"item"`
	Error string `json:"error"`
}

type CategoryTreeResponse struct {
	Categories []Category `json:"categories"`
}

type HistoryListResponse struct {
	Records []HistoryRecord `json:"records"`
	Total   int             `json:"total"`
}

type FAQStatistics struct {
	TotalFAQ      int `json:"total_faq"`
	EnabledFAQ    int `json:"enabled_faq"`
	HotFAQ        int `json:"hot_faq"`
	NeedOptimize  int `json:"need_optimize"`
	TotalViews    int `json:"total_views"`
	TotalClicks   int `json:"total_clicks"`
}
