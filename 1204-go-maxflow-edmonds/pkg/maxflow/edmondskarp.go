package maxflow

import (
	"fmt"
	"math"
)

type Result struct {
	MaxFlow int
	Flows   []EdgeFlow
}

type EdgeFlow struct {
	From  string
	To    string
	Flow  int
	Max   int
}

func (g *Graph) ComputeMaxFlow(source, sink string) (*Result, error) {
	if source == sink {
		flows := make([]EdgeFlow, len(g.origEdges))
		for i, e := range g.origEdges {
			flows[i] = EdgeFlow{
				From: g.nodes[e.from],
				To:   g.nodes[e.to],
				Flow: 0,
				Max:  e.capacity,
			}
		}
		return &Result{MaxFlow: 0, Flows: flows}, nil
	}

	srcIdx, ok := g.nodeMap[source]
	if !ok {
		return nil, fmt.Errorf("source node %s not found", source)
	}

	sinkIdx, ok := g.nodeMap[sink]
	if !ok {
		return nil, fmt.Errorf("sink node %s not found", sink)
	}

	edgeFlowMap := make([]int, len(g.origEdges))

	for {
		parent := make([]struct{ u, edgeIdx int }, len(g.nodes))
		for i := range parent {
			parent[i] = struct{ u, edgeIdx int }{u: -1, edgeIdx: -1}
		}

		minResidual := g.bfs(srcIdx, sinkIdx, parent)
		if minResidual == 0 {
			break
		}

		g.augmentFlow(sinkIdx, minResidual, parent, edgeFlowMap)
	}

	result := &Result{
		MaxFlow: 0,
		Flows:   make([]EdgeFlow, len(g.origEdges)),
	}

	for i, e := range g.origEdges {
		flow := edgeFlowMap[i]
		result.Flows[i] = EdgeFlow{
			From: g.nodes[e.from],
			To:   g.nodes[e.to],
			Flow: flow,
			Max:  e.capacity,
		}
		if e.from == srcIdx && flow > 0 {
			result.MaxFlow += flow
		}
	}

	if err := validateResult(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (g *Graph) bfs(src, sink int, parent []struct{ u, edgeIdx int }) int {
	visited := make([]bool, len(g.nodes))
	queue := make([]int, 0)
	queue = append(queue, src)
	visited[src] = true

	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]

		for edgeIdx, edge := range g.adj[u] {
			if !visited[edge.to] && edge.cap > 0 {
				visited[edge.to] = true
				parent[edge.to] = struct{ u, edgeIdx int }{u: u, edgeIdx: edgeIdx}

				if edge.to == sink {
					minCap := math.MaxInt32
					v := sink
					for v != src {
						p := parent[v]
						minCap = min(minCap, g.adj[p.u][p.edgeIdx].cap)
						v = p.u
					}
					return minCap
				}

				queue = append(queue, edge.to)
			}
		}
	}

	return 0
}

func (g *Graph) augmentFlow(sink, flow int, parent []struct{ u, edgeIdx int }, edgeFlowMap []int) {
	v := sink
	src := -1
	for v != src {
		p := parent[v]
		if p.u == -1 {
			src = v
			break
		}

		edge := &g.adj[p.u][p.edgeIdx]
		revEdge := &g.adj[v][edge.rev]

		edge.cap -= flow
		revEdge.cap += flow

		if edge.origIdx != -1 {
			edgeFlowMap[edge.origIdx] += flow
		}
		if revEdge.origIdx != -1 {
			edgeFlowMap[revEdge.origIdx] -= flow
		}

		v = p.u
	}
}

func validateResult(result *Result) error {
	for _, f := range result.Flows {
		if f.Flow < 0 {
			return fmt.Errorf("flow on edge %s->%s is negative: %d", f.From, f.To, f.Flow)
		}
		if f.Flow > f.Max {
			return fmt.Errorf("flow on edge %s->%s exceeds capacity: flow=%d, capacity=%d", f.From, f.To, f.Flow, f.Max)
		}
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
