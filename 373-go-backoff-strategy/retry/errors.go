package retry

import (
	"errors"
	"fmt"
)

var (
	ErrNilFunction        = errors.New("function cannot be nil")
	ErrNilStrategy        = errors.New("strategy cannot be nil")
	ErrInvalidMaxRetries  = errors.New("max retries must be >= 0")
	ErrInvalidInitialWait = errors.New("initial wait duration must be > 0")
	ErrInvalidMaxWait     = errors.New("max wait duration must be >= 0")
	ErrInvalidMultiplier  = errors.New("multiplier must be > 0")
	ErrInvalidJitterFactor = errors.New("jitter factor must be between 0 and 1")
	ErrAllRetriesFailed   = errors.New("all retries failed")
)

type parameterError struct {
	param   string
	message string
	err     error
}

func (e *parameterError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("parameter error: %s - %s: %v", e.param, e.message, e.err)
	}
	return fmt.Sprintf("parameter error: %s - %s", e.param, e.message)
}

func (e *parameterError) Unwrap() error {
	return e.err
}

func newParameterError(param, message string, err error) error {
	return &parameterError{
		param:   param,
		message: message,
		err:     err,
	}
}

func IsParameterError(err error) bool {
	var pErr *parameterError
	return errors.As(err, &pErr)
}

type RetryError struct {
	Attempts int
	LastErr  error
}

func (e *RetryError) Error() string {
	if e.LastErr != nil {
		return fmt.Sprintf("retry failed after %d attempts: %v", e.Attempts, e.LastErr)
	}
	return fmt.Sprintf("retry failed after %d attempts", e.Attempts)
}

func (e *RetryError) Unwrap() error {
	return e.LastErr
}
