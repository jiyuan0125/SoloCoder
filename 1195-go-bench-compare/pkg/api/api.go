package api

type CompareRequest struct {
	Baseline string `json:"baseline"`
	Current  string `json:"current"`
}

type ChangeStatus string

const (
	StatusAdded      ChangeStatus = "added"
	StatusRemoved    ChangeStatus = "removed"
	StatusUnchanged  ChangeStatus = "unchanged"
	StatusChanged    ChangeStatus = "changed"
	StatusRegression ChangeStatus = "regression"
	StatusImprovement ChangeStatus = "improvement"
)

type BenchmarkResult struct {
	Name       string  `json:"name"`
	Iterations int     `json:"iterations"`
	NsPerOp    float64 `json:"ns_per_op"`
	BytesPerOp int     `json:"bytes_per_op"`
	AllocsPerOp int    `json:"allocs_per_op"`
	Uncertain  bool    `json:"uncertain"`
}

type ComparisonItem struct {
	Name            string       `json:"name"`
	Status          ChangeStatus `json:"status"`
	Baseline        *BenchmarkResult `json:"baseline,omitempty"`
	Current         *BenchmarkResult `json:"current,omitempty"`
	NsChangePercent float64      `json:"ns_change_percent"`
	MemChangePercent float64     `json:"mem_change_percent"`
}

type CompareResponse struct {
	Items []ComparisonItem `json:"items"`
}
