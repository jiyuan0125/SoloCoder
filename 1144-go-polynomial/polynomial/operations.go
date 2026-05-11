package polynomial

import (
	"errors"
	"math"
)

func Add(a, b Polynomial) Polynomial {
	maxLen := max(len(a), len(b))
	aPadded := pad(a, maxLen)
	bPadded := pad(b, maxLen)
	
	result := make(Polynomial, maxLen)
	for i := 0; i < maxLen; i++ {
		result[i] = aPadded[i] + bPadded[i]
	}
	return trim(result)
}

func Sub(a, b Polynomial) Polynomial {
	maxLen := max(len(a), len(b))
	aPadded := pad(a, maxLen)
	bPadded := pad(b, maxLen)
	
	result := make(Polynomial, maxLen)
	for i := 0; i < maxLen; i++ {
		result[i] = aPadded[i] - bPadded[i]
	}
	return trim(result)
}

func Mul(a, b Polynomial) Polynomial {
	if a.IsZero() || b.IsZero() {
		return Polynomial{0}
	}
	
	resultLen := len(a) + len(b) - 1
	result := make(Polynomial, resultLen)
	
	for i := 0; i < len(a); i++ {
		if math.Abs(a[i]) > 1e-10 {
			for j := 0; j < len(b); j++ {
				if math.Abs(b[j]) > 1e-10 {
					result[i+j] += a[i] * b[j]
				}
			}
		}
	}
	
	return trim(result)
}

func Div(dividend, divisor Polynomial) (quotient, remainder Polynomial, err error) {
	if divisor.IsZero() {
		return nil, nil, errors.New("除数不能为零多项式")
	}
	
	divDeg := dividend.Degree()
	divisorDeg := divisor.Degree()
	
	if divDeg < divisorDeg {
		return Polynomial{0}, dividend.Clone(), nil
	}
	
	quotient = make(Polynomial, 0)
	remainder = dividend.Clone()
	
	for remainder.Degree() >= divisorDeg {
		leadCoeff := remainder[remainder.Degree()] / divisor[divisorDeg]
		leadDeg := remainder.Degree() - divisorDeg
		
		term := make(Polynomial, leadDeg+1)
		term[leadDeg] = leadCoeff
		
		quotient = Add(quotient, term)
		remainder = Sub(remainder, Mul(term, divisor))
	}
	
	return trim(quotient), trim(remainder), nil
}

func Eval(p Polynomial, x float64) float64 {
	if p.IsZero() {
		return 0
	}
	
	result := 0.0
	power := 1.0
	for _, coeff := range p {
		result += coeff * power
		power *= x
	}
	return result
}

func Derivative(p Polynomial) Polynomial {
	if p.IsZero() || p.Degree() == 0 {
		return Polynomial{0}
	}
	
	result := make(Polynomial, p.Degree())
	for i := 1; i < len(p); i++ {
		if i-1 < len(result) {
			result[i-1] = float64(i) * p[i]
		}
	}
	
	return trim(result)
}
