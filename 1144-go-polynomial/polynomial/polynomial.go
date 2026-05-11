package polynomial

import (
	"errors"
	"math"
)

type Polynomial []float64

var (
	ErrEmptyPolynomial = errors.New("多项式系数数组不能为空")
	ErrInvalidDegree   = errors.New("多项式次数必须为非负整数")
)

func New(coeffs []float64) (Polynomial, error) {
	if len(coeffs) == 0 {
		return nil, ErrEmptyPolynomial
	}
	result := make(Polynomial, len(coeffs))
	copy(result, coeffs)
	return trim(result), nil
}

func (p Polynomial) Degree() int {
	for i := len(p) - 1; i >= 0; i-- {
		if math.Abs(p[i]) > 1e-10 {
			return i
		}
	}
	return -1
}

func (p Polynomial) IsZero() bool {
	return p.Degree() == -1
}

func (p Polynomial) Clone() Polynomial {
	result := make(Polynomial, len(p))
	copy(result, p)
	return result
}

func trim(p Polynomial) Polynomial {
	degree := p.Degree()
	if degree == -1 {
		return Polynomial{0}
	}
	return p[:degree+1]
}

func pad(p Polynomial, length int) Polynomial {
	if len(p) >= length {
		return p
	}
	result := make(Polynomial, length)
	copy(result, p)
	return result
}
