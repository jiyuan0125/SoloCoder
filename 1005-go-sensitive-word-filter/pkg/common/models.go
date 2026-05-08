package common

type MatchType string

const (
	DirectMatch  MatchType = "direct"
	SkipMatch    MatchType = "skip"
	WildcardMatch MatchType = "wildcard"
)

type Position struct {
	Start  int `json:"start"`
	Length int `json:"length"`
}

type HitResult struct {
	SensitiveWord string      `json:"sensitive_word"`
	MatchType     MatchType   `json:"match_type"`
	Positions     []Position  `json:"positions"`
}

type FilterRequest struct {
	Text string `json:"text"`
}

type FilterResponse struct {
	Text   string       `json:"text"`
	Hits   []HitResult  `json:"hits"`
	Hit    bool         `json:"hit"`
}

type AddSensitiveWordRequest struct {
	Word string `json:"word"`
}

type UpdateSensitiveWordRequest struct {
	OldWord string `json:"old_word"`
	NewWord string `json:"new_word"`
}

type DeleteSensitiveWordRequest struct {
	Word string `json:"word"`
}

type ListSensitiveWordsResponse struct {
	Words []string `json:"words"`
	Total int      `json:"total"`
}

type AddWhitelistRequest struct {
	Word string `json:"word"`
}

type DeleteWhitelistRequest struct {
	Word string `json:"word"`
}

type ListWhitelistResponse struct {
	Words []string `json:"words"`
	Total int      `json:"total"`
}

type DailyStats struct {
	Date      string `json:"date"`
	Total     int64  `json:"total"`
	HitCount  int64  `json:"hit_count"`
}

type TopSensitiveWord struct {
	Word  string `json:"word"`
	Count int64  `json:"count"`
}

type StatsResponse struct {
	DailyStats   []DailyStats       `json:"daily_stats"`
	TopHits      []TopSensitiveWord `json:"top_hits"`
	TotalTotal   int64              `json:"total_total"`
	TotalHits    int64              `json:"total_hits"`
}

type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
