package model

type File struct {
	Name    string
	Content string
}

type Request struct {
	PackagePath string
	Files       []File
}

type Response struct {
	Success bool
	Message string
	Format  string
	Output  string
}

type EdgeType string

const (
	EdgeTypeCall    EdgeType = "call"
	EdgeTypeStarts  EdgeType = "starts"
	EdgeTypeDefer   EdgeType = "defer"
)

type Node struct {
	Name       string
	Package    string
	File       string
	IsMain     bool
	IsInit     bool
	IsExported bool
	IsAnon     bool
}

type Edge struct {
	From      string
	To        string
	Type      EdgeType
	Interface bool
}

type Graph struct {
	Nodes []Node
	Edges []Edge
}
