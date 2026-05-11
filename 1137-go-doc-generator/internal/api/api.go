package api

type UploadRequest struct {
	Files       []FileUpload `json:"files"`
	IncludeUnexported bool   `json:"include_unexported"`
}

type FileUpload struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Package string `json:"package"`
}

type SearchRequest struct {
	Query           string `json:"query"`
	IncludeUnexported bool `json:"include_unexported"`
}

type SearchResponse struct {
	Success  bool           `json:"success"`
	Message  string         `json:"message"`
	Results  []SearchResult `json:"results"`
}

type SearchResult struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Comment  string `json:"comment"`
	Summary  string `json:"summary"`
	Exported bool   `json:"exported"`
}

type ExportRequest struct {
	Format          string `json:"format"`
	IncludeUnexported bool `json:"include_unexported"`
}

type ExportResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Format  string `json:"format"`
	Content string `json:"content"`
}

type PackageResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type TypeRequest struct {
	Name            string `json:"name"`
	IncludeUnexported bool `json:"include_unexported"`
}

type FunctionRequest struct {
	Name            string `json:"name"`
	IncludeUnexported bool `json:"include_unexported"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
