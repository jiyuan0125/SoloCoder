package ga

import (
	"errors"
	"math"
	"math/rand"
)

type Algorithm struct {
	Config      *Config
	Problem     *Problem
	Objective   Objective
	Population  Population
	Statistics  []Statistics
	CurrentGen  int
	Running     bool
	BestInd     *Individual
}

func NewAlgorithm(problem *Problem, config *Config, obj Objective) (*Algorithm, error) {
	if problem == nil {
		return nil, errors.New("problem is required")
	}
	if config == nil {
		config = DefaultConfig(problem.Dimensions)
	}
	if err := config.Validate(problem); err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, errors.New("objective function is required")
	}

	return &Algorithm{
		Config:     config,
		Problem:    problem,
		Objective:  obj,
		CurrentGen: 0,
		Running:    false,
	}, nil
}

func (ga *Algorithm) Initialize() {
	ga.Population = InitializePopulation(ga.Problem, ga.Config.PopulationSize)
	ga.Statistics = make([]Statistics, 0)
	ga.CurrentGen = 0
}

func (ga *Algorithm) Step() Statistics {
	if ga.Population == nil {
		ga.Initialize()
	}

	EvaluatePopulation(ga.Population, ga.Objective)
	stats := ComputeStatistics(ga.Population, ga.CurrentGen, ga.Config.Minimize)
	ga.Statistics = append(ga.Statistics, stats)

	if ga.BestInd == nil || CompareFitness(ga.Population[0].Fitness, ga.BestInd.Fitness, ga.Config.Minimize) < 0 {
		best := &Individual{
			Genes:   make([]float64, len(ga.Population[0].Genes)),
			Fitness: ga.Population[0].Fitness,
		}
		copy(best.Genes, ga.Population[0].Genes)
		ga.BestInd = best
	}

	newPop := make(Population, 0, ga.Config.PopulationSize)

	ga.Population.SortByFitness(ga.Config.Minimize)
	for i := 0; i < ga.Config.EliteCount; i++ {
		elite := &Individual{
			Genes:   make([]float64, len(ga.Population[i].Genes)),
			Fitness: ga.Population[i].Fitness,
		}
		copy(elite.Genes, ga.Population[i].Genes)
		elite.Evaluated = true
		newPop = append(newPop, elite)
	}

	tournamentSize := max(2, ga.Config.PopulationSize/10)

	currentSigma := AdaptiveSigma(
		ga.Config.InitialMutationSigma,
		ga.Config.FinalMutationSigma,
		ga.CurrentGen,
		ga.Config.MaxGenerations,
	)

	for len(newPop) < ga.Config.PopulationSize {
		parent1 := TournamentSelection(ga.Population, ga.Config.Minimize, tournamentSize)
		parent2 := TournamentSelection(ga.Population, ga.Config.Minimize, tournamentSize)

		var child1, child2 *Individual
		if rand.Float64() <= ga.Config.CrossoverRate {
			child1, child2 = SBXCrossover(parent1, parent2, ga.Problem, ga.Config.SBXIndex)
		} else {
			child1 = &Individual{Genes: make([]float64, len(parent1.Genes))}
			child2 = &Individual{Genes: make([]float64, len(parent2.Genes))}
			copy(child1.Genes, parent1.Genes)
			copy(child2.Genes, parent2.Genes)
		}

		GaussianMutation(child1, ga.Problem, currentSigma, ga.Config.MutationRate)
		GaussianMutation(child2, ga.Problem, currentSigma, ga.Config.MutationRate)

		newPop = append(newPop, child1)
		if len(newPop) < ga.Config.PopulationSize {
			newPop = append(newPop, child2)
		}
	}

	ga.Population = newPop[:ga.Config.PopulationSize]
	ga.CurrentGen++

	return stats
}

func (ga *Algorithm) Run() (*Individual, []Statistics, error) {
	ga.Initialize()
	ga.Running = true

	for ga.CurrentGen < ga.Config.MaxGenerations {
		stats := ga.Step()

		if CheckConvergence(ga.Statistics, ga.Config.ConvergenceThreshold, ga.Config.ConvergenceGenerations) {
			break
		}

		if ga.Config.Minimize && stats.BestFitness <= Epsilon {
			if math.Abs(stats.BestFitness) <= Epsilon {
				break
			}
		}
	}

	ga.Running = false
	return ga.BestInd, ga.Statistics, nil
}

func (ga *Algorithm) GetBest() *Individual {
	if ga.BestInd == nil {
		return nil
	}
	best := &Individual{
		Genes:   make([]float64, len(ga.BestInd.Genes)),
		Fitness: ga.BestInd.Fitness,
	}
	copy(best.Genes, ga.BestInd.Genes)
	return best
}
