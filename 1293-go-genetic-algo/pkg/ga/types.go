package ga

import "math"

type Individual struct {
	Genes     []float64
	Fitness   float64
	Evaluated bool
}

type Population []*Individual

type Objective func(genes []float64) float64

type Problem struct {
	Dimensions int
	LowerBound []float64
	UpperBound []float64
	Minimize   bool
}

type Statistics struct {
	Generation   int
	BestFitness  float64
	MeanFitness  float64
	StdDeviation float64
	BestGenes    []float64
}

type ConvergenceChecker interface {
	ShouldStop(stats []Statistics) bool
}

const Epsilon = 1e-15

func CompareFitness(a, b float64, minimize bool) int {
	if math.Abs(a-b) < Epsilon {
		return 0
	}
	if minimize {
		if a < b {
			return -1
		}
		return 1
	}
	if a > b {
		return -1
	}
	return 1
}

func (p Population) Len() int { return len(p) }

func (p Population) Less(i, j int) bool {
	if !p[i].Evaluated || !p[j].Evaluated {
		return false
	}
	return p[i].Fitness < p[j].Fitness
}

func (p Population) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
