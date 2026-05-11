package common

type TableInfo struct {
	Headers    []string   `json:"headers"`
	Rows       [][]string `json:"rows"`
	Alignments []string   `json:"alignments"`
}

type ExtractTablesRequest struct {
	Markdown string `json:"markdown"`
}

type ExtractTablesResponse struct {
	Tables []TableInfo `json:"tables"`
}

type MarkdownToCSVRequest struct {
	Markdown string `json:"markdown"`
}

type MarkdownToCSVResponse struct {
	CSVContent string `json:"csv_content"`
	TableCount int    `json:"table_count"`
}

type CSVToMarkdownRequest struct {
	CSVContent string `json:"csv_content"`
	HasHeader  bool   `json:"has_header"`
}

type CSVToMarkdownResponse struct {
	Markdown string `json:"markdown"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
