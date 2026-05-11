package api

type ParseRequest struct {
	Format   string `json:"format"`
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type ParseResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ValidateRequest struct {
	Format   string `json:"format"`
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type ValidateError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidateResponse struct {
	Success bool            `json:"success"`
	Valid   bool            `json:"valid"`
	Errors  []ValidateError `json:"errors,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type ListPathsRequest struct {
	Format   string `json:"format"`
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type PathInfo struct {
	Path    string   `json:"path"`
	Methods []string `json:"methods"`
}

type ListPathsResponse struct {
	Success bool       `json:"success"`
	Paths   []PathInfo `json:"paths,omitempty"`
	Error   string     `json:"error,omitempty"`
}

type ResolveRequest struct {
	Format   string `json:"format"`
	Filename string `json:"filename"`
	Content  string `json:"content"`
	Ref      string `json:"ref"`
	All      bool   `json:"all,omitempty"`
}

type ResolveResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
