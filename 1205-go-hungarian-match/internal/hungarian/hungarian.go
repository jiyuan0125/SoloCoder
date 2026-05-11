package hungarian

type Edge struct {
	Intern   string
	Position string
}

type MatchResult struct {
	MaxMatch    int
	Assignments []Assignment
}

type Assignment struct {
	Intern   string
	Position string
}

type BipartiteGraph struct {
	interns    []string
	positions  []string
	edges      []Edge
	internIdx  map[string]int
	posIdx     map[string]int
	adjList    [][]int
}

func NewBipartiteGraph(interns, positions []string, edges []Edge) *BipartiteGraph {
	g := &BipartiteGraph{
		interns:   interns,
		positions: positions,
		edges:     edges,
		internIdx: make(map[string]int),
		posIdx:    make(map[string]int),
	}

	for i, id := range interns {
		g.internIdx[id] = i
	}
	for i, id := range positions {
		g.posIdx[id] = i
	}

	g.adjList = make([][]int, len(interns))
	for _, e := range edges {
		iIdx, ok := g.internIdx[e.Intern]
		if !ok {
			continue
		}
		pIdx, ok := g.posIdx[e.Position]
		if !ok {
			continue
		}
		g.adjList[iIdx] = append(g.adjList[iIdx], pIdx)
	}

	return g
}

func (g *BipartiteGraph) MaxMatch() MatchResult {
	if len(g.interns) == 0 || len(g.positions) == 0 {
		return MatchResult{MaxMatch: 0, Assignments: []Assignment{}}
	}

	matchTo := make([]int, len(g.positions))
	for i := range matchTo {
		matchTo[i] = -1
	}

	result := 0
	for i := 0; i < len(g.interns); i++ {
		visited := make([]bool, len(g.positions))
		if g.dfs(i, visited, matchTo) {
			result++
		}
	}

	assignments := make([]Assignment, 0, result)
	for pIdx, iIdx := range matchTo {
		if iIdx != -1 {
			assignments = append(assignments, Assignment{
				Intern:   g.interns[iIdx],
				Position: g.positions[pIdx],
			})
		}
	}

	return MatchResult{
		MaxMatch:    result,
		Assignments: assignments,
	}
}

func (g *BipartiteGraph) dfs(internIdx int, visited []bool, matchTo []int) bool {
	for _, posIdx := range g.adjList[internIdx] {
		if !visited[posIdx] {
			visited[posIdx] = true
			if matchTo[posIdx] == -1 || g.dfs(matchTo[posIdx], visited, matchTo) {
				matchTo[posIdx] = internIdx
				return true
			}
		}
	}
	return false
}

func (m MatchResult) GetInternAssignment(internID string) (position string, found bool) {
	for _, a := range m.Assignments {
		if a.Intern == internID {
			return a.Position, true
		}
	}
	return "", false
}
