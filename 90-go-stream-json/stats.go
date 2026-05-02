package main

import "fmt"

type StatsMode int

const (
	ModeCount StatsMode = iota
	ModeSum
	ModeAvg
)

type StatsHandler struct {
	mode  StatsMode
	count int64
	sum   float64
}

func NewStatsHandler(mode StatsMode) *StatsHandler {
	return &StatsHandler{
		mode:  mode,
		count: 0,
		sum:   0.0,
	}
}

func (sh *StatsHandler) HandleMatch(value interface{}) {
	switch v := value.(type) {
	case []interface{}:
		for range v {
			sh.count++
		}
	case float64:
		sh.count++
		sh.sum += v
	case int:
		sh.count++
		sh.sum += float64(v)
	case int64:
		sh.count++
		sh.sum += float64(v)
	case string:
		sh.count++
	case bool:
		sh.count++
	case nil:
		sh.count++
	default:
		sh.count++
	}
}

func (sh *StatsHandler) Count() int64 {
	return sh.count
}

func (sh *StatsHandler) Sum() float64 {
	return sh.sum
}

func (sh *StatsHandler) Avg() float64 {
	if sh.count == 0 {
		return 0.0
	}
	return sh.sum / float64(sh.count)
}

func (sh *StatsHandler) Result() interface{} {
	switch sh.mode {
	case ModeCount:
		return sh.count
	case ModeSum:
		return sh.sum
	case ModeAvg:
		return sh.Avg()
	default:
		return nil
	}
}

func ParseStatsMode(mode string) (StatsMode, error) {
	switch mode {
	case "count":
		return ModeCount, nil
	case "sum":
		return ModeSum, nil
	case "avg":
		return ModeAvg, nil
	default:
		return ModeCount, fmt.Errorf("unknown stats mode: %s (use count, sum, or avg)", mode)
	}
}
