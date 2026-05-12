package dag

import (
	"errors"
	"sort"
)

type Node struct {
	ID    string
	Value interface{}
}

type Edge struct {
	From string
	To   string
}

type Graph struct {
	nodes map[string]*Node
	edges map[string][]Edge
	inDegree map[string]int
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
		edges: make(map[string][]Edge),
		inDegree: make(map[string]int),
	}
}

func (g *Graph) AddNode(id string, value interface{}) error {
	if _, exists := g.nodes[id]; exists {
		return errors.New("node already exists")
	}
	g.nodes[id] = &Node{ID: id, Value: value}
	g.inDegree[id] = 0
	return nil
}

func (g *Graph) GetNode(id string) (*Node, bool) {
	node, exists := g.nodes[id]
	return node, exists
}

func (g *Graph) AddEdge(from, to string) error {
	if _, exists := g.nodes[from]; !exists {
		return errors.New("from node does not exist")
	}
	if _, exists := g.nodes[to]; !exists {
		return errors.New("to node does not exist")
	}

	existingEdges := g.edges[from]
	for _, e := range existingEdges {
		if e.From == from && e.To == to {
			return errors.New("edge already exists")
		}
	}

	testGraph := g.clone()
	testGraph.edges[from] = append(testGraph.edges[from], Edge{From: from, To: to})
	testGraph.inDegree[to]++

	if hasCycle(testGraph) {
		return errors.New("adding this edge would create a cycle")
	}

	g.edges[from] = append(g.edges[from], Edge{From: from, To: to})
	g.inDegree[to]++
	return nil
}

func (g *Graph) HasNode(id string) bool {
	_, exists := g.nodes[id]
	return exists
}

func (g *Graph) HasCycle() bool {
	return hasCycle(g)
}

func hasCycle(g *Graph) bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for id := range g.nodes {
		if !visited[id] {
			if dfsCycleCheck(g, id, visited, recStack) {
				return true
			}
		}
	}
	return false
}

func dfsCycleCheck(g *Graph, nodeID string, visited, recStack map[string]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	for _, edge := range g.edges[nodeID] {
		if !visited[edge.To] {
			if dfsCycleCheck(g, edge.To, visited, recStack) {
				return true
			}
		} else if recStack[edge.To] {
			return true
		}
	}

	recStack[nodeID] = false
	return false
}

func (g *Graph) TopologicalSort() ([]string, error) {
	if g.HasCycle() {
		return nil, errors.New("graph has a cycle")
	}

	inDegreeCopy := make(map[string]int)
	for k, v := range g.inDegree {
		inDegreeCopy[k] = v
	}

	var queue []string
	for id, degree := range inDegreeCopy {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	sort.Strings(queue)

	var result []string
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		result = append(result, nodeID)

		for _, edge := range g.edges[nodeID] {
			inDegreeCopy[edge.To]--
			if inDegreeCopy[edge.To] == 0 {
				queue = append(queue, edge.To)
				sort.Strings(queue)
			}
		}
	}

	if len(result) != len(g.nodes) {
		return nil, errors.New("graph has a cycle")
	}

	return result, nil
}

func (g *Graph) ReverseTopologicalSort() ([]string, error) {
	topo, err := g.TopologicalSort()
	if err != nil {
		return nil, err
	}

	reversed := make([]string, len(topo))
	for i, j := 0, len(topo)-1; i < len(topo); i, j = i+1, j-1 {
		reversed[i] = topo[j]
	}

	return reversed, nil
}

func (g *Graph) GetLevels() (map[string]int, error) {
	topo, err := g.TopologicalSort()
	if err != nil {
		return nil, err
	}

	levels := make(map[string]int)
	for _, id := range topo {
		levels[id] = 0
	}

	for _, id := range topo {
		for _, edge := range g.edges[id] {
			if levels[edge.To] <= levels[id] {
				levels[edge.To] = levels[id] + 1
			}
		}
	}

	return levels, nil
}

func (g *Graph) GetNodesByLevel() (map[int][]string, error) {
	levels, err := g.GetLevels()
	if err != nil {
		return nil, err
	}

	nodesByLevel := make(map[int][]string)
	for id, level := range levels {
		nodesByLevel[level] = append(nodesByLevel[level], id)
	}

	for level := range nodesByLevel {
		sort.Strings(nodesByLevel[level])
	}

	return nodesByLevel, nil
}

func (g *Graph) GetPrerequisites(courseID string) []string {
	var prereqs []string
	for fromID, edges := range g.edges {
		for _, edge := range edges {
			if edge.To == courseID {
				prereqs = append(prereqs, fromID)
			}
		}
	}
	sort.Strings(prereqs)
	return prereqs
}

func (g *Graph) GetSuccessors(courseID string) []string {
	var successors []string
	for _, edge := range g.edges[courseID] {
		successors = append(successors, edge.To)
	}
	sort.Strings(successors)
	return successors
}

func (g *Graph) CanStartCourse(courseID string, completedCourses map[string]bool) bool {
	prereqs := g.GetPrerequisites(courseID)
	for _, prereqID := range prereqs {
		if !completedCourses[prereqID] {
			return false
		}
	}
	return true
}

func (g *Graph) clone() *Graph {
	newGraph := NewGraph()

	for id, node := range g.nodes {
		newGraph.nodes[id] = node
		newGraph.inDegree[id] = g.inDegree[id]
	}

	for from, edges := range g.edges {
		newGraph.edges[from] = append(newGraph.edges[from], edges...)
	}

	return newGraph
}

func (g *Graph) GetNodes() map[string]*Node {
	return g.nodes
}

func (g *Graph) GetEdges() map[string][]Edge {
	return g.edges
}
