package api

type BookDTO struct {
	ISBN         string   `json:"isbn"`
	Title        string   `json:"title"`
	Authors      []string `json:"authors"`
	Publisher    string   `json:"publisher"`
	Year         int      `json:"year"`
	CategoryCode string   `json:"category_code"`
}

type QueryRequest struct {
	ISBN         string `json:"isbn"`
	Title        string `json:"title"`
	Authors      string `json:"authors"`
	Publisher    string `json:"publisher"`
	Page         int    `json:"page"`
	PageSize     int    `json:"page_size"`
}

type QueryResponse struct {
	Success     bool      `json:"success"`
	Total       int       `json:"total"`
	TotalPages  int       `json:"total_pages"`
	CurrentPage int       `json:"current_page"`
	PageSize    int       `json:"page_size"`
	Books       []BookDTO `json:"books"`
	Message     string    `json:"message,omitempty"`
}

type LoadRequest struct {
	FilePath string `json:"file_path"`
}

type LoadResponse struct {
	Success        bool              `json:"success"`
	TotalRecords   int               `json:"total_records"`
	LoadedRecords  int               `json:"loaded_records"`
	InvalidRecords []InvalidRecordDTO `json:"invalid_records,omitempty"`
	Warnings       []string          `json:"warnings,omitempty"`
	Message        string            `json:"message,omitempty"`
}

type InvalidRecordDTO struct {
	LineNumber int    `json:"line_number"`
	Line       string `json:"line"`
	Error      string `json:"error"`
}

type AddBookRequest struct {
	ISBN         string `json:"isbn"`
	Title        string `json:"title"`
	Authors      string `json:"authors"`
	Publisher    string `json:"publisher"`
	Year         int    `json:"year"`
	CategoryCode string `json:"category_code"`
}

type AddBookResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
