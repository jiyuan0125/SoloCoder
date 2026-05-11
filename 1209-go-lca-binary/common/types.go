package common

import (
	"encoding/json"
)

type Node struct {
	ID       string  `json:"id"`
	ParentID *string `json:"parent_id"`
}

type Query struct {
	NodeA string `json:"node_a"`
	NodeB string `json:"node_b"`
}

type Request struct {
	Tree    TreeData `json:"tree"`
	Queries []Query  `json:"queries"`
}

type TreeData struct {
	Nodes []Node `json:"nodes"`
}

type Response struct {
	Results []ResultItem `json:"results"`
	Error   string       `json:"error,omitempty"`
}

type ResultItem struct {
	Query Query  `json:"query"`
	LCA   string `json:"lca"`
	Error string `json:"error,omitempty"`
}

func ParseRequest(data []byte) (*Request, error) {
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *Response) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
