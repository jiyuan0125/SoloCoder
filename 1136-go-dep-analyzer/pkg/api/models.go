package api

type UploadRequest struct {
	GoModContent string `json:"go_mod_content"`
	GoSumContent string `json:"go_sum_content"`
}

type UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DependencyGraphResponse struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type GraphNode struct {
	Module   string `json:"module"`
	Version  string `json:"version"`
	IsRoot   bool   `json:"is_root"`
	IsIndirect bool `json:"is_indirect"`
}

type GraphEdge struct {
	From       string `json:"from"`
	To         string `json:"to"`
	IsDirect   bool   `json:"is_direct"`
	FromModule string `json:"from_module"`
	ToModule   string `json:"to_module"`
}

type CyclesResponse struct {
	Cycles []Cycle `json:"cycles"`
}

type Cycle struct {
	Modules []string `json:"modules"`
}

type TopologyResponse struct {
	Modules []string `json:"modules"`
}

type DepthAnalysisResponse struct {
	ModuleDepths []ModuleDepth `json:"module_depths"`
}

type ModuleDepth struct {
	Module string `json:"module"`
	Depth  int    `json:"depth"`
}

type VersionConflictResponse struct {
	Conflicts []VersionConflict `json:"conflicts"`
}

type VersionConflict struct {
	Module          string `json:"module"`
	RequiredVersion string `json:"required_version"`
	ActualVersion   string `json:"actual_version"`
	Reason          string `json:"reason"`
}

type DependencyChainRequest struct {
	TargetModule string `json:"target_module"`
}

type DependencyChainResponse struct {
	TargetModule string            `json:"target_module"`
	Chains       []DependencyChain `json:"chains"`
}

type DependencyChain struct {
	Path []ChainNode `json:"path"`
}

type ChainNode struct {
	Module  string `json:"module"`
	Version string `json:"version"`
}

type FullReportResponse struct {
	TotalDeps          int                 `json:"total_deps"`
	DirectDeps         int                 `json:"direct_deps"`
	IndirectDeps       int                 `json:"indirect_deps"`
	Cycles             []Cycle             `json:"cycles"`
	Conflicts          []VersionConflict   `json:"conflicts"`
	DeepestModules     []ModuleDepth       `json:"deepest_modules"`
	SlimmingSuggestions []SlimmingSuggestion `json:"slimming_suggestions"`
}

type SlimmingSuggestion struct {
	Module      string `json:"module"`
	Suggestion  string `json:"suggestion"`
}
