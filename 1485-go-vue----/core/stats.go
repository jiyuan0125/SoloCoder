package core

import (
	"math"
)

type StatsCalculator struct {
	store *Store
}

func NewStatsCalculator(store *Store) *StatsCalculator {
	return &StatsCalculator{store: store}
}

func (sc *StatsCalculator) CalculateTaskStats(taskID string) (*TaskStats, error) {
	task, err := sc.store.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	dataList, err := sc.store.GetSensorData(taskID)
	if err != nil {
		return nil, err
	}

	if len(dataList) == 0 {
		return &TaskStats{
			TaskID:   taskID,
			PassRate: 1.0,
		}, nil
	}

	shouldMonitor, err := ShouldMonitor(task.CargoType)
	if err != nil {
		return nil, err
	}

	totalSamples := len(dataList)
	validSamples := totalSamples
	passCount := totalSamples

	var temperatures []float64
	for _, data := range dataList {
		temperatures = append(temperatures, data.Temperature)
	}

	if shouldMonitor {
		limits := TemperatureLimits{
			Min: task.TempMin,
			Max: task.TempMax,
		}

		passCount = 0
		for _, data := range dataList {
			if IsTemperatureWithinLimits(data.Temperature, limits) {
				passCount++
			}
		}
	}

	passRate := float64(passCount) / float64(totalSamples)

	stats := &TaskStats{
		TaskID:       taskID,
		TotalSamples: totalSamples,
		ValidSamples: validSamples,
		PassRate:     passRate,
		NeedRecheck:  passRate < PassRateThreshold,
		TempMax:      calculateMax(temperatures),
		TempMin:      calculateMin(temperatures),
		TempAvg:      calculateMean(temperatures),
		TempStdDev:   calculateStdDev(temperatures),
	}

	return stats, nil
}

func calculateMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func calculateMin(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateStdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := calculateMean(values)
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	return math.Sqrt(variance)
}
