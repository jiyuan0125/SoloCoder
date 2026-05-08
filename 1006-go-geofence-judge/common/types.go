package common

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type FenceCreateRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Points      []Point `json:"points"`
}

type FenceCreateResponse struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Area    float64 `json:"area"`
	Warning string  `json:"warning,omitempty"`
}

type FenceListResponse struct {
	Fences []FenceInfo `json:"fences"`
}

type FenceInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Area        float64 `json:"area"`
	PointCount  int     `json:"point_count"`
}

type JudgeRequest struct {
	Point   Point  `json:"point"`
	FenceID string `json:"fence_id,omitempty"`
}

type JudgeResult struct {
	In         bool     `json:"in"`
	FenceIDs   []string `json:"fence_ids,omitempty"`
	FenceNames []string `json:"fence_names,omitempty"`
}

type JudgeResponse struct {
	Point       Point         `json:"point"`
	Results     []JudgeResult `json:"results"`
	MinDistance float64       `json:"min_distance"`
}

type BatchJudgeRequest struct {
	Points []Point `json:"points"`
}

type BatchJudgeResponse struct {
	Results []PointResult `json:"results"`
}

type PointResult struct {
	Point       Point   `json:"point"`
	In          bool    `json:"in"`
	FenceIDs    []string `json:"fence_ids,omitempty"`
	FenceNames  []string `json:"fence_names,omitempty"`
	MinDistance float64 `json:"min_distance"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
