package common

import "fmt"

type SimulationType string

const (
	TypePi          SimulationType = "pi"
	TypeOption      SimulationType = "option"
	TypeProbability SimulationType = "probability"
)

type ProbabilityEvent string

const (
	EventDiceSumGreater ProbabilityEvent = "dice_sum_greater"
	EventCoinHeads      ProbabilityEvent = "coin_heads"
)

type PiRequest struct {
	Samples    int     `json:"samples"`
	Confidence float64 `json:"confidence"`
}

type OptionRequest struct {
	Samples    int        `json:"samples"`
	Confidence float64    `json:"confidence"`
	S0         float64    `json:"s0"`
	K          float64    `json:"k"`
	T          float64    `json:"t"`
	R          float64    `json:"r"`
	Sigma      float64    `json:"sigma"`
	Mu         float64    `json:"mu"`
	Type       string     `json:"type"`
	Steps      int        `json:"steps"`
}

type ProbabilityRequest struct {
	Samples     int               `json:"samples"`
	Confidence  float64           `json:"confidence"`
	Event       ProbabilityEvent  `json:"event"`
	EventParams map[string]interface{} `json:"event_params"`
}

type SimulationRequest struct {
	Type        SimulationType    `json:"type"`
	Pi          *PiRequest        `json:"pi,omitempty"`
	Option      *OptionRequest    `json:"option,omitempty"`
	Probability *ProbabilityRequest `json:"probability,omitempty"`
}

type SimulationResponse struct {
	Success  bool     `json:"success"`
	Message  string   `json:"message,omitempty"`
	Estimate float64  `json:"estimate,omitempty"`
	StdDev   float64  `json:"std_dev,omitempty"`
	StdErr   float64  `json:"std_err,omitempty"`
	CIUpper  float64  `json:"ci_upper,omitempty"`
	CILower  float64  `json:"ci_lower,omitempty"`
	Samples  int      `json:"samples,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

func (r *SimulationRequest) Validate() error {
	switch r.Type {
	case TypePi:
		if r.Pi == nil {
			return fmt.Errorf("pi参数缺失")
		}
		if r.Pi.Samples <= 0 {
			return fmt.Errorf("采样次数必须为正整数")
		}
	case TypeOption:
		if r.Option == nil {
			return fmt.Errorf("option参数缺失")
		}
		if r.Option.Samples <= 0 {
			return fmt.Errorf("采样次数必须为正整数")
		}
		if r.Option.Type != "call" && r.Option.Type != "put" {
			return fmt.Errorf("期权类型必须为call或put")
		}
	case TypeProbability:
		if r.Probability == nil {
			return fmt.Errorf("probability参数缺失")
		}
		if r.Probability.Samples <= 0 {
			return fmt.Errorf("采样次数必须为正整数")
		}
	default:
		return fmt.Errorf("未知的模拟类型: %s", r.Type)
	}
	return nil
}
