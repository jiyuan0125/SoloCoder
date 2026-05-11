package montecarlo

import (
	"errors"
	"math"
)

type OptionType string

const (
	Call OptionType = "call"
	Put  OptionType = "put"
)

type OptionParams struct {
	S0          float64
	K           float64
	T           float64
	R           float64
	Sigma       float64
	Mu          float64
	Type        OptionType
	Steps       int
}

func (p *OptionParams) Validate() error {
	if p.S0 < 0 {
		return errors.New("当前股价S0不能为负")
	}
	if p.K < 0 {
		return errors.New("执行价K不能为负")
	}
	if p.T <= 0 {
		return errors.New("到期时间T必须为正")
	}
	if p.Sigma < 0 {
		return errors.New("波动率Sigma不能为负")
	}
	if p.Steps <= 0 {
		p.Steps = 1
	}
	if p.Type != Call && p.Type != Put {
		return errors.New("期权类型必须为call或put")
	}
	return nil
}

func PriceEuropeanOption(params *OptionParams, samples int, confidence float64) (*SimulationResult, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	dt := params.T / float64(params.Steps)
	discount := math.Exp(-params.R * params.T)

	if params.Sigma == 0 {
		ST := params.S0 * math.Exp(params.Mu*params.T)
		var payoff float64
		if params.Type == Call {
			payoff = math.Max(ST-params.K, 0)
		} else {
			payoff = math.Max(params.K-ST, 0)
		}
		price := discount * payoff
		return &SimulationResult{
			Estimate:    price,
			StdDev:      0,
			StdErr:      0,
			Confidence:  confidence,
			CIUpper:     price,
			CILower:     price,
			SampleCount: 1,
			Warnings:    []string{"波动率为零，直接解析计算，无需模拟"},
		}, nil
	}

	if params.K == 0 && params.Type == Call {
		price := params.S0 * math.Exp((params.Mu-params.R)*params.T)
		return &SimulationResult{
			Estimate:    price,
			StdDev:      0,
			StdErr:      0,
			Confidence:  confidence,
			CIUpper:     price,
			CILower:     price,
			SampleCount: 1,
			Warnings:    []string{"执行价为零的看涨期权等价于持有股票，直接解析计算"},
		}, nil
	}

	values := make([]float64, samples)
	for i := 0; i < samples; i++ {
		S := params.S0
		for step := 0; step < params.Steps; step++ {
			Z := RandomNormal()
			S = S * math.Exp((params.Mu-0.5*params.Sigma*params.Sigma)*dt + params.Sigma*math.Sqrt(dt)*Z)
		}
		var payoff float64
		if params.Type == Call {
			payoff = math.Max(S-params.K, 0)
		} else {
			payoff = math.Max(params.K-S, 0)
		}
		values[i] = discount * payoff
	}

	return computeStats(values, confidence, samples)
}
