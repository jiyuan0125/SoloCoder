package annealing

import (
	"math"
	"math/rand"
	"time"
)

type SolverConfig struct {
	InitialTemperature       float64
	FinalTemperature         float64
	CoolingRate              float64
	IterationsPerTemperature int
	AdaptiveCooling          bool
	AdaptiveThreshold        int
	RandomSeed               int64
}

type ConvergencePoint struct {
	Temperature   float64
	ObjectiveValue float64
}

type SolverResult struct {
	Path              []int
	TotalDistance     float64
	Convergence       []ConvergencePoint
	Duration          time.Duration
	TotalIterations   int
	InitialDistance   float64
	ImprovementPct    float64
	RandomSeed        int64
}

type SolverProgress struct {
	CurrentTemperature     float64
	CurrentIteration       int
	TotalIterations        int
	CurrentDistance        float64
	BestDistance           float64
	InitialDistance        float64
	ImprovementPct         float64
	IsComplete             bool
}

type Solver struct {
	cities   []City
	config   SolverConfig
	rng      *rand.Rand
	progress SolverProgress
	result   *SolverResult
	stop     chan struct{}
	stopped  bool
}

func NewSolver(cities []City, config SolverConfig) (*Solver, error) {
	if len(cities) == 0 {
		return nil, ErrNoCities
	}

	seed := config.RandomSeed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	if config.InitialTemperature <= 0 {
		config.InitialTemperature = 1000.0
	}
	if config.FinalTemperature <= 0 {
		config.FinalTemperature = 0.01
	}
	if config.CoolingRate <= 0 || config.CoolingRate >= 1 {
		config.CoolingRate = 0.99
	}
	if config.IterationsPerTemperature <= 0 {
		config.IterationsPerTemperature = 100
	}
	if config.AdaptiveThreshold <= 0 {
		config.AdaptiveThreshold = 5
	}

	return &Solver{
		cities: cities,
		config: config,
		rng:    rand.New(rand.NewSource(seed)),
		stop:   make(chan struct{}),
	}, nil
}

func (s *Solver) Solve() (*SolverResult, error) {
	if len(s.cities) <= 1 {
		return s.handleTrivialCase(0)
	}
	if len(s.cities) == 2 {
		return s.handleTrivialCase(1)
	}

	start := time.Now()
	order, initialDistance := s.greedyInitialSolution()
	currentOrder := make([]int, len(order))
	copy(currentOrder, order)

	currentDistance := initialDistance
	bestOrder := make([]int, len(order))
	copy(bestOrder, order)
	bestDistance := initialDistance

	var convergence []ConvergencePoint
	totalIterations := 0
	temperature := s.config.InitialTemperature
	noImproveCount := 0
	lastBestDistance := initialDistance

	for temperature > s.config.FinalTemperature && !s.stopped {
		improvedThisTemp := false
		for i := 0; i < s.config.IterationsPerTemperature; i++ {
			if s.stopped {
				break
			}
			totalIterations++

			newOrder := s.twoOpt(currentOrder)
			newDistance := CalculateTotalDistance(s.cities, newOrder)

			if IsSolutionBetter(newDistance, currentDistance) {
				currentOrder = newOrder
				currentDistance = newDistance
				if IsSolutionBetter(newDistance, bestDistance) {
					bestOrder = make([]int, len(newOrder))
					copy(bestOrder, newOrder)
					bestDistance = newDistance
					improvedThisTemp = true
				}
			} else {
				delta := newDistance - currentDistance
				probability := math.Exp(-delta / temperature)
				if s.rng.Float64() < probability {
					currentOrder = newOrder
					currentDistance = newDistance
				}
			}

			s.progress = SolverProgress{
				CurrentTemperature: temperature,
				CurrentIteration:   totalIterations,
				CurrentDistance:    currentDistance,
				BestDistance:       bestDistance,
				InitialDistance:    initialDistance,
				ImprovementPct:     calculateImprovementPct(initialDistance, bestDistance),
				IsComplete:         false,
			}
		}

		convergence = append(convergence, ConvergencePoint{
			Temperature:    temperature,
			ObjectiveValue: bestDistance,
		})

		if improvedThisTemp || IsSolutionBetter(lastBestDistance, bestDistance) {
			noImproveCount = 0
			lastBestDistance = bestDistance
		} else {
			noImproveCount++
		}

		if s.config.AdaptiveCooling && noImproveCount >= s.config.AdaptiveThreshold {
			temperature *= s.config.CoolingRate * s.config.CoolingRate
		} else {
			temperature *= s.config.CoolingRate
		}
	}

	duration := time.Since(start)
	improvementPct := calculateImprovementPct(initialDistance, bestDistance)

	result := &SolverResult{
		Path:            bestOrder,
		TotalDistance:   bestDistance,
		Convergence:     convergence,
		Duration:        duration,
		TotalIterations: totalIterations,
		InitialDistance: initialDistance,
		ImprovementPct:  improvementPct,
		RandomSeed:      s.config.RandomSeed,
	}
	s.result = result
	s.progress.IsComplete = true

	return result, nil
}

func (s *Solver) GetProgress() SolverProgress {
	return s.progress
}

func (s *Solver) GetResult() *SolverResult {
	return s.result
}

func (s *Solver) Stop() {
	if !s.stopped {
		s.stopped = true
		close(s.stop)
	}
}

func (s *Solver) handleTrivialCase(cityCount int) (*SolverResult, error) {
	start := time.Now()
	order := make([]int, cityCount+1)
	for i := range order {
		order[i] = i
		if i == cityCount && cityCount > 0 {
			order[i] = 0
		}
	}
	if cityCount == 0 {
		order = []int{0}
	}
	totalDistance := CalculateTotalDistance(s.cities, order[:len(s.cities)])
	if cityCount == 0 {
		order = []int{}
	}

	return &SolverResult{
		Path:            order[:len(s.cities)],
		TotalDistance:   totalDistance,
		Convergence:     []ConvergencePoint{},
		Duration:        time.Since(start),
		TotalIterations: 0,
		InitialDistance: totalDistance,
		ImprovementPct:  0,
		RandomSeed:      s.config.RandomSeed,
	}, nil
}

func (s *Solver) greedyInitialSolution() ([]int, float64) {
	n := len(s.cities)
	visited := make([]bool, n)
	order := make([]int, 0, n)
	current := 0
	visited[current] = true
	order = append(order, current)

	for len(order) < n {
		nearest := -1
		nearestDist := math.Inf(1)
		for i := 0; i < n; i++ {
			if !visited[i] && i != current {
				dist := s.cities[current].Distance(s.cities[i])
				if dist < nearestDist {
					nearestDist = dist
					nearest = i
				}
			}
		}
		if nearest == -1 {
			for i := 0; i < n; i++ {
				if !visited[i] {
					nearest = i
					break
				}
			}
		}
		visited[nearest] = true
		order = append(order, nearest)
		current = nearest
	}

	totalDist := CalculateTotalDistance(s.cities, order)
	return order, totalDist
}

func (s *Solver) twoOpt(order []int) []int {
	n := len(order)
	newOrder := make([]int, n)
	copy(newOrder, order)

	i := s.rng.Intn(n)
	j := s.rng.Intn(n)
	if i > j {
		i, j = j, i
	}

	for k, l := i+1, j; k < l; k, l = k+1, l-1 {
		newOrder[k], newOrder[l] = newOrder[l], newOrder[k]
	}

	return newOrder
}

func calculateImprovementPct(initial, current float64) float64 {
	if initial == 0 {
		return 0
	}
	return ((initial - current) / initial) * 100
}

var ErrNoCities = &SolverError{msg: "no cities provided"}

type SolverError struct {
	msg string
}

func (e *SolverError) Error() string {
	return e.msg
}
