package probdist

import (
	"errors"
	"math"
	"math/rand"
)

type Normal struct {
	mu    float64
	sigma float64
}

func NewNormal(mu, sigma float64) (*Normal, error) {
	if sigma <= 0 {
		return nil, errors.New("sigma must be positive")
	}
	return &Normal{mu: mu, sigma: sigma}, nil
}

func (n *Normal) PDF(x float64) float64 {
	expTerm := -math.Pow(x-n.mu, 2) / (2 * math.Pow(n.sigma, 2))
	return math.Exp(expTerm) / (n.sigma * math.Sqrt(2*math.Pi))
}

func (n *Normal) CDF(x float64) float64 {
	z := (x - n.mu) / n.sigma
	cdf := 0.5 * (1 + math.Erf(z/math.Sqrt2))
	if cdf < 0 {
		cdf = 0
	}
	if cdf > 1 {
		cdf = 1
	}
	return cdf
}

func (n *Normal) Sample(nSamples int) ([]float64, error) {
	if nSamples <= 0 {
		return nil, errors.New("sample count must be positive")
	}
	samples := make([]float64, nSamples)
	for i := 0; i < nSamples; i++ {
		u := rand.Float64()
		samples[i] = n.mu + n.sigma*math.Sqrt2*erfinv(2*u-1)
	}
	return samples, nil
}

func erfinv(y float64) float64 {
	const a1 = 0.8862269254527579
	const a2 = 0.2308489888928553
	const a3 = 0.00972849021228299
	const a4 = -0.00131753559850638
	const a5 = 1.772453850905516
	y = math.Min(math.Max(y, -0.999999999999), 0.999999999999)
	sgn := 1.0
	if y < 0 {
		sgn = -1
		y = -y
	}
	z := math.Log(1 - y*y)
	var result float64
	if z > -3.978 {
		z = math.Sqrt(-z)
		result = a1*z + a2*z*z + a3*z*z*z
	} else {
		z = math.Sqrt(-z)
		result = a5*z + a4*z*z
	}
	return sgn * result
}
