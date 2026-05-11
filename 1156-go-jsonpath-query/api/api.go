package api

type QueryRequest struct {
	JSON string `json:"json"`
	Path string `json:"path"`
}

type QueryResponse struct {
	Results    interface{} `json:"results,omitempty"`
	MatchCount int         `json:"matchCount"`
	Paths      []string    `json:"paths,omitempty"`
	Error      string      `json:"error,omitempty"`
}
