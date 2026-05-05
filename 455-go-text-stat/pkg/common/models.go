package common

type TextStatsRequest struct {
	Text            string            `json:"text"`
	CustomStopWords map[string]bool   `json:"custom_stop_words,omitempty"`
	TopN            int               `json:"top_n,omitempty"`
}

type TextStatsResponse struct {
	CharCountWithSpaces    int             `json:"char_count_with_spaces"`
	CharCountWithoutSpaces int             `json:"char_count_without_spaces"`
	WordCount              int             `json:"word_count"`
	LineCount              int             `json:"line_count"`
	ParagraphCount         int             `json:"paragraph_count"`
	TopWords               []TopWordItem   `json:"top_words"`
}

type TopWordItem struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

type SimilarityRequest struct {
	Text1           string          `json:"text1"`
	Text2           string          `json:"text2"`
	CustomStopWords map[string]bool `json:"custom_stop_words,omitempty"`
}

type SimilarityResponse struct {
	Similarity float64 `json:"similarity"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
