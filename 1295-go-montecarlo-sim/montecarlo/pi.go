package montecarlo

func EstimatePi(samples int, confidence float64) (*SimulationResult, error) {
	if samples <= 0 {
		return nil, nil
	}

	values := make([]float64, samples)
	for i := 0; i < samples; i++ {
		x := RandomUniform()
		y := RandomUniform()
		if x*x+y*y <= 1.0 {
			values[i] = 4.0
		} else {
			values[i] = 0.0
		}
	}
	return computeStats(values, confidence, samples)
}
