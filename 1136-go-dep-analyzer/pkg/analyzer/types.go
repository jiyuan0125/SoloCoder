package analyzer

type GoModFile struct {
	Module    string
	GoVersion string
	Requires  []Require
	Excludes  []Exclude
	Replaces  []Replace
	Retracts  []Retract
}

type Require struct {
	Path       string
	Version    string
	IsIndirect bool
}

type Exclude struct {
	Path    string
	Version string
}

type Replace struct {
	OldPath    string
	OldVersion string
	NewPath    string
	NewVersion string
	IsLocal    bool
}

type Retract struct {
	Version string
}

type GoSumEntry struct {
	Module  string
	Version string
	Hash    string
	IsMod   bool
}

type DependencyNode struct {
	Path       string
	Version    string
	IsRoot     bool
	IsIndirect bool
}

type DependencyEdge struct {
	From     string
	To       string
	Version  string
	IsDirect bool
}

type DependencyGraph struct {
	Nodes map[string]*DependencyNode
	Edges []DependencyEdge
	Replaces map[string]Replace
}

type VersionConflict struct {
	Module          string
	RequiredVersion string
	ActualVersion   string
	Reason          string
}

type SlimmingSuggestion struct {
	Module     string
	Suggestion string
}

type FullAnalysis struct {
	Graph              *DependencyGraph
	Cycles             [][]string
	Topology           []string
	Depths             map[string]int
	Conflicts          []VersionConflict
	SlimmingSuggestions []SlimmingSuggestion
}
