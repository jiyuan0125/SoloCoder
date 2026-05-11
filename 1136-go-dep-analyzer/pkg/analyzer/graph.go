package analyzer

func BuildDependencyGraph(mod *GoModFile, sumVersions map[string]string) *DependencyGraph {
	graph := &DependencyGraph{
		Nodes:    make(map[string]*DependencyNode),
		Edges:    []DependencyEdge{},
		Replaces: make(map[string]Replace),
	}

	for _, rep := range mod.Replaces {
		key := rep.OldPath
		if rep.OldVersion != "" {
			key = rep.OldPath + "@" + rep.OldVersion
		}
		graph.Replaces[key] = rep
	}

	rootNode := &DependencyNode{
		Path:       mod.Module,
		Version:    "",
		IsRoot:     true,
		IsIndirect: false,
	}
	graph.Nodes[mod.Module] = rootNode

	for _, req := range mod.Requires {
		targetPath := req.Path
		targetVersion := req.Version

		replaceKey := req.Path
		if rep, ok := graph.Replaces[replaceKey]; ok {
			targetPath = rep.NewPath
			if rep.NewVersion != "" {
				targetVersion = rep.NewVersion
			}
		}
		replaceKeyWithVersion := req.Path + "@" + req.Version
		if rep, ok := graph.Replaces[replaceKeyWithVersion]; ok {
			targetPath = rep.NewPath
			if rep.NewVersion != "" {
				targetVersion = rep.NewVersion
			}
		}

		if sumVer, ok := sumVersions[targetPath]; ok {
			targetVersion = sumVer
		}

		if _, exists := graph.Nodes[targetPath]; !exists {
			graph.Nodes[targetPath] = &DependencyNode{
				Path:       targetPath,
				Version:    targetVersion,
				IsRoot:     false,
				IsIndirect: req.IsIndirect,
			}
		}

		graph.Edges = append(graph.Edges, DependencyEdge{
			From:     mod.Module,
			To:       targetPath,
			Version:  targetVersion,
			IsDirect: !req.IsIndirect,
		})
	}

	for module, version := range sumVersions {
		if module == mod.Module {
			continue
		}
		if _, exists := graph.Nodes[module]; !exists {
			graph.Nodes[module] = &DependencyNode{
				Path:       module,
				Version:    version,
				IsRoot:     false,
				IsIndirect: true,
			}
		}
	}

	return graph
}

func DetectCycles(graph *DependencyGraph) [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(node string)
	dfs = func(node string) {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, edge := range graph.Edges {
			if edge.From == node {
				neighbor := edge.To
				if !visited[neighbor] {
					dfs(neighbor)
				} else if recStack[neighbor] {
					cycle := []string{}
					for i := len(path) - 1; i >= 0; i-- {
						cycle = append([]string{path[i]}, cycle...)
						if path[i] == neighbor {
							break
						}
					}
					cycle = append(cycle, neighbor)
					cycles = append(cycles, cycle)
				}
			}
		}

		recStack[node] = false
		path = path[:len(path)-1]
	}

	for node := range graph.Nodes {
		if !visited[node] {
			dfs(node)
		}
	}

	return cycles
}

func TopologicalSort(graph *DependencyGraph) []string {
	inDegree := make(map[string]int)
	for node := range graph.Nodes {
		inDegree[node] = 0
	}

	for _, edge := range graph.Edges {
		inDegree[edge.To]++
	}

	queue := []string{}
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	result := []string{}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		for _, edge := range graph.Edges {
			if edge.From == node {
				inDegree[edge.To]--
				if inDegree[edge.To] == 0 {
					queue = append(queue, edge.To)
				}
			}
		}
	}

	if len(result) != len(graph.Nodes) {
		result = []string{}
		for node := range graph.Nodes {
			result = append(result, node)
		}
	}

	return result
}

func CalculateDepths(graph *DependencyGraph) map[string]int {
	depths := make(map[string]int)
	for node := range graph.Nodes {
		depths[node] = 0
	}

	for i := 0; i < len(graph.Nodes); i++ {
		updated := false
		for _, edge := range graph.Edges {
			from := edge.From
			to := edge.To
			if depths[to] < depths[from]+1 {
				depths[to] = depths[from] + 1
				updated = true
			}
		}
		if !updated {
			break
		}
	}

	return depths
}

func FindDependencyChains(graph *DependencyGraph, targetModule string) [][]string {
	var allChains [][]string
	visited := make(map[string]bool)

	var dfs func(current string, path []string)
	dfs = func(current string, path []string) {
		path = append(path, current)
		visited[current] = true

		if current == targetModule && len(path) > 1 {
			chainCopy := make([]string, len(path))
			copy(chainCopy, path)
			allChains = append(allChains, chainCopy)
		} else {
			for _, edge := range graph.Edges {
				if edge.From == current && !visited[edge.To] {
					dfs(edge.To, path)
				}
			}
		}

		visited[current] = false
		path = path[:len(path)-1]
	}

	for node := range graph.Nodes {
		if graph.Nodes[node].IsRoot {
			clear(visited)
			dfs(node, []string{})
		}
	}

	return allChains
}
