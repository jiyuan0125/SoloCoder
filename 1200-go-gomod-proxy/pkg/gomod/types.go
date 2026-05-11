package gomod

import "fmt"

type GoModFile struct {
	Module   string
	GoVersion string
	Requires []Require
	Replaces []Replace
	Excludes []Exclude
	Retracts []Retract
}

type Require struct {
	Path     string
	Version  string
	Indirect bool
}

type Replace struct {
	OldPath    string
	OldVersion string
	NewPath    string
	NewVersion string
}

type Exclude struct {
	Path    string
	Version string
}

type Retract struct {
	LowVersion  string
	HighVersion string
}

type GoSumEntry struct {
	Path    string
	Version string
	Hash    string
}

type DependencyNode struct {
	Path        string
	Version     string
	Requires    []DependencyNode
	IsDirect    bool
	IsIndirect  bool
}

type MVSResult struct {
	Selected      map[string]string
	Conflicts     []Conflict
	Retracted     []RetractedVersion
	AllVersions   map[string][]string
}

type Conflict struct {
	Path          string
	RequiredBy    map[string]string
	Selected      string
}

type RetractedVersion struct {
	Path    string
	Version string
	Reason  string
}

type VerifyResult struct {
	AllChecked bool
	Orphans    []GoSumEntry
	Mismatches []Mismatch
}

type Mismatch struct {
	Path    string
	Version string
	Error   string
}

type DependencyGraph struct {
	Root   string
	Nodes  []GraphNode
	Edges  []GraphEdge
}

type GraphNode struct {
	ID       string
	Path     string
	Version  string
	IsDirect bool
}

type GraphEdge struct {
	From string
	To   string
}

type Version struct {
	Raw         string
	Major       int
	Minor       int
	Patch       int
	IsPseudo    bool
	PseudoTime  string
	PseudoHash  string
	PreRelease  string
	Build       string
}

type WhyResult struct {
	Found   bool
	Target  string
	Chains  [][]string
}

func (v Version) String() string {
	return v.Raw
}

func (v Version) Less(other Version) bool {
	if v.Major != other.Major {
		return v.Major < other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor < other.Minor
	}
	if v.Patch != other.Patch {
		return v.Patch < other.Patch
	}
	if v.IsPseudo && other.IsPseudo {
		return v.PseudoTime < other.PseudoTime
	}
	if v.PreRelease != "" && other.PreRelease == "" {
		return true
	}
	if v.PreRelease == "" && other.PreRelease != "" {
		return false
	}
	if v.PreRelease != other.PreRelease {
		return v.PreRelease < other.PreRelease
	}
	return v.Build < other.Build
}

func (v Version) Equal(other Version) bool {
	return v.Raw == other.Raw
}

func (v Version) Greater(other Version) bool {
	return !v.Less(other) && !v.Equal(other)
}

func (v Version) GreaterOrEqual(other Version) bool {
	return !v.Less(other)
}

func (r Retract) Contains(version string) bool {
	v, err := ParseVersion(version)
	if err != nil {
		return false
	}
	
	low, err := ParseVersion(r.LowVersion)
	if err != nil {
		return false
	}
	
	if r.HighVersion == "" {
		return v.Equal(low)
	}
	
	high, err := ParseVersion(r.HighVersion)
	if err != nil {
		return false
	}
	
	return v.GreaterOrEqual(low) && v.Less(high)
}

func NewGoModFile() *GoModFile {
	return &GoModFile{
		Requires: make([]Require, 0),
		Replaces: make([]Replace, 0),
		Excludes: make([]Exclude, 0),
		Retracts: make([]Retract, 0),
	}
}

func (g *GoModFile) String() string {
	return fmt.Sprintf("module %s (go %s)", g.Module, g.GoVersion)
}
