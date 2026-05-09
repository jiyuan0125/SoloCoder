package common

type AllocateRequest struct {
}

type AllocateResponse struct {
	ID string `json:"id"`
}

type AllocateBatchRequest struct {
	Count int `json:"count"`
}

type AllocateBatchResponse struct {
	IDs []string `json:"ids"`
}

type StatusRequest struct {
}

type SegmentInfo struct {
	Start        int64 `json:"start"`
	End          int64 `json:"end"`
	Current      int64 `json:"current"`
	Allocated    int64 `json:"allocated"`
	Remaining    int64 `json:"remaining"`
}

type StatusResponse struct {
	CurrentSegment      *SegmentInfo `json:"currentSegment"`
	PreloadedSegment    *SegmentInfo `json:"preloadedSegment,omitempty"`
	HasPreloadedSegment bool         `json:"hasPreloadedSegment"`
	IsPreloading        bool         `json:"isPreloading"`
}

type CenterGetSegmentRequest struct {
	BusinessKey string `json:"businessKey"`
}

type CenterGetSegmentResponse struct {
	Success bool   `json:"success"`
	Start   int64  `json:"start"`
	End     int64  `json:"end"`
	Msg     string `json:"msg,omitempty"`
}

type CenterReportProgressRequest struct {
	BusinessKey string `json:"businessKey"`
	Current     int64  `json:"current"`
	Max         int64  `json:"max"`
}

type CenterReportProgressResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
