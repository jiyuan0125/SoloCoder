package protocol

type DependencyNode struct {
	Name        string            `json:"name"`
	Version     string            `json:"version,omitempty"`
	IsStandard  bool              `json:"is_standard"`
	IsIndirect  bool              `json:"is_indirect,omitempty"`
	IsReused    bool              `json:"is_reused,omitempty"`
	IsRoot      bool              `json:"is_root,omitempty"`
	Children    []*DependencyNode `json:"children,omitempty"`
	ReusedRef   string            `json:"reused_ref,omitempty"`
}

type ProjectInfo struct {
	ModuleName     string           `json:"module_name"`
	Root           *DependencyNode  `json:"root"`
	CircularDeps   []CircularDep    `json:"circular_deps,omitempty"`
	TotalDeps      int              `json:"total_deps"`
	ThirdPartyDeps int              `json:"third_party_deps"`
}

type CircularDep struct {
	Path []string `json:"path"`
}

type FilterOptions struct {
	FilterStandard   bool   `json:"filter_standard,omitempty"`
	FilterThirdParty bool   `json:"filter_third_party,omitempty"`
	SearchPattern    string `json:"search_pattern,omitempty"`
}
