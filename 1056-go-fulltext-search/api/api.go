package api

type AddDocRequest struct {
	DocID   string `json:"doc_id"`
	Content string `json:"content"`
}

type AddDocResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type DeleteDocRequest struct {
	DocID string `json:"doc_id"`
}

type DeleteDocResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type SearchRequest struct {
	Query string `json:"query"`
}

type SearchResultItem struct {
	DocID string  `json:"doc_id"`
	Score float64 `json:"score"`
}

type SearchResponse struct {
	Success bool               `json:"success"`
	Results []SearchResultItem `json:"results"`
	Message string             `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
