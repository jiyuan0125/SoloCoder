package common

type StrengthLevel string

const (
	LevelWeak     StrengthLevel = "弱"
	LevelMedium   StrengthLevel = "中"
	LevelStrong   StrengthLevel = "强"
	LevelVeryStrong StrengthLevel = "很强"
)

type DimensionScore struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
	Max   float64 `json:"max"`
}

type EntropyInfo struct {
	Theoretical float64 `json:"theoretical"`
	Estimated   float64 `json:"estimated"`
}

type EvaluateRequest struct {
	Password string `json:"password"`
}

type EvaluateResponse struct {
	TotalScore   float64         `json:"total_score"`
	Level        StrengthLevel   `json:"level"`
	Dimensions   []DimensionScore `json:"dimensions"`
	Entropy      EntropyInfo     `json:"entropy"`
	Suggestions  []string        `json:"suggestions"`
	IsCommon     bool            `json:"is_common"`
}

type BatchSummary struct {
	TotalCount       int     `json:"total_count"`
	WeakCount        int     `json:"weak_count"`
	MediumCount      int     `json:"medium_count"`
	StrongCount      int     `json:"strong_count"`
	VeryStrongCount  int     `json:"very_strong_count"`
	AverageScore     float64 `json:"average_score"`
	StrongestExample string  `json:"strongest_example"`
	WeakestExample   string  `json:"weakest_example"`
}
