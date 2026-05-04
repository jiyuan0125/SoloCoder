package api

type ParseRequest struct {
	Version string `json:"version"`
}

type ParseResponse struct {
	Success    bool     `json:"success"`
	Version    *Version `json:"version,omitempty"`
	Error      string   `json:"error,omitempty"`
}

type Version struct {
	Major      int      `json:"major"`
	Minor      int      `json:"minor"`
	Patch      int      `json:"patch"`
	PreRelease []string `json:"pre_release,omitempty"`
	String     string   `json:"string"`
}

type CompareRequest struct {
	Version1 string `json:"version1"`
	Version2 string `json:"version2"`
}

type CompareResponse struct {
	Success bool   `json:"success"`
	Result  int    `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

type RangeCheckRequest struct {
	Version string `json:"version"`
	Range   string `json:"range"`
}

type RangeCheckResponse struct {
	Success bool   `json:"success"`
	InRange bool   `json:"in_range,omitempty"`
	Error   string `json:"error,omitempty"`
}

type IncrementRequest struct {
	Version string `json:"version"`
	Part    string `json:"part"`
}

type IncrementResponse struct {
	Success bool     `json:"success"`
	Version *Version `json:"version,omitempty"`
	Error   string   `json:"error,omitempty"`
}
