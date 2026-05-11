package euler

import (
	"fmt"
)

type EulerResult struct {
	HasEulerPath    bool
	HasEulerCircuit bool
	Path            []Node
}

type EulerError struct {
	Msg string
}

func (e *EulerError) Error() string {
	return e.Msg
}

func NewEulerError(msg string) *EulerError {
	return &EulerError{Msg: msg}
}

func (g *Graph) HasEulerCircuit() bool {
	if !g.hasEdges() {
		return true
	}
	if !g.IsConnected() {
		return false
	}
	odds := g.OddDegreeNodes()
	return len(odds) == 0
}

func (g *Graph) HasEulerPath() bool {
	if !g.hasEdges() {
		return true
	}
	if !g.IsConnected() {
		return false
	}
	odds := g.OddDegreeNodes()
	return len(odds) == 0 || len(odds) == 2
}

func (g *Graph) validateStartNode(start Node) error {
	if !g.Nodes[start] {
		return NewEulerError(fmt.Sprintf("start node %s does not exist", start))
	}

	odds := g.OddDegreeNodes()
	degree := g.Degree(start)

	if !g.hasEdges() {
		return nil
	}

	if len(odds) == 0 {
		if degree%2 != 0 {
			return NewEulerError(fmt.Sprintf("for euler circuit, start node must have even degree, but node %s has degree %d", start, degree))
		}
	} else if len(odds) == 2 {
		if degree%2 == 0 {
			return NewEulerError(fmt.Sprintf("for euler path, start node must be one of the two odd-degree nodes (%v), but node %s has even degree %d", odds, start, degree))
		}
		if start != odds[0] && start != odds[1] {
			return NewEulerError(fmt.Sprintf("for euler path, start node must be one of the two odd-degree nodes: %v", odds))
		}
	}

	return nil
}

func (g *Graph) selectStartNode() (Node, error) {
	if !g.hasEdges() {
		for n := range g.Nodes {
			return n, nil
		}
		return "", NewEulerError("no nodes in graph")
	}

	odds := g.OddDegreeNodes()
	if len(odds) == 2 {
		return odds[0], nil
	}

	for n := range g.getNodesWithEdges() {
		return n, nil
	}

	return "", NewEulerError("no valid start node")
}

func (g *Graph) FindEulerPath(start Node) (*EulerResult, error) {
	if !g.hasEdges() {
		if len(g.Nodes) == 0 {
			return nil, NewEulerError("无街道数据")
		}
		return &EulerResult{
			HasEulerPath:    false,
			HasEulerCircuit: false,
			Path:            nil,
		}, NewEulerError("无街道数据")
	}

	if !g.IsConnected() {
		return &EulerResult{
			HasEulerPath:    false,
			HasEulerCircuit: false,
			Path:            nil,
		}, NewEulerError("图不连通，不存在欧拉路径或回路")
	}

	odds := g.OddDegreeNodes()
	hasCircuit := len(odds) == 0
	hasPath := len(odds) == 0 || len(odds) == 2

	if !hasPath {
		return &EulerResult{
			HasEulerPath:    false,
			HasEulerCircuit: false,
			Path:            nil,
		}, NewEulerError(fmt.Sprintf("图中存在 %d 个奇数度节点，不存在欧拉路径或回路", len(odds)))
	}

	var actualStart Node
	var err error

	if start == "" {
		actualStart, err = g.selectStartNode()
		if err != nil {
			return nil, err
		}
	} else {
		if err := g.validateStartNode(start); err != nil {
			return &EulerResult{
				HasEulerPath:    hasPath,
				HasEulerCircuit: hasCircuit,
				Path:            nil,
			}, err
		}
		actualStart = start
	}

	path, err := g.fleuryAlgorithm(actualStart)
	if err != nil {
		return &EulerResult{
			HasEulerPath:    hasPath,
			HasEulerCircuit: hasCircuit,
			Path:            nil,
		}, err
	}

	return &EulerResult{
		HasEulerPath:    hasPath,
		HasEulerCircuit: hasCircuit,
		Path:            path,
	}, nil
}

func (g *Graph) fleuryAlgorithm(start Node) ([]Node, error) {
	path := []Node{start}
	usedEdges := make(map[int]bool)

	current := start
	for len(usedEdges) < len(g.Edges) {
		availableEdges := make([]int, 0)
		for _, edgeIdx := range g.adj[current] {
			if !usedEdges[edgeIdx] {
				availableEdges = append(availableEdges, edgeIdx)
			}
		}

		if len(availableEdges) == 0 {
			return nil, NewEulerError("没有可用边，但尚未遍历所有边")
		}

		nextEdgeIdx := -1
		for _, edgeIdx := range availableEdges {
			if !g.isBridge(edgeIdx, usedEdges) {
				nextEdgeIdx = edgeIdx
				break
			}
		}

		if nextEdgeIdx == -1 {
			nextEdgeIdx = availableEdges[0]
		}

		usedEdges[nextEdgeIdx] = true

		nextNode, err := g.OtherEnd(nextEdgeIdx, current)
		if err != nil {
			return nil, err
		}

		path = append(path, nextNode)
		current = nextNode
	}

	return path, nil
}
