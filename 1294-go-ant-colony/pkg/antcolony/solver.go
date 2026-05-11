package antcolony

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
)

type Solver struct {
	graph  *internalGraph
	params Parameters

	startIdx int
	endIdx   int

	pheromones [][]float64
	heuristics [][]float64

	bestPath    []int
	bestLength  float64

	currentIteration int
	status           string
	progress         float64
	result           *SolveResult
	history          []float64

	mu sync.RWMutex
}

func NewSolver(graph Graph, start, end string, params Parameters) (*Solver, error) {
	g, err := buildInternalGraph(graph)
	if err != nil {
		return nil, err
	}

	startIdx, ok := g.getNodeIndex(start)
	if !ok {
		return nil, fmt.Errorf("start node not found: %s", start)
	}

	endIdx, ok := g.getNodeIndex(end)
	if !ok {
		return nil, fmt.Errorf("end node not found: %s", end)
	}

	if startIdx == endIdx {
		s := &Solver{
			graph:      g,
			params:     params,
			startIdx:   startIdx,
			endIdx:     endIdx,
			status:     StatusCompleted,
			progress:   1.0,
			bestLength: 0,
			result: &SolveResult{
				Success:    true,
				Path:       Path{Nodes: []string{start}, Length: 0},
				Iterations: 0,
				Converged:  true,
			},
		}
		return s, nil
	}

	if err := validateParameters(params); err != nil {
		return nil, err
	}

	if !isConnected(g, startIdx, endIdx) {
		return nil, fmt.Errorf("no path exists from %s to %s (graph not connected)", start, end)
	}

	s := &Solver{
		graph:      g,
		params:     params,
		startIdx:   startIdx,
		endIdx:     endIdx,
		bestLength: math.Inf(1),
		status:     StatusIdle,
		history:    make([]float64, 0),
	}

	nodeCount := len(g.nodes)
	s.pheromones = make([][]float64, nodeCount)
	s.heuristics = make([][]float64, nodeCount)
	for i := range s.pheromones {
		s.pheromones[i] = make([]float64, nodeCount)
		s.heuristics[i] = make([]float64, nodeCount)
	}

	for i := 0; i < nodeCount; i++ {
		for j := 0; j < nodeCount; j++ {
			if weight, ok := g.getEdgeWeight(i, j); ok {
				s.pheromones[i][j] = params.InitialPheromone
				s.heuristics[i][j] = 1.0 / weight
			}
		}
	}

	return s, nil
}

func (s *Solver) GetStatus() SolveStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := SolveStatus{
		Status:           s.status,
		Progress:         s.progress,
		CurrentIteration: s.currentIteration,
		BestLength:       s.bestLength,
		Result:           s.result,
	}
	return status
}

func (s *Solver) IsLargeGraph() bool {
	return checkLargeGraph(s.graph)
}

func (s *Solver) Solve() SolveResult {
	s.mu.Lock()
	if s.status == StatusRunning {
		s.mu.Unlock()
		return SolveResult{Success: false, Error: "solver is already running"}
	}
	if s.status == StatusCompleted && s.result != nil {
		result := *s.result
		s.mu.Unlock()
		return result
	}
	s.status = StatusRunning
	s.progress = 0.0
	s.currentIteration = 0
	s.mu.Unlock()

	var finalResult SolveResult
	maxIters := s.params.MaxIterations
	consecutiveSame := 0
	prevBestLength := s.bestLength

	for iter := 0; iter < maxIters; iter++ {
		s.mu.Lock()
		s.currentIteration = iter + 1
		s.progress = float64(iter+1) / float64(maxIters)
		s.mu.Unlock()

		paths, lengths := s.constructAllPaths()

		s.updatePheromones(paths, lengths)

		currentBestLength, currentBestPath := s.findBestInIteration(paths, lengths)
		if currentBestLength < s.bestLength {
			s.mu.Lock()
			s.bestLength = currentBestLength
			s.bestPath = make([]int, len(currentBestPath))
			copy(s.bestPath, currentBestPath)
			s.mu.Unlock()
		}

		s.mu.Lock()
		s.history = append(s.history, s.bestLength)
		s.mu.Unlock()

		if math.Abs(prevBestLength-s.bestLength) < s.params.ConvergenceThreshold {
			consecutiveSame++
			if consecutiveSame >= 5 {
				break
			}
		} else {
			consecutiveSame = 0
		}
		prevBestLength = s.bestLength
	}

	finalPath := make([]string, len(s.bestPath))
	for i, idx := range s.bestPath {
		finalPath[i] = s.graph.nodes[idx]
	}

	finalResult = SolveResult{
		Success:          true,
		Path:             Path{Nodes: finalPath, Length: s.bestLength},
		Iterations:       s.currentIteration,
		Converged:        consecutiveSame >= 5,
		BestLengthHistory: s.history,
	}

	s.mu.Lock()
	s.status = StatusCompleted
	s.progress = 1.0
	s.result = &finalResult
	s.mu.Unlock()

	return finalResult
}

func (s *Solver) constructAllPaths() ([][]int, []float64) {
	antCount := s.params.AntCount
	paths := make([][]int, antCount)
	lengths := make([]float64, antCount)

	var wg sync.WaitGroup
	for i := 0; i < antCount; i++ {
		wg.Add(1)
		go func(antIdx int) {
			defer wg.Done()
			path, length := s.constructPath()
			paths[antIdx] = path
			lengths[antIdx] = length
		}(i)
	}
	wg.Wait()

	return paths, lengths
}

func (s *Solver) constructPath() ([]int, float64) {
	path := []int{s.startIdx}
	visited := make(map[int]bool)
	visited[s.startIdx] = true
	current := s.startIdx
	totalLength := 0.0

	for current != s.endIdx {
		neighbors := s.graph.getNeighbors(current)
		available := make([]int, 0, len(neighbors))
		for _, n := range neighbors {
			if !visited[n] {
				available = append(available, n)
			}
		}

		if len(available) == 0 {
			return nil, math.Inf(1)
		}

		nextNode := s.selectNextNode(current, available)
		weight, _ := s.graph.getEdgeWeight(current, nextNode)
		totalLength += weight
		path = append(path, nextNode)
		visited[nextNode] = true
		current = nextNode
	}

	return path, totalLength
}

func (s *Solver) selectNextNode(current int, available []int) int {
	probs := make([]float64, len(available))
	total := 0.0

	s.mu.RLock()
	for i, next := range available {
		tau := s.pheromones[current][next]
		eta := s.heuristics[current][next]

		tauPow := math.Pow(tau, s.params.Alpha)
		etaPow := math.Pow(eta, s.params.Beta)

		if math.IsInf(tauPow, 0) || math.IsInf(etaPow, 0) {
			probs[i] = math.Inf(1)
		} else {
			probs[i] = tauPow * etaPow
		}
		total += probs[i]
	}
	s.mu.RUnlock()

	if math.IsInf(total, 1) {
		maxProb := 0.0
		maxIdx := 0
		for i, p := range probs {
			if p > maxProb {
				maxProb = p
				maxIdx = i
			}
		}
		return available[maxIdx]
	}

	if total == 0 {
		return available[rand.Intn(len(available))]
	}

	for i := range probs {
		probs[i] /= total
	}

	r := rand.Float64()
	cumulative := 0.0
	for i, p := range probs {
		cumulative += p
		if r < cumulative {
			return available[i]
		}
	}

	return available[len(available)-1]
}

func (s *Solver) updatePheromones(paths [][]int, lengths []float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nodeCount := len(s.graph.nodes)
	for i := 0; i < nodeCount; i++ {
		for j := 0; j < nodeCount; j++ {
			s.pheromones[i][j] *= (1.0 - s.params.EvaporationRate)
		}
	}

	for antIdx, path := range paths {
		length := lengths[antIdx]
		if math.IsInf(length, 1) || len(path) == 0 {
			continue
		}
		delta := 1.0 / length
		for i := 0; i < len(path)-1; i++ {
			from := path[i]
			to := path[i+1]
			s.pheromones[from][to] += delta
		}
	}

	if s.bestPath != nil {
		eliteDelta := s.params.EliteWeight / s.bestLength
		for i := 0; i < len(s.bestPath)-1; i++ {
			from := s.bestPath[i]
			to := s.bestPath[i+1]
			s.pheromones[from][to] += eliteDelta
		}
	}
}

func (s *Solver) findBestInIteration(paths [][]int, lengths []float64) (float64, []int) {
	bestLen := math.Inf(1)
	var bestPath []int
	for i, length := range lengths {
		if length < bestLen {
			bestLen = length
			bestPath = paths[i]
		}
	}
	return bestLen, bestPath
}
