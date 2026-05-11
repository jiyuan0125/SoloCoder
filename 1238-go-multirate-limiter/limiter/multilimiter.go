package limiter

import (
	"multirate-limiter/common"
)

type MultiLimiter struct {
	ruleLimiters []*ruleLimiter
}

func NewMultiLimiter(rules []common.Rule) *MultiLimiter {
	ml := &MultiLimiter{
		ruleLimiters: make([]*ruleLimiter, 0, len(rules)),
	}
	for _, rule := range rules {
		ml.ruleLimiters = append(ml.ruleLimiters, newRuleLimiter(rule))
	}
	return ml
}

func (m *MultiLimiter) Allow() (allowed bool, results []common.RuleResult) {
	results = make([]common.RuleResult, 0, len(m.ruleLimiters))
	allowed = true

	for _, rl := range m.ruleLimiters {
		ok, current := rl.allow(false)
		result := common.RuleResult{
			Granularity: rl.getRule().Granularity,
			Quota:       rl.getRule().Quota,
			Current:     current + 1,
			Allowed:     ok,
		}
		if !ok {
			result.Current = current
			allowed = false
		}
		results = append(results, result)
	}

	return allowed, results
}

func (m *MultiLimiter) Stats() []common.RuleResult {
	results := make([]common.RuleResult, 0, len(m.ruleLimiters))
	for _, rl := range m.ruleLimiters {
		current, quota := rl.stats()
		results = append(results, common.RuleResult{
			Granularity: rl.getRule().Granularity,
			Quota:       quota,
			Current:     current,
			Allowed:     current < quota,
		})
	}
	return results
}

func (m *MultiLimiter) Rules() []common.Rule {
	rules := make([]common.Rule, 0, len(m.ruleLimiters))
	for _, rl := range m.ruleLimiters {
		rules = append(rules, rl.getRule())
	}
	return rules
}
