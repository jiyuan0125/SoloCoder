package api

type AddDataRequest struct {
	Data []int64 `json:"data"`
}

type AddDataResponse struct {
	Success   bool  `json:"success"`
	TotalRows int64 `json:"total_rows"`
}

type SetMemoryLimitRequest struct {
	Limit int `json:"limit"`
}

type SetMemoryLimitResponse struct {
	Success bool `json:"success"`
	Limit   int  `json:"limit"`
}

type SetMergeWaysRequest struct {
	Ways int `json:"ways"`
}

type SetMergeWaysResponse struct {
	Success bool `json:"success"`
	Ways    int  `json:"ways"`
}

type SortResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type StatsResponse struct {
	Success       bool    `json:"success"`
	TotalRecords  int64   `json:"total_records"`
	ChunksCreated int     `json:"chunks_created"`
	MergePasses   int     `json:"merge_passes"`
	DiskReads     int64   `json:"disk_reads"`
	DiskWrites    int64   `json:"disk_writes"`
	ChunkSizes    []int   `json:"chunk_sizes"`
	MemoryLimit   int     `json:"memory_limit"`
	MergeWays     int     `json:"merge_ways"`
	IsSorted      bool    `json:"is_sorted"`
	MergeDetails  [][]int `json:"merge_details,omitempty"`
}

type GetResultResponse struct {
	Success bool    `json:"success"`
	Data    []int64 `json:"data,omitempty"`
	Message string  `json:"message,omitempty"`
}

type GetConfigResponse struct {
	Success     bool `json:"success"`
	MemoryLimit int  `json:"memory_limit"`
	MergeWays   int  `json:"merge_ways"`
}

type ResetResponse struct {
	Success bool `json:"success"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
