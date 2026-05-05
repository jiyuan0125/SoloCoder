package protocol

type CreateFAQRequest struct {
	CategoryID string                  `json:"category_id"`
	Content    map[Language]FAQContent `json:"content"`
}

type UpdateFAQRequest struct {
	Content map[Language]FAQContent `json:"content"`
}

type SearchFAQRequest struct {
	Keyword   string   `json:"keyword"`
	Language  Language `json:"language"`
	Page      int      `json:"page"`
	PageSize  int      `json:"page_size"`
}

type BatchImportFAQItem struct {
	CategoryID string                  `json:"category_id"`
	Content    map[Language]FAQContent `json:"content"`
}

type BatchImportFAQRequest struct {
	Items []BatchImportFAQItem `json:"items"`
}

type CreateCategoryRequest struct {
	ParentID string `json:"parent_id"`
	Name     string `json:"name"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name"`
}

type RecordClickRequest struct {
	UserID    string `json:"user_id"`
	IsHelpful bool   `json:"is_helpful"`
}

type PinFAQRequest struct {
	IsPinned bool `json:"is_pinned"`
}

type EnableFAQRequest struct {
	IsEnabled bool `json:"is_enabled"`
}
