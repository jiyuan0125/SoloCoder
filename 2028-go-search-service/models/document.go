package models

type Document struct {
	ID           int                    `json:"id"`
	Title        string                 `json:"title"`
	Body         string                 `json:"body"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	CreatedAt    string                 `json:"created_at,omitempty"`
	UpdatedAt    string                 `json:"updated_at,omitempty"`
}

type DocumentIndexRequest struct {
	Title        string                 `json:"title" binding:"required"`
	Body         string                 `json:"body"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

type SearchRequest struct {
	Keyword string `form:"q" json:"q"`
	Field   string `form:"field" json:"field"`
}

type SearchResult struct {
	Document
	Score       float64  `json:"score"`
	Highlights  []string `json:"highlights,omitempty"`
}

type HotKeyword struct {
	Keyword string `json:"keyword"`
	Count   int    `json:"count"`
}
