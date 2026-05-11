package benchmark

import (
	"github.com/example/benchcompare/pkg/api"
	"sort"
)

func AggregateBenchmarks(raw map[string][]RawBenchmark) map[string]api.BenchmarkResult {
	aggregated := make(map[string]api.BenchmarkResult)
	
	for name, runs := range raw {
		if len(runs) == 0 {
			continue
		}
		
		var nsValues []float64
		var bytesValues []int
		var allocsValues []int
		hasUncertain := false
		
		for _, run := range runs {
			nsValues = append(nsValues, run.NsPerOp)
			if run.BytesPerOp > 0 {
				bytesValues = append(bytesValues, run.BytesPerOp)
			}
			if run.AllocsPerOp > 0 {
				allocsValues = append(allocsValues, run.AllocsPerOp)
			}
			if run.Uncertain {
				hasUncertain = true
			}
		}
		
		medianNs := medianFloat64(nsValues)
		medianBytes := 0
		medianAllocs := 0
		
		if len(bytesValues) > 0 {
			medianBytes = medianInt(bytesValues)
		}
		if len(allocsValues) > 0 {
			medianAllocs = medianInt(allocsValues)
		}
		
		result := api.BenchmarkResult{
			Name:        name,
			Iterations:  runs[len(runs)-1].Iterations,
			NsPerOp:     medianNs,
			BytesPerOp:  medianBytes,
			AllocsPerOp: medianAllocs,
			Uncertain:   hasUncertain,
		}
		
		aggregated[name] = result
	}
	
	return aggregated
}

func medianFloat64(values []float64) float64 {
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	
	n := len(sorted)
	if n == 0 {
		return 0
	}
	
	if n%2 == 1 {
		return sorted[n/2]
	}
	
	return (sorted[n/2-1] + sorted[n/2]) / 2.0
}

func medianInt(values []int) int {
	sorted := make([]int, len(values))
	copy(sorted, values)
	sort.Ints(sorted)
	
	n := len(sorted)
	if n == 0 {
		return 0
	}
	
	if n%2 == 1 {
		return sorted[n/2]
	}
	
	return (sorted[n/2-1] + sorted[n/2]) / 2
}
