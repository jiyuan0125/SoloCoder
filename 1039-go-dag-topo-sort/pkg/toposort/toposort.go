package toposort

type SortResult struct {
	Order           []string
	MaxParallelism  int
	CriticalPath    []string
	MinCompletionTime int
}

func (g *Graph) Sort() (*SortResult, error) {
	if g.TaskCount() == 0 {
		return &SortResult{
			Order:             []string{},
			MaxParallelism:    0,
			CriticalPath:      []string{},
			MinCompletionTime: 0,
		}, nil
	}

	for taskID, task := range g.tasks {
		for _, dep := range task.Dependencies {
			if dep == taskID {
				return nil, ErrSelfDependency{TaskID: taskID}
			}
			if _, exists := g.tasks[dep]; !exists {
				return nil, ErrDependencyNotFound{TaskID: taskID, DependencyID: dep}
			}
		}
	}

	inDegree := make(map[string]int)
	for taskID := range g.tasks {
		inDegree[taskID] = 0
	}
	for _, task := range g.tasks {
		for _, dep := range task.Dependencies {
			inDegree[dep]++
		}
	}

	levels := g.calculateLevels()

	maxParallelism := 0
	levelCounts := make(map[int]int)
	for _, level := range levels {
		levelCounts[level]++
		if levelCounts[level] > maxParallelism {
			maxParallelism = levelCounts[level]
		}
	}

	order, err := g.kahnSort()
	if err != nil {
		return nil, err
	}

	criticalPath := g.calculateCriticalPath()

	return &SortResult{
		Order:             order,
		MaxParallelism:    maxParallelism,
		CriticalPath:      criticalPath,
		MinCompletionTime: len(criticalPath),
	}, nil
}

func (g *Graph) kahnSort() ([]string, error) {
	inDegree := make(map[string]int)
	for taskID := range g.tasks {
		inDegree[taskID] = 0
	}

	adjacency := make(map[string][]string)
	for taskID, task := range g.tasks {
		for _, dep := range task.Dependencies {
			adjacency[dep] = append(adjacency[dep], taskID)
			inDegree[taskID]++
		}
	}

	queue := make([]string, 0)
	for taskID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, taskID)
		}
	}

	result := make([]string, 0, g.TaskCount())
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		for _, neighbor := range adjacency[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(result) != g.TaskCount() {
		cycle := g.detectCycle()
		return nil, ErrCyclicDependency{Path: cycle}
	}

	return result, nil
}

func (g *Graph) detectCycle() []string {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var cycle []string

	var dfs func(taskID string, path []string) bool
	dfs = func(taskID string, path []string) bool {
		visited[taskID] = true
		recStack[taskID] = true
		newPath := append(path, taskID)

		task, exists := g.tasks[taskID]
		if !exists {
			return false
		}

		for _, dep := range task.Dependencies {
			if !visited[dep] {
				if dfs(dep, newPath) {
					return true
				}
			} else if recStack[dep] {
				startIdx := -1
				for i, id := range newPath {
					if id == dep {
						startIdx = i
						break
					}
				}
				if startIdx != -1 {
					cycle = append(newPath[startIdx:], dep)
				}
				return true
			}
		}

		recStack[taskID] = false
		return false
	}

	for taskID := range g.tasks {
		if !visited[taskID] {
			if dfs(taskID, []string{}) {
				return cycle
			}
		}
	}

	return []string{}
}

func (g *Graph) calculateLevels() map[string]int {
	levels := make(map[string]int)
	inDegree := make(map[string]int)
	for taskID := range g.tasks {
		inDegree[taskID] = 0
	}

	adjacency := make(map[string][]string)
	for taskID, task := range g.tasks {
		for _, dep := range task.Dependencies {
			adjacency[dep] = append(adjacency[dep], taskID)
			inDegree[taskID]++
		}
	}

	queue := make([]string, 0)
	for taskID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, taskID)
			levels[taskID] = 0
		}
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, neighbor := range adjacency[current] {
			inDegree[neighbor]--
			if levels[neighbor] < levels[current]+1 {
				levels[neighbor] = levels[current] + 1
			}
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	return levels
}

func (g *Graph) calculateCriticalPath() []string {
	if g.TaskCount() == 0 {
		return []string{}
	}

	dist := make(map[string]int)
	prev := make(map[string]string)
	for taskID := range g.tasks {
		dist[taskID] = 1
		prev[taskID] = ""
	}

	order, err := g.kahnSort()
	if err != nil {
		return []string{}
	}

	for _, taskID := range order {
		task := g.tasks[taskID]
		for _, dep := range task.Dependencies {
			if dist[taskID] < dist[dep]+1 {
				dist[taskID] = dist[dep] + 1
				prev[taskID] = dep
			}
		}
	}

	maxDist := 0
	endNode := ""
	for taskID, d := range dist {
		if d > maxDist {
			maxDist = d
			endNode = taskID
		}
	}

	if endNode == "" {
		return []string{}
	}

	path := []string{}
	current := endNode
	for current != "" {
		path = append([]string{current}, path...)
		current = prev[current]
	}

	return path
}
