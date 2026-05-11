package probdist

import (
	"errors"
	"math"
	"math/rand"
)

type Poisson struct {
	lambda float64
}

func NewPoisson(lambda float64) (*Poisson, error) {
	if lambda <= 0 {
		return nil, errors.New("lambda must be positive")
	}
	return &Poisson{lambda: lambda}, nil
}

func (p *Poisson) PMF(k int) float64 {
	if k < 0 {
		return 0
	}
	if p.lambda < 10 {
		return p.pmfDirect(k)
	}
	return p.pmfLog(k)
}

func (p *Poisson) pmfDirect(k int) float64 {
	logFact := 0.0
	for i := 2; i <= k; i++ {
		logFact += math.Log(float64(i))
	}
	logPMF := float64(k)*math.Log(p.lambda) - p.lambda - logFact
	return math.Exp(logPMF)
}

func (p *Poisson) pmfLog(k int) float64 {
	logFact := logGamma(float64(k + 1))
	logPMF := float64(k)*math.Log(p.lambda) - p.lambda - logFact
	return math.Exp(logPMF)
}

func (p *Poisson) CDF(k int) float64 {
	if k < 0 {
		return 0
	}
	if p.lambda < 100 {
		var cdf float64
		pmf := math.Exp(-p.lambda)
		cdf = pmf
		for i := 1; i <= k; i++ {
			pmf *= p.lambda / float64(i)
			cdf += pmf
		}
		if cdf > 1 {
			cdf = 1
		}
		return cdf
	}
	z := (float64(k) + 0.5 - p.lambda) / math.Sqrt(p.lambda)
	cdf := 0.5 * (1 + math.Erf(z/math.Sqrt2))
	if cdf < 0 {
		cdf = 0
	}
	if cdf > 1 {
		cdf = 1
	}
	return cdf
}

func (p *Poisson) Sample(nSamples int) ([]int, error) {
	if nSamples <= 0 {
		return nil, errors.New("sample count must be positive")
	}
	samples := make([]int, nSamples)
	for i := 0; i < nSamples; i++ {
		u := rand.Float64()
		samples[i] = p.inverseCDF(u)
	}
	return samples, nil
}

func (p *Poisson) inverseCDF(u float64) int {
	if u <= 0 {
		return 0
	}
	if u >= 1 {
		return int(p.lambda) + 10
	}
	if p.lambda < 10 {
		return p.inverseCDFNaive(u)
	}
	return p.inverseCDFApprox(u)
}

func (p *Poisson) inverseCDFNaive(u float64) int {
	var cdf float64
	k := 0
	for {
		cdf += p.PMF(k)
		if u < cdf {
			return k
		}
		k++
	}
}

func (p *Poisson) inverseCDFApprox(u float64) int {
	mu := p.lambda
	sigma := math.Sqrt(p.lambda)
	k := int(math.Max(0, mu+math.Sqrt2*sigma*erfinv(2*u-1)))
	for p.CDF(k) < u {
		k++
	}
	for k > 0 && p.CDF(k-1) >= u {
		k--
	}
	return k
}

func logGamma(x float64) float64 {
	if x <= 0 {
		return math.Inf(1)
	}
	if x == 1 || x == 2 {
		return 0
	}
	if x < 1 {
		return logGamma(x+1) - math.Log(x)
	}
	const coef = 12
	series := 1.0 / (12 * x)
	series -= 1.0 / (360 * math.Pow(x, 3))
	series += 1.0 / (1260 * math.Pow(x, 5))
	series -= 1.0 / (1680 * math.Pow(x, 7))
	series += 1.0 / (1188 * math.Pow(x, 9))
	series -= 691.0 / (360360 * math.Pow(x, 11))
	series *= coef
	result := (x-0.5)*math.Log(x) - x + 0.5*math.Log(2*math.Pi) + series
	return result
}
