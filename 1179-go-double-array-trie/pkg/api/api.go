package api

type AddRequest struct {
	Word string `json:"word"`
}

type AddResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ImportRequest struct {
	Words []string `json:"words"`
}

type ImportResponse struct {
	Success bool   `json:"success"`
	Added   int    `json:"added"`
	Message string `json:"message,omitempty"`
}

type SearchRequest struct {
	Word string `json:"word"`
}

type SearchResponse struct {
	Exists bool   `json:"exists"`
	Message string `json:"message,omitempty"`
}

type PrefixRequest struct {
	Prefix string `json:"prefix"`
}

type PrefixResponse struct {
	Success bool     `json:"success"`
	Words   []string `json:"words"`
	Message string   `json:"message,omitempty"`
}

type DeleteRequest struct {
	Word string `json:"word"`
}

type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type StatResponse struct {
	BaseSize    int     `json:"base_size"`
	CheckSize   int     `json:"check_size"`
	UsedSlots   int     `json:"used_slots"`
	WordCount   int     `json:"word_count"`
	Utilization float64 `json:"utilization"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
