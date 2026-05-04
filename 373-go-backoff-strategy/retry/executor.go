package retry

import (
	"errors"
	"time"
)

type RetryFunc func() error

type OnRetryFunc func(attempt int, delay time.Duration, err error)

type NonRetryableErrorChecker func(err error) bool

type Executor struct {
	strategy              Strategy
	onRetry               OnRetryFunc
	nonRetryableCheckers  []NonRetryableErrorChecker
	sleeper               func(d time.Duration)
}

func NewExecutor(strategy Strategy) (*Executor, error) {
	if strategy == nil {
		return nil, newParameterError("strategy", "cannot be nil", ErrNilStrategy)
	}
	if err := strategy.Validate(); err != nil {
		return nil, err
	}
	return &Executor{
		strategy:             strategy,
		nonRetryableCheckers: make([]NonRetryableErrorChecker, 0),
		sleeper:              time.Sleep,
	}, nil
}

func (e *Executor) WithOnRetry(callback OnRetryFunc) *Executor {
	e.onRetry = callback
	return e
}

func (e *Executor) WithNonRetryableError(checker NonRetryableErrorChecker) *Executor {
	if checker != nil {
		e.nonRetryableCheckers = append(e.nonRetryableCheckers, checker)
	}
	return e
}

func (e *Executor) Execute(fn RetryFunc) error {
	if fn == nil {
		return newParameterError("fn", "cannot be nil", ErrNilFunction)
	}

	maxRetries := e.strategy.GetMaxRetries()
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if e.isNonRetryableError(err) {
			return &RetryError{
				Attempts: attempt + 1,
				LastErr:  err,
			}
		}

		if attempt < maxRetries {
			delay := e.strategy.NextDelay(attempt + 1)
			if e.onRetry != nil {
				e.onRetry(attempt+1, delay, err)
			}
			e.sleeper(delay)
		}
	}

	return &RetryError{
		Attempts: maxRetries + 1,
		LastErr:  lastErr,
	}
}

func (e *Executor) isNonRetryableError(err error) bool {
	for _, checker := range e.nonRetryableCheckers {
		if checker(err) {
			return true
		}
	}
	return false
}

func IsNonRetryableError(err error, checker NonRetryableErrorChecker) bool {
	if err == nil || checker == nil {
		return false
	}
	return checker(err)
}

func IsRetryError(err error) bool {
	var rErr *RetryError
	return errors.As(err, &rErr)
}

func GetRetryAttempts(err error) int {
	var rErr *RetryError
	if errors.As(err, &rErr) {
		return rErr.Attempts
	}
	return 0
}
