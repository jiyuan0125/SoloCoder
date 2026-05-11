package regression

import (
	"errors"
	"math"
)

type Model struct {
	isSimple   bool
	intercept  float64
	slope      float64
	coefficients []float64
	featureCount int
}

func validateSimpleData(points [][]float64) error {
	if len(points) == 0 {
		return errors.New("no data points provided")
	}
	if len(points) < 2 {
		return errors.New("at least 2 data points required")
	}
	for _, pt := range points {
		if len(pt) != 2 {
			return errors.New("each point must have exactly 2 coordinates")
		}
		if math.IsNaN(pt[0]) || math.IsNaN(pt[1]) || math.IsInf(pt[0], 0) || math.IsInf(pt[1], 0) {
			return errors.New("data contains NaN or infinite values")
		}
	}
	return nil
}

func validateMultipleData(features [][]float64, target []float64) error {
	if len(features) == 0 {
		return errors.New("no data points provided")
	}
	if len(target) == 0 {
		return errors.New("target values not provided")
	}
	if len(features) != len(target) {
		return errors.New("features and target length mismatch")
	}
	
	n := len(features)
	p := len(features[0])
	
	if n < 2 {
		return errors.New("at least 2 data points required")
	}
	if p >= n {
		return errors.New("number of features must be less than number of samples")
	}
	
	for i, row := range features {
		if len(row) != p {
			return errors.New("all feature vectors must have the same length")
		}
		for _, val := range row {
			if math.IsNaN(val) || math.IsInf(val, 0) {
				return errors.New("features contain NaN or infinite values")
			}
		}
		if math.IsNaN(target[i]) || math.IsInf(target[i], 0) {
			return errors.New("target contains NaN or infinite values")
		}
	}
	
	for j := 0; j < p; j++ {
		firstVal := features[0][j]
		allSame := true
		for i := 1; i < n; i++ {
			if features[i][j] != firstVal {
				allSame = false
				break
			}
		}
		if allSame {
			return errors.New("feature variance is zero (all values are the same)")
		}
	}
	
	return nil
}

func FitSimple(points [][]float64) (*Model, float64, error) {
	if err := validateSimpleData(points); err != nil {
		return nil, 0, err
	}
	
	n := float64(len(points))
	var sumX, sumY, sumXY, sumX2 float64
	
	for _, pt := range points {
		sumX += pt[0]
		sumY += pt[1]
		sumXY += pt[0] * pt[1]
		sumX2 += pt[0] * pt[0]
	}
	
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return nil, 0, errors.New("X values are all the same, cannot fit regression")
	}
	
	slope := (n*sumXY - sumX*sumY) / denom
	intercept := (sumY - slope*sumX) / n
	
	r2 := calculateR2Simple(points, slope, intercept)
	
	return &Model{
		isSimple:  true,
		slope:     slope,
		intercept: intercept,
	}, r2, nil
}

func calculateR2Simple(points [][]float64, slope, intercept float64) float64 {
	var sumY float64
	n := len(points)
	for _, pt := range points {
		sumY += pt[1]
	}
	meanY := sumY / float64(n)
	
	var ssTotal, ssResidual float64
	for _, pt := range points {
		y := pt[1]
		yPred := slope*pt[0] + intercept
		ssResidual += (y - yPred) * (y - yPred)
		ssTotal += (y - meanY) * (y - meanY)
	}
	
	if ssTotal == 0 {
		return math.NaN()
	}
	
	return 1 - ssResidual/ssTotal
}

func FitMultiple(features [][]float64, target []float64) (*Model, float64, error) {
	if err := validateMultipleData(features, target); err != nil {
		return nil, 0, err
	}
	
	n := len(features)
	p := len(features[0])
	
	X := make([][]float64, n)
	for i := range X {
		X[i] = make([]float64, p+1)
		X[i][0] = 1
		copy(X[i][1:], features[i])
	}
	
	XT := transpose(X)
	XTX := multiply(XT, X)
	XTy := multiplyVector(XT, target)
	
	invXTX, err := inverse(XTX)
	if err != nil {
		return nil, 0, errors.New("matrix is singular, cannot compute regression coefficients")
	}
	
	beta := multiplyVector(invXTX, XTy)
	
	coefficients := make([]float64, p)
	for i := 0; i < p; i++ {
		coefficients[i] = beta[i+1]
	}
	
	r2 := calculateR2Multiple(features, target, beta[0], coefficients)
	
	return &Model{
		isSimple:   false,
		intercept:  beta[0],
		coefficients: coefficients,
		featureCount: p,
	}, r2, nil
}

func calculateR2Multiple(features [][]float64, target []float64, intercept float64, coefficients []float64) float64 {
	var sumY float64
	n := len(target)
	for _, y := range target {
		sumY += y
	}
	meanY := sumY / float64(n)
	
	var ssTotal, ssResidual float64
	for i, y := range target {
		yPred := intercept
		for j, coef := range coefficients {
			yPred += coef * features[i][j]
		}
		ssResidual += (y - yPred) * (y - yPred)
		ssTotal += (y - meanY) * (y - meanY)
	}
	
	if ssTotal == 0 {
		return math.NaN()
	}
	
	return 1 - ssResidual/ssTotal
}

func (m *Model) PredictSimple(x float64) (float64, error) {
	if !m.isSimple {
		return 0, errors.New("model is not a simple linear regression")
	}
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 0, errors.New("x contains NaN or infinite values")
	}
	return m.slope*x + m.intercept, nil
}

func (m *Model) PredictMultiple(features []float64) (float64, error) {
	if m.isSimple {
		return 0, errors.New("model is a simple linear regression")
	}
	if len(features) != m.featureCount {
		return 0, errors.New("feature count mismatch")
	}
	for _, val := range features {
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return 0, errors.New("features contain NaN or infinite values")
		}
	}
	
	yPred := m.intercept
	for i, coef := range m.coefficients {
		yPred += coef * features[i]
	}
	return yPred, nil
}

func (m *Model) Slope() float64 { return m.slope }
func (m *Model) Intercept() float64 { return m.intercept }
func (m *Model) Coefficients() []float64 { return m.coefficients }
func (m *Model) IsSimple() bool { return m.isSimple }
