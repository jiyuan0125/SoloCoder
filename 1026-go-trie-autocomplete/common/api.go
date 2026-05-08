package common

type SearchRequest struct {
	Prefix string `json:"prefix"`
	Limit  int    `json:"limit"`
}

type SearchResponse struct {
	Words []string `json:"words"`
}

type AddRequest struct {
	Word string `json:"word"`
}

type DeleteRequest struct {
	Word string `json:"word"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
