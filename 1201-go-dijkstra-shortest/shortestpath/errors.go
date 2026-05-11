package shortestpath

import "errors"

var (
	ErrNegativeWeight   = errors.New("edge weight must be positive")
	ErrNodeNotFound    = errors.New("node not found in graph")
	ErrNoPath         = errors.New("no path exists between nodes")
	ErrEmptyGraph     = errors.New("graph is empty")
	ErrStartEqualsEnd  = errors.New("start and end are the same node")
)
