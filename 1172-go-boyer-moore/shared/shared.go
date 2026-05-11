package shared

type SearchRequest struct {
	Text    string `json:"text"`
	Pattern string `json:"pattern"`
}

type SearchResponse struct {
	Pattern string `json:"pattern"`
	TextLen int    `json:"text_length"`
	Count   int    `json:"count"`
	Matches []int  `json:"matches"`
}

type PreprocessRequest struct {
	Pattern string `json:"pattern"`
}

type PreprocessResponse struct {
	Pattern string `json:"pattern"`
	PatternLen int `json:"pattern_length"`
	BmBc    map[rune]int `json:"bm_bc,omitempty"`
	BmGs    []int `json:"bm_gs"`
}

type BenchmarkResult struct {
	Pattern string
	Count   int
	NsOp    int64
}
