package api

type AnalyzeRequest struct {
	GoModContent string `json:"go_mod"`
	GoSumContent string `json:"go_sum"`
}

type AnalyzeResponse struct {
	Module       string           `json:"module"`
	GoVersion    string           `json:"go_version"`
	Dependencies []DependencyInfo `json:"dependencies"`
	MVSResult    MVSResultInfo    `json:"mvs_result"`
	Graph        GraphInfo        `json:"graph"`
	VerifyResult VerifyResultInfo `json:"verify_result"`
	Error        string           `json:"error,omitempty"`
}

type DependencyInfo struct {
	Path     string `json:"path"`
	Version  string `json:"version"`
	Indirect bool   `json:"indirect"`
}

type MVSResultInfo struct {
	Selected      map[string]string      `json:"selected"`
	Conflicts     []ConflictInfo         `json:"conflicts,omitempty"`
	Retracted     []RetractedInfo        `json:"retracted,omitempty"`
	AllVersions   map[string][]string    `json:"all_versions,omitempty"`
}

type ConflictInfo struct {
	Path          string            `json:"path"`
	Selected      string            `json:"selected"`
	RequiredBy    map[string]string `json:"required_by"`
}

type RetractedInfo struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	Reason  string `json:"reason"`
}

type GraphInfo struct {
	Root  string       `json:"root"`
	Nodes []NodeInfo   `json:"nodes"`
	Edges []EdgeInfo   `json:"edges"`
}

type NodeInfo struct {
	ID       string `json:"id"`
	Path     string `json:"path"`
	Version  string `json:"version,omitempty"`
	IsDirect bool   `json:"is_direct"`
}

type EdgeInfo struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type VerifyResultInfo struct {
	AllChecked bool           `json:"all_checked"`
	Orphans    []OrphanInfo   `json:"orphans,omitempty"`
	Mismatches []MismatchInfo `json:"mismatches,omitempty"`
}

type OrphanInfo struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	Hash    string `json:"hash,omitempty"`
}

type MismatchInfo struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	Error   string `json:"error"`
}

type WhyRequest struct {
	GoModContent string `json:"go_mod"`
	GoSumContent string `json:"go_sum"`
	TargetPath   string `json:"target_path"`
}

type WhyResponse struct {
	Found  bool       `json:"found"`
	Target string     `json:"target"`
	Chains [][]string `json:"chains,omitempty"`
	Error  string     `json:"error,omitempty"`
}

type GraphRequest struct {
	GoModContent string `json:"go_mod"`
	GoSumContent string `json:"go_sum"`
}

type GraphResponse struct {
	Graph GraphInfo `json:"graph"`
	Error string    `json:"error,omitempty"`
}

type VerifyRequest struct {
	GoModContent string `json:"go_mod"`
	GoSumContent string `json:"go_sum"`
}

type VerifyResponse struct {
	Result VerifyResultInfo `json:"result"`
	Error  string           `json:"error,omitempty"`
}
