package task

import (
	"errors"
	"sync"
	"time"

	"genetic-algo/pkg/api"
	"genetic-algo/pkg/ga"
	"genetic-algo/pkg/testfuncs"
)

type Task struct {
	ID          string
	Status      api.TaskStatus
	Request     *api.OptimizeRequest
	Algorithm   *ga.Algorithm
	BestGenes   []float64
	BestFitness float64
	CurrentGen  int
	TotalGens   int
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
	Error       string
}

type Manager struct {
	tasks map[string]*Task
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		tasks: make(map[string]*Task),
	}
}

func (m *Manager) CreateTask(req *api.OptimizeRequest) (*Task, error) {
	taskID := generateTaskID()

	t := &Task{
		ID:        taskID,
		Status:    api.TaskStatusPending,
		Request:   req,
		CreatedAt: time.Now(),
		CurrentGen: 0,
	}

	m.mu.Lock()
	m.tasks[taskID] = t
	m.mu.Unlock()

	go m.runTask(t)

	return t, nil
}

func (m *Manager) GetTask(id string) (*Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	t, ok := m.tasks[id]
	if !ok {
		return nil, errors.New("task not found")
	}
	return t, nil
}

func (m *Manager) ListTasks() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

func (m *Manager) runTask(t *Task) {
	m.mu.Lock()
	t.Status = api.TaskStatusRunning
	t.StartedAt = time.Now()
	m.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			m.mu.Lock()
			t.Status = api.TaskStatusFailed
			t.Error = "panic occurred during execution"
			t.CompletedAt = time.Now()
			m.mu.Unlock()
		}
	}()

	problem, obj, err := buildProblemAndObjective(t.Request.Problem)
	if err != nil {
		m.mu.Lock()
		t.Status = api.TaskStatusFailed
		t.Error = err.Error()
		t.CompletedAt = time.Now()
		m.mu.Unlock()
		return
	}

	config := buildConfig(t.Request.Config, problem.Dimensions)

	alg, err := ga.NewAlgorithm(problem, config, obj)
	if err != nil {
		m.mu.Lock()
		t.Status = api.TaskStatusFailed
		t.Error = err.Error()
		t.CompletedAt = time.Now()
		m.mu.Unlock()
		return
	}

	t.Algorithm = alg
	alg.Initialize()

	for alg.CurrentGen < config.MaxGenerations {
		stats := alg.Step()

		m.mu.Lock()
		t.CurrentGen = alg.CurrentGen
		t.TotalGens = config.MaxGenerations
		if alg.BestInd != nil {
			t.BestFitness = alg.BestInd.Fitness
			t.BestGenes = make([]float64, len(alg.BestInd.Genes))
			copy(t.BestGenes, alg.BestInd.Genes)
		}
		m.mu.Unlock()

		if ga.CheckConvergence(alg.Statistics, config.ConvergenceThreshold, config.ConvergenceGenerations) {
			break
		}

		if config.Minimize && stats.BestFitness <= ga.Epsilon {
			total := 0.0
			for _, g := range stats.BestGenes {
				total += g * g
			}
			if total <= 1e-6 {
				break
			}
		}
	}

	m.mu.Lock()
	t.Status = api.TaskStatusCompleted
	t.CompletedAt = time.Now()
	if alg.BestInd != nil {
		t.BestFitness = alg.BestInd.Fitness
		t.BestGenes = make([]float64, len(alg.BestInd.Genes))
		copy(t.BestGenes, alg.BestInd.Genes)
	}
	m.mu.Unlock()
}

func buildProblemAndObjective(spec *api.ProblemSpec) (*ga.Problem, func([]float64) float64, error) {
	if spec == nil {
		return nil, nil, errors.New("problem spec is required")
	}
	if spec.Dimensions <= 0 {
		return nil, nil, errors.New("dimensions must be greater than 0")
	}

	var lower, upper []float64
	var obj func([]float64) float64

	if spec.CustomExpr != "" {
		variables := spec.Variables
		if len(variables) == 0 {
			variables = make([]string, spec.Dimensions)
			for i := 0; i < spec.Dimensions; i++ {
				variables[i] = string(rune('x' + i))
			}
		}
		if len(variables) != spec.Dimensions {
			return nil, nil, errors.New("variables count must match dimensions")
		}

		customFn, err := testfuncs.ParseCustomFunction(spec.CustomExpr, variables)
		if err != nil {
			return nil, nil, err
		}
		obj = customFn.Evaluate

		lower = spec.LowerBound
		upper = spec.UpperBound
		if len(lower) == 0 {
			lower = make([]float64, spec.Dimensions)
			for i := range lower {
				lower[i] = -10.0
			}
		}
		if len(upper) == 0 {
			upper = make([]float64, spec.Dimensions)
			for i := range upper {
				upper[i] = 10.0
			}
		}
	} else {
		testFn, err := testfuncs.GetFunction(spec.Function)
		if err != nil {
			return nil, nil, err
		}
		obj = testFn.Function

		lower = spec.LowerBound
		upper = spec.UpperBound
		if len(lower) == 0 || len(upper) == 0 {
			defLower, defUpper, err := testfuncs.GetDefaultBounds(spec.Function, spec.Dimensions)
			if err != nil {
				return nil, nil, err
			}
			if len(lower) == 0 {
				lower = defLower
			}
			if len(upper) == 0 {
				upper = defUpper
			}
		}
	}

	if len(lower) != spec.Dimensions {
		return nil, nil, errors.New("lower bound dimensions mismatch")
	}
	if len(upper) != spec.Dimensions {
		return nil, nil, errors.New("upper bound dimensions mismatch")
	}

	problem := &ga.Problem{
		Dimensions: spec.Dimensions,
		LowerBound: lower,
		UpperBound: upper,
		Minimize:   spec.Minimize,
	}

	return problem, obj, nil
}

func buildConfig(cfg *api.AlgorithmConfig, dimensions int) *ga.Config {
	def := ga.DefaultConfig(dimensions)
	if cfg == nil {
		return def
	}

	if cfg.PopulationSize > 1 {
		def.PopulationSize = cfg.PopulationSize
	}
	if cfg.MaxGenerations > 0 {
		def.MaxGenerations = cfg.MaxGenerations
	}
	if cfg.EliteCount >= 0 && cfg.EliteCount < def.PopulationSize {
		def.EliteCount = cfg.EliteCount
	}
	if cfg.CrossoverRate >= 0 && cfg.CrossoverRate <= 1 {
		def.CrossoverRate = cfg.CrossoverRate
	}
	if cfg.MutationRate >= 0 && cfg.MutationRate <= 1 {
		def.MutationRate = cfg.MutationRate
	}
	if cfg.SBXIndex > 0 {
		def.SBXIndex = cfg.SBXIndex
	}
	if cfg.InitialMutationSigma > 0 {
		def.InitialMutationSigma = cfg.InitialMutationSigma
	}
	if cfg.FinalMutationSigma > 0 {
		def.FinalMutationSigma = cfg.FinalMutationSigma
	}

	return def
}

func generateTaskID() string {
	return "task-" + time.Now().Format("20060102150405.000000000")
}
