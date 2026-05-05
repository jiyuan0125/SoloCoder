package common

type IndexRequest struct {
	Directory string `json:"directory"`
}

type IndexResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type SearchRequest struct {
	Query string `json:"query"`
}

type SearchHit struct {
	FileName    string `json:"file_name"`
	LineNumber  int    `json:"line_number"`
	LineContent string `json:"line_content"`
	Score       int    `json:"score"`
}

type SearchResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Hits    []SearchHit `json:"hits"`
}

type StatsResponse struct {
	Success         bool   `json:"success"`
	IndexedFiles    int    `json:"indexed_files"`
	IndexedWords    int    `json:"indexed_words"`
	LastIndexedTime string `json:"last_indexed_time"`
}
