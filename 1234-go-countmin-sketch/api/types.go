package api

type CreateRequest struct {
	Name  string `json:"name"`
	Width int    `json:"width,omitempty"`
	Depth int    `json:"depth,omitempty"`
}

type CreateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type AddItem struct {
	Item  string `json:"item"`
	Count int64  `json:"count"`
}

type AddRequest struct {
	Name  string    `json:"name"`
	Items []AddItem `json:"items"`
}

type AddResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type QueryRequest struct {
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

type QueryResult struct {
	Item       string  `json:"item"`
	Frequency  int64   `json:"frequency"`
	ErrorBound float64 `json:"error_bound"`
}

type QueryResponse struct {
	Success bool          `json:"success"`
	Results []QueryResult `json:"results,omitempty"`
	Error   string        `json:"error,omitempty"`
}

type MergeRequest struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type MergeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ResetRequest struct {
	Name string `json:"name"`
}

type ResetResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type StatsRequest struct {
	Name string `json:"name,omitempty"`
}

type SketchStats struct {
	Name       string `json:"name"`
	Width      int    `json:"width"`
	Depth      int    `json:"depth"`
	TotalCount int64  `json:"total_count"`
}

type StatsResponse struct {
	Success bool         `json:"success"`
	Sketches []SketchStats `json:"sketches,omitempty"`
	Error   string       `json:"error,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
