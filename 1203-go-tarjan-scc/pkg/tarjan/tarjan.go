package tarjan

type Graph struct {
	nodes    map[string]bool
	adj      map[string][]string
	selfLoop map[string]bool
}

func NewGraph() *Graph {
	return &Graph{
		nodes:    make(map[string]bool),
		adj:      make(map[string][]string),
		selfLoop: make(map[string]bool),
	}
}

func (g *Graph) AddNode(node string) {
	g.nodes[node] = true
	if _, exists := g.adj[node]; !exists {
		g.adj[node] = []string{}
	}
}

func (g *Graph) AddEdge(from, to string) {
	if !g.nodes[from] {
		g.AddNode(from)
	}
	if !g.nodes[to] {
		g.AddNode(to)
	}
	g.adj[from] = append(g.adj[from], to)
	if from == to {
		g.selfLoop[from] = true
	}
}

type SCCResult struct {
	Nodes   []string
	IsCycle bool
}

func (g *Graph) TarjanSCC() ([]SCCResult, error) {
	discovery := make(map[string]int)
	low := make(map[string]int)
	inStack := make(map[string]bool)
	stack := []string{}
	time := 0
	var result []SCCResult

	var dfs func(u string)
	dfs = func(u string) {
		discovery[u] = time
		low[u] = time
		time++
		stack = append(stack, u)
		inStack[u] = true

		for _, v := range g.adj[u] {
			if _, visited := discovery[v]; !visited {
				dfs(v)
				if low[v] < low[u] {
					low[u] = low[v]
				}
			} else if inStack[v] {
				if discovery[v] < low[u] {
					low[u] = discovery[v]
				}
			}
		}

		if low[u] == discovery[u] {
			component := []string{}
			for {
				node := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				inStack[node] = false
				component = append(component, node)
				if node == u {
					break
				}
			}

			isCycle := len(component) > 1
			if len(component) == 1 {
				if g.selfLoop[component[0]] {
					isCycle = true
				}
			}

			result = append(result, SCCResult{
				Nodes:   component,
				IsCycle: isCycle,
			})
		}
	}

	for node := range g.nodes {
		if _, visited := discovery[node]; !visited {
			dfs(node)
		}
	}

	if len(stack) != 0 {
		return nil, &StackNotEmptyError{Remaining: stack}
	}

	return result, nil
}

type StackNotEmptyError struct {
	Remaining []string
}

func (e *StackNotEmptyError) Error() string {
	return "algorithm bug: stack not empty after completion"
}
