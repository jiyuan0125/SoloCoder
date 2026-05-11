package ga

import "errors"

type Config struct {
	PopulationSize   int
	MaxGenerations   int
	EliteCount       int
	CrossoverRate    float64
	MutationRate     float64
	SBXIndex         float64
	InitialMutationSigma float64
	FinalMutationSigma   float64
	Minimize         bool
	ConvergenceThreshold float64
	ConvergenceGenerations int
}

func DefaultConfig(dimensions int) *Config {
	return &Config{
		PopulationSize:   max(50, 10*dimensions),
		MaxGenerations:   200,
		EliteCount:       2,
		CrossoverRate:    0.8,
		MutationRate:     0.1,
		SBXIndex:         20.0,
		InitialMutationSigma: 0.5,
		FinalMutationSigma:   0.01,
		Minimize:         true,
		ConvergenceThreshold: 1e-10,
		ConvergenceGenerations: 20,
	}
}

func (c *Config) Validate(problem *Problem) error {
	if problem == nil {
		return errors.New("problem is nil")
	}
	if problem.Dimensions <= 0 {
		return errors.New("dimensions must be greater than 0")
	}
	if len(problem.LowerBound) != problem.Dimensions || len(problem.UpperBound) != problem.Dimensions {
		return errors.New("boundaries must match dimensions")
	}
	for i := 0; i < problem.Dimensions; i++ {
		if problem.LowerBound[i] >= problem.UpperBound[i] {
			return errors.New("lower bound must be less than upper bound")
		}
	}

	if c.PopulationSize <= 1 {
		return errors.New("population size must be greater than 1")
	}
	if c.MaxGenerations <= 0 {
		return errors.New("max generations must be greater than 0")
	}
	if c.EliteCount < 0 || c.EliteCount >= c.PopulationSize {
		return errors.New("elite count must be between 0 and population size-1")
	}
	if c.CrossoverRate < 0 || c.CrossoverRate > 1 {
		return errors.New("crossover rate must be between 0 and 1")
	}
	if c.MutationRate < 0 || c.MutationRate > 1 {
		return errors.New("mutation rate must be between 0 and 1")
	}
	if c.SBXIndex <= 0 {
		return errors.New("SBX index must be positive")
	}
	if c.InitialMutationSigma <= 0 || c.FinalMutationSigma <= 0 {
		return errors.New("mutation sigma must be positive")
	}
	if c.ConvergenceGenerations <= 0 {
		c.ConvergenceGenerations = 20
	}
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
