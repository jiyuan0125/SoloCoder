package ga

import (
	"math"
	"math/rand"
)

func InitializePopulation(problem *Problem, size int) Population {
	pop := make(Population, size)
	for i := 0; i < size; i++ {
		genes := make([]float64, problem.Dimensions)
		for j := 0; j < problem.Dimensions; j++ {
			genes[j] = problem.LowerBound[j] + rand.Float64()*(problem.UpperBound[j]-problem.LowerBound[j])
		}
		pop[i] = &Individual{Genes: genes}
	}
	return pop
}

func EvaluatePopulation(pop Population, obj Objective) {
	for _, ind := range pop {
		if !ind.Evaluated {
			ind.Fitness = obj(ind.Genes)
			ind.Evaluated = true
		}
	}
}

func TournamentSelection(pop Population, minimize bool, tournamentSize int) *Individual {
	bestIdx := rand.Intn(len(pop))
	for i := 1; i < tournamentSize; i++ {
		idx := rand.Intn(len(pop))
		if CompareFitness(pop[idx].Fitness, pop[bestIdx].Fitness, minimize) < 0 {
			bestIdx = idx
		}
	}
	return pop[bestIdx]
}

func SBXCrossover(parent1, parent2 *Individual, problem *Problem, eta float64) (*Individual, *Individual) {
	dim := len(parent1.Genes)
	child1 := &Individual{Genes: make([]float64, dim)}
	child2 := &Individual{Genes: make([]float64, dim)}

	for i := 0; i < dim; i++ {
		if rand.Float64() <= 0.5 {
			x1 := parent1.Genes[i]
			x2 := parent2.Genes[i]
			xl := problem.LowerBound[i]
			xu := problem.UpperBound[i]

			if math.Abs(x1-x2) > Epsilon {
				var beta1, beta2 float64
				if x2 > x1 {
					beta1 = 1.0 + (2.0 * (x1 - xl) / (x2 - x1))
					beta2 = 1.0 + (2.0 * (xu - x2) / (x2 - x1))
				} else {
					beta1 = 1.0 + (2.0 * (x2 - xl) / (x1 - x2))
					beta2 = 1.0 + (2.0 * (xu - x1) / (x1 - x2))
				}

				alpha1 := 2.0 - math.Pow(1.0/beta1, eta+1.0)
				alpha2 := 2.0 - math.Pow(1.0/beta2, eta+1.0)

				var betaq1, betaq2 float64
				u := rand.Float64()
				if u <= 1.0/alpha1 {
					betaq1 = math.Pow(u*alpha1, 1.0/(eta+1.0))
				} else {
					betaq1 = math.Pow(1.0/(2.0-u*alpha1), 1.0/(eta+1.0))
				}
				if u <= 1.0/alpha2 {
					betaq2 = math.Pow(u*alpha2, 1.0/(eta+1.0))
				} else {
					betaq2 = math.Pow(1.0/(2.0-u*alpha2), 1.0/(eta+1.0))
				}

				var c1, c2 float64
				if x2 > x1 {
					c1 = 0.5 * ((x1 + x2) - betaq1*(x2-x1))
					c2 = 0.5 * ((x1 + x2) + betaq2*(x2-x1))
				} else {
					c1 = 0.5 * ((x1 + x2) - betaq2*(x1-x2))
					c2 = 0.5 * ((x1 + x2) + betaq1*(x1-x2))
				}

				c1 = clamp(c1, xl, xu)
				c2 = clamp(c2, xl, xu)

				child1.Genes[i] = c1
				child2.Genes[i] = c2
			} else {
				child1.Genes[i] = x1
				child2.Genes[i] = x2
			}
		} else {
			child1.Genes[i] = parent1.Genes[i]
			child2.Genes[i] = parent2.Genes[i]
		}
	}
	return child1, child2
}

func GaussianMutation(individual *Individual, problem *Problem, sigma float64, mutationRate float64) {
	dim := len(individual.Genes)
	for i := 0; i < dim; i++ {
		if rand.Float64() <= mutationRate {
			individual.Genes[i] += rand.NormFloat64() * sigma
			individual.Genes[i] = reflectClamp(individual.Genes[i], problem.LowerBound[i], problem.UpperBound[i])
		}
	}
	individual.Evaluated = false
}

func AdaptiveSigma(initialSigma, finalSigma float64, currentGen, maxGen int) float64 {
	t := float64(currentGen) / float64(maxGen)
	return initialSigma*math.Pow(finalSigma/initialSigma, t)
}

func clamp(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

func reflectClamp(x, min, max float64) float64 {
	for x < min || x > max {
		if x < min {
			x = min + (min - x)
		}
		if x > max {
			x = max - (x - max)
		}
	}
	return x
}

func (p Population) SortByFitness(minimize bool) {
	quickSort(p, 0, len(p)-1, minimize)
}

func quickSort(p Population, low, high int, minimize bool) {
	if low < high {
		pi := partition(p, low, high, minimize)
		quickSort(p, low, pi-1, minimize)
		quickSort(p, pi+1, high, minimize)
	}
}

func partition(p Population, low, high int, minimize bool) int {
	pivot := p[high]
	i := low - 1
	for j := low; j < high; j++ {
		if CompareFitness(p[j].Fitness, pivot.Fitness, minimize) < 0 {
			i++
			p[i], p[j] = p[j], p[i]
		}
	}
	p[i+1], p[high] = p[high], p[i+1]
	return i + 1
}

func ComputeStatistics(pop Population, generation int, minimize bool) Statistics {
	if len(pop) == 0 {
		return Statistics{Generation: generation}
	}

	sortable := make(Population, len(pop))
	copy(sortable, pop)
	sortable.SortByFitness(minimize)

	best := sortable[0]

	sum := 0.0
	for _, ind := range pop {
		sum += ind.Fitness
	}
	mean := sum / float64(len(pop))

	variance := 0.0
	for _, ind := range pop {
		diff := ind.Fitness - mean
		variance += diff * diff
	}
	variance /= float64(len(pop))
	stdDev := math.Sqrt(variance)

	bestGenes := make([]float64, len(best.Genes))
	copy(bestGenes, best.Genes)

	return Statistics{
		Generation:   generation,
		BestFitness:  best.Fitness,
		MeanFitness:  mean,
		StdDeviation: stdDev,
		BestGenes:    bestGenes,
	}
}

func CheckConvergence(stats []Statistics, threshold float64, minGens int) bool {
	if len(stats) < minGens {
		return false
	}
	recent := stats[len(stats)-minGens:]
	first := recent[0].BestFitness
	for _, s := range recent[1:] {
		if math.Abs(s.BestFitness-first) > threshold {
			return false
		}
	}
	return true
}
