package probdist

import (
	"errors"
	"math/rand"
)

type Uniform struct {
	a float64
	b float64
}

func NewUniform(a, b float64) (*Uniform, error) {
	if a >= b {
		return nil, errors.New("a must be less than b")
	}
	return &Uniform{a: a, b: b}, nil
}

func (u *Uniform) PDF(x float64) float64 {
	if x < u.a || x > u.b {
		return 0
	}
	return 1.0 / (u.b - u.a)
}

func (u *Uniform) CDF(x float64) float64 {
	if x < u.a {
		return 0
	}
	if x > u.b {
		return 1
	}
	return (x - u.a) / (u.b - u.a)
}

func (u *Uniform) Sample(nSamples int) ([]float64, error) {
	if nSamples <= 0 {
		return nil, errors.New("sample count must be positive")
	}
	samples := make([]float64, nSamples)
	for i := 0; i < nSamples; i++ {
		samples[i] = u.a + rand.Float64()*(u.b-u.a)
	}
	return samples, nil
}
