package apimodels

type ProfileType string

const (
	ProfileTypeHeap  ProfileType = "heap"
	ProfileTypeCPU   ProfileType = "cpu"
	ProfileTypeOther ProfileType = "other"
)

type FunctionInfo struct {
	FullName    string  `json:"full_name"`
	PackageName string  `json:"package_name"`
	FuncName    string  `json:"func_name"`
	Bytes       int64   `json:"bytes"`
	Objects     int64   `json:"objects"`
	Percentage  float64 `json:"percentage"`
}

type AnalysisResponse struct {
	ProfileType  ProfileType    `json:"profile_type"`
	TotalBytes   int64          `json:"total_bytes"`
	TotalObjects int64          `json:"total_objects"`
	Functions    []FunctionInfo `json:"functions"`
}

type HistoryRecord struct {
	ID        string       `json:"id"`
	FileName  string       `json:"file_name"`
	ProfileType ProfileType `json:"profile_type"`
	TotalBytes int64       `json:"total_bytes"`
	AnalyzedAt string      `json:"analyzed_at"`
}

type HistoryResponse struct {
	Records []HistoryRecord `json:"records"`
}
