package circuitbreaker

import "time"

type State string

const (
	StateClosed   State = "closed"
	StateOpen     State = "open"
	StateHalfOpen State = "half-open"
)

type StateChangeEvent struct {
	Name      string
	From      State
	To        State
	Timestamp time.Time
}

type StateChangeCallback func(event StateChangeEvent)

type RequestResult bool

const (
	ResultSuccess RequestResult = true
	ResultFailure RequestResult = false
)

type Statistics struct {
	TotalRequests  int64
	SuccessCount   int64
	FailureCount   int64
	RecentFailures int
	RecentSuccesses int
}

type CircuitBreaker interface {
	Name() string
	State() State
	Allow() (bool, error)
	MarkSuccess()
	MarkFailure()
	Execute(func() error) error
	Reset()
	ForceState(state State)
	GetStatistics() Statistics
	OnStateChange(callback StateChangeCallback)
	SaveToFile(path string) error
	LoadFromFile(path string) error
	GetConfig() Config
	GetOpenAt() time.Time
	GetLastStateChange() time.Time
}

var ErrCircuitOpen = &CircuitOpenError{}

type CircuitOpenError struct {
	CircuitName string
	OpenSince   time.Time
}

func (e *CircuitOpenError) Error() string {
	return "circuit breaker is open for " + e.CircuitName
}

type WindowRecord struct {
	Result    RequestResult
	Timestamp time.Time
}
