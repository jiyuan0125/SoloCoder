package retry

import "time"

type FixedIntervalBuilder struct {
	strategy *FixedIntervalStrategy
}

func NewFixedIntervalBuilder() *FixedIntervalBuilder {
	return &FixedIntervalBuilder{
		strategy: &FixedIntervalStrategy{
			BaseStrategy: BaseStrategy{
				MaxRetries: 3,
				MaxWait:    0,
			},
			InitialWait: 1 * time.Second,
		},
	}
}

func (b *FixedIntervalBuilder) WithMaxRetries(max int) *FixedIntervalBuilder {
	b.strategy.MaxRetries = max
	return b
}

func (b *FixedIntervalBuilder) WithMaxWait(wait time.Duration) *FixedIntervalBuilder {
	b.strategy.MaxWait = wait
	return b
}

func (b *FixedIntervalBuilder) WithInitialWait(wait time.Duration) *FixedIntervalBuilder {
	b.strategy.InitialWait = wait
	return b
}

func (b *FixedIntervalBuilder) Build() Strategy {
	return b.strategy
}

type LinearBuilder struct {
	strategy *LinearStrategy
}

func NewLinearBuilder() *LinearBuilder {
	return &LinearBuilder{
		strategy: &LinearStrategy{
			BaseStrategy: BaseStrategy{
				MaxRetries: 3,
				MaxWait:    0,
			},
			InitialWait: 1 * time.Second,
			Increment:   1 * time.Second,
		},
	}
}

func (b *LinearBuilder) WithMaxRetries(max int) *LinearBuilder {
	b.strategy.MaxRetries = max
	return b
}

func (b *LinearBuilder) WithMaxWait(wait time.Duration) *LinearBuilder {
	b.strategy.MaxWait = wait
	return b
}

func (b *LinearBuilder) WithInitialWait(wait time.Duration) *LinearBuilder {
	b.strategy.InitialWait = wait
	return b
}

func (b *LinearBuilder) WithIncrement(increment time.Duration) *LinearBuilder {
	b.strategy.Increment = increment
	return b
}

func (b *LinearBuilder) Build() Strategy {
	return b.strategy
}

type ExponentialBuilder struct {
	strategy *ExponentialStrategy
}

func NewExponentialBuilder() *ExponentialBuilder {
	return &ExponentialBuilder{
		strategy: &ExponentialStrategy{
			BaseStrategy: BaseStrategy{
				MaxRetries: 3,
				MaxWait:    0,
			},
			InitialWait: 1 * time.Second,
			Multiplier:  2.0,
		},
	}
}

func (b *ExponentialBuilder) WithMaxRetries(max int) *ExponentialBuilder {
	b.strategy.MaxRetries = max
	return b
}

func (b *ExponentialBuilder) WithMaxWait(wait time.Duration) *ExponentialBuilder {
	b.strategy.MaxWait = wait
	return b
}

func (b *ExponentialBuilder) WithInitialWait(wait time.Duration) *ExponentialBuilder {
	b.strategy.InitialWait = wait
	return b
}

func (b *ExponentialBuilder) WithMultiplier(multiplier float64) *ExponentialBuilder {
	b.strategy.Multiplier = multiplier
	return b
}

func (b *ExponentialBuilder) Build() Strategy {
	return b.strategy
}

type ExponentialJitterBuilder struct {
	strategy *ExponentialJitterStrategy
}

func NewExponentialJitterBuilder() *ExponentialJitterBuilder {
	return &ExponentialJitterBuilder{
		strategy: &ExponentialJitterStrategy{
			BaseStrategy: BaseStrategy{
				MaxRetries: 3,
				MaxWait:    0,
			},
			InitialWait:  1 * time.Second,
			Multiplier:   2.0,
			JitterFactor: 0.5,
		},
	}
}

func (b *ExponentialJitterBuilder) WithMaxRetries(max int) *ExponentialJitterBuilder {
	b.strategy.MaxRetries = max
	return b
}

func (b *ExponentialJitterBuilder) WithMaxWait(wait time.Duration) *ExponentialJitterBuilder {
	b.strategy.MaxWait = wait
	return b
}

func (b *ExponentialJitterBuilder) WithInitialWait(wait time.Duration) *ExponentialJitterBuilder {
	b.strategy.InitialWait = wait
	return b
}

func (b *ExponentialJitterBuilder) WithMultiplier(multiplier float64) *ExponentialJitterBuilder {
	b.strategy.Multiplier = multiplier
	return b
}

func (b *ExponentialJitterBuilder) WithJitterFactor(factor float64) *ExponentialJitterBuilder {
	b.strategy.JitterFactor = factor
	return b
}

func (b *ExponentialJitterBuilder) Build() Strategy {
	return b.strategy
}
