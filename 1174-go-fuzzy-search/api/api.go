package api

type AddWordRequest struct {
	Word string `json:"word"`
}

type AddWordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type RemoveWordRequest struct {
	Word string `json:"word"`
}

type RemoveWordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ImportWordsRequest struct {
	Words []string `json:"words"`
}

type ImportWordsResponse struct {
	Success   bool   `json:"success"`
	Imported  int    `json:"imported"`
	Total     int    `json:"total"`
	Message   string `json:"message,omitempty"`
}

type SearchRequest struct {
	Query     string `json:"query"`
	Threshold int    `json:"threshold"`
}

type SearchResponse struct {
	Success bool           `json:"success"`
	Results []SearchResult `json:"results"`
	Message string         `json:"message,omitempty"`
}

type SearchResult struct {
	Word     string `json:"word"`
	Distance int    `json:"distance"`
}

type WildcardSearchRequest struct {
	Pattern   string `json:"pattern"`
	Threshold int    `json:"threshold"`
}

type WildcardSearchResponse struct {
	Success bool           `json:"success"`
	Results []SearchResult `json:"results"`
	Message string         `json:"message,omitempty"`
}

type GetDictionaryResponse struct {
	Success bool     `json:"success"`
	Size    int      `json:"size"`
	Words   []string `json:"words"`
	Message string   `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
