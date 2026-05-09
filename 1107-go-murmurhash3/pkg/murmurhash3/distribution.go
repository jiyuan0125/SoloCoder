package murmurhash3

import (
	"math"
)

type DistributionReport struct {
	Min          uint64
	Max          uint64
	Mean         float64
	StdDev       float64
	Variance     float64
	ChiSquare    float64
	UniformScore float64
	BucketCount  int
	SampleCount  int
}

func AnalyzeDistribution32(data [][]byte, seed uint32, bucketCount int) DistributionReport {
	n := len(data)
	if n == 0 || bucketCount <= 0 {
		return DistributionReport{}
	}

	buckets := make([]int, bucketCount)
	values := make([]uint64, n)

	for i, d := range data {
		h := Hash32(d, seed)
		bucket := int(h % uint32(bucketCount))
		buckets[bucket]++
		values[i] = uint64(h)
	}

	return buildReport(values, buckets, n, bucketCount)
}

func AnalyzeDistribution128(data [][]byte, seed uint32, bucketCount int) DistributionReport {
	n := len(data)
	if n == 0 || bucketCount <= 0 {
		return DistributionReport{}
	}

	buckets := make([]int, bucketCount)
	values := make([]uint64, n)

	for i, d := range data {
		h1, h2 := Hash128(d, seed)
		combined := h1 ^ h2
		bucket := int(combined % uint64(bucketCount))
		buckets[bucket]++
		values[i] = combined
	}

	return buildReport(values, buckets, n, bucketCount)
}

func buildReport(values []uint64, buckets []int, n, bucketCount int) DistributionReport {
	var min, max uint64
	if n > 0 {
		min, max = values[0], values[0]
	}

	var sum float64
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += float64(v)
	}

	mean := sum / float64(n)

	var variance float64
	for _, v := range values {
		diff := float64(v) - mean
		variance += diff * diff
	}
	variance /= float64(n)
	stdDev := math.Sqrt(variance)

	expected := float64(n) / float64(bucketCount)
	var chiSquare float64
	for _, observed := range buckets {
		diff := float64(observed) - expected
		if expected > 0 {
			chiSquare += (diff * diff) / expected
		}
	}

	uniformScore := calculateUniformScore(chiSquare, bucketCount, n)

	return DistributionReport{
		Min:          min,
		Max:          max,
		Mean:         mean,
		StdDev:       stdDev,
		Variance:     variance,
		ChiSquare:    chiSquare,
		UniformScore: uniformScore,
		BucketCount:  bucketCount,
		SampleCount:  n,
	}
}

func calculateUniformScore(chiSquare float64, bucketCount, sampleCount int) float64 {
	df := bucketCount - 1
	if df <= 0 {
		return 0.0
	}

	normalizedChi := chiSquare / float64(df)

	var score float64
	if normalizedChi <= 1.0 {
		score = 1.0 - (normalizedChi * 0.3)
	} else if normalizedChi <= 2.0 {
		score = 0.7 - ((normalizedChi - 1.0) * 0.3)
	} else if normalizedChi <= 3.0 {
		score = 0.4 - ((normalizedChi - 2.0) * 0.3)
	} else {
		score = math.Max(0.0, 0.1 - ((normalizedChi - 3.0) * 0.05))
	}

	return score
}
