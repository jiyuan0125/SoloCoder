package anchor

import "fmt"

type AnchorInfo struct {
	Name     string
	Location string
}

type ReferenceInfo struct {
	Name       string
	Location   string
	IsMergeKey bool
}

type ReferenceRelation struct {
	Anchor    string
	Reference string
	IsMerge   bool
}

type AnalysisResult struct {
	Anchors     []AnchorInfo
	References  []ReferenceInfo
	Relations   []ReferenceRelation
	Warnings    []string
	ResolvedData interface{}
}

type CircularReferenceError struct {
	Chain []string
}

func (e *CircularReferenceError) Error() string {
	return fmt.Sprintf("circular reference detected: %v", e.Chain)
}
