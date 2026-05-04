package retry

import (
	"math"
	"math/rand"
	"time"
)

type Strategy interface {
	NextDelay(attempt int) time.Duration
	GetMaxRetries() int
	Validate() error
}

type BaseStrategy struct {
	MaxRetries int
	MaxWait    time.Duration
}

func (b *BaseStrategy) GetMaxRetries() int {
	return b.MaxRetries
}

func (b *BaseStrategy) applyMaxWait(delay time.Duration) time.Duration {
	if b.MaxWait > 0 && delay > b.MaxWait {
		return b.MaxWait
	}
	return delay
}

type FixedIntervalStrategy struct {
	BaseStrategy
	InitialWait time.Duration
}

func (s *FixedIntervalStrategy) NextDelay(attempt int) time.Duration {
	return s.applyMaxWait(s.InitialWait)
}

func (s *FixedIntervalStrategy) Validate() error {
	if s.MaxRetries < 0 {
		return newParameterError("MaxRetries", "must be >= 0", ErrInvalidMaxRetries)
	}
	if s.InitialWait <= 0 {
		return newParameterError("InitialWait", "must be > 0", ErrInvalidInitialWait)
	}
	if s.MaxWait < 0 {
		return newParameterError("MaxWait", "must be >= 0", ErrInvalidMaxWait)
	}
	return nil
}

type LinearStrategy struct {
	BaseStrategy
	InitialWait time.Duration
	Increment   time.Duration
}

func (s *LinearStrategy) NextDelay(attempt int) time.Duration {
	delay := s.InitialWait + s.Increment*time.Duration(attempt-1)
	return s.applyMaxWait(delay)
}

func (s *LinearStrategy) Validate() error {
	if s.MaxRetries < 0 {
		return newParameterError("MaxRetries", "must be >= 0", ErrInvalidMaxRetries)
	}
	if s.InitialWait <= 0 {
		return newParameterError("InitialWait", "must be > 0", ErrInvalidInitialWait)
	}
	if s.Increment < 0 {
		return newParameterError("Increment", "must be >= 0", nil)
	}
	if s.MaxWait < 0 {
		return newParameterError("MaxWait", "must be >= 0", ErrInvalidMaxWait)
	}
	return nil
}

type ExponentialStrategy struct {
	BaseStrategy
	InitialWait time.Duration
	Multiplier  float64
}

func (s *ExponentialStrategy) NextDelay(attempt int) time.Duration {
	delay := float64(s.InitialWait) * math.Pow(s.Multiplier, float64(attempt-1))
	return s.applyMaxWait(time.Duration(delay))
}

func (s *ExponentialStrategy) Validate() error {
	if s.MaxRetries < 0 {
		return newParameterError("MaxRetries", "must be >= 0", ErrInvalidMaxRetries)
	}
	if s.InitialWait <= 0 {
		return newParameterError("InitialWait", "must be > 0", ErrInvalidInitialWait)
	}
	if s.Multiplier <= 0 {
		return newParameterError("Multiplier", "must be > 0", ErrInvalidMultiplier)
	}
	if s.MaxWait < 0 {
		return newParameterError("MaxWait", "must be >= 0", ErrInvalidMaxWait)
	}
	return nil
}

type ExponentialJitterStrategy struct {
	BaseStrategy
	InitialWait  time.Duration
	Multiplier   float64
	JitterFactor float64
	random       *rand.Rand
}

func (s *ExponentialJitterStrategy) NextDelay(attempt int) time.Duration {
	baseDelay := float64(s.InitialWait) * math.Pow(s.Multiplier, float64(attempt-1))
	jitterRange := baseDelay * s.JitterFactor
	minDelay := baseDelay - jitterRange
	maxDelay := baseDelay + jitterRange
	
	if s.random == nil {
		s.random = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	
	jitteredDelay := minDelay + s.random.Float64()*(maxDelay-minDelay)
	return s.applyMaxWait(time.Duration(jitteredDelay))
}

func (s *ExponentialJitterStrategy) Validate() error {
	if s.MaxRetries < 0 {
		return newParameterError("MaxRetries", "must be >= 0", ErrInvalidMaxRetries)
	}
	if s.InitialWait <= 0 {
		return newParameterError("InitialWait", "must be > 0", ErrInvalidInitialWait)
	}
	if s.Multiplier <= 0 {
		return newParameterError("Multiplier", "must be > 0", ErrInvalidMultiplier)
	}
	if s.JitterFactor < 0 || s.JitterFactor > 1 {
		return newParameterError("JitterFactor", "must be between 0 and 1", ErrInvalidJitterFactor)
	}
	if s.MaxWait < 0 {
		return newParameterError("MaxWait", "must be >= 0", ErrInvalidMaxWait)
	}
	return nil
}
