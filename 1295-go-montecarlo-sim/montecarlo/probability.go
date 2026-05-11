package montecarlo

import (
	"errors"
)

type EventCheck func() bool

func EstimateProbability(check EventCheck, samples int, confidence float64) (*SimulationResult, error) {
	if check == nil {
		return nil, errors.New("事件检查函数不能为空")
	}

	values := make([]float64, samples)
	for i := 0; i < samples; i++ {
		if check() {
			values[i] = 1.0
		} else {
			values[i] = 0.0
		}
	}
	return computeStats(values, confidence, samples)
}
