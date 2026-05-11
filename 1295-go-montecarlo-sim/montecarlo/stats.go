package montecarlo

import (
	"errors"
	"math"
)

type SimulationResult struct {
	Estimate     float64
	StdDev       float64
	StdErr       float64
	Confidence   float64
	CIUpper      float64
	CILower      float64
	SampleCount  int
	Warnings     []string
}

func zScore(confidence float64) float64 {
	switch confidence {
	case 0.90:
		return 1.645
	case 0.95:
		return 1.96
	case 0.98:
		return 2.326
	case 0.99:
		return 2.576
	case 0.999:
		return 3.291
	case 0.9999:
		return 3.891
	default:
		if confidence <= 0.5 {
			return 0.674
		}
		return 1.96
	}
}

func computeStats(values []float64, confidence float64, sampleCount int) (*SimulationResult, error) {
	if sampleCount <= 0 {
		return nil, errors.New("采样次数必须为正整数")
	}

	n := len(values)
	if n == 0 {
		return nil, errors.New("采样值列表为空")
	}

	var sum, sumSq float64
	for _, v := range values {
		sum += v
		sumSq += v * v
	}
	mean := sum / float64(n)
	variance := (sumSq / float64(n)) - mean*mean
	if variance < 0 {
		variance = 0
	}
	stdDev := math.Sqrt(variance)
	stdErr := stdDev / math.Sqrt(float64(sampleCount))

	warnings := []string{}
	if sampleCount < 30 {
		warnings = append(warnings, "采样次数较少，中心极限定理近似效果可能不佳")
	}
	if confidence >= 0.99 && sampleCount < 100 {
		warnings = append(warnings, "高置信水平配合低采样次数会导致置信区间过宽，结果可能无意义")
	}

	z := zScore(confidence)
	margin := z * stdErr
	ciLower := mean - margin
	ciUpper := mean + margin

	return &SimulationResult{
		Estimate:    mean,
		StdDev:      stdDev,
		StdErr:      stdErr,
		Confidence:  confidence,
		CIUpper:     ciUpper,
		CILower:     ciLower,
		SampleCount: sampleCount,
		Warnings:    warnings,
	}, nil
}
