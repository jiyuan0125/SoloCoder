package retry

import (
	"errors"
	"sync"
	"time"
)

// CircuitBreaker 定义熔断器接口
type CircuitBreaker interface {
	// Allow 检查是否允许请求通过
	Allow() bool
	// Success 记录一次成功的请求
	Success()
	// Failure 记录一次失败的请求
	Failure()
	// State 获取当前状态
	State() CircuitBreakerState
	// Name 获取熔断器名称
	Name() string
	// Reset 重置熔断器状态
	Reset()
}

// DefaultCircuitBreaker 实现了一个默认的熔断器
type DefaultCircuitBreaker struct {
	// 配置
	config *CircuitBreakerConfig

	// 状态
	state CircuitBreakerState
	mu    sync.RWMutex

	// 统计
	failureCount     int
	successCount     int
	lastStateChange  time.Time
	halfOpenAttempts int

	// 回调
	onStateChange func(from, to CircuitBreakerState)
}

// 全局熔断器注册表
var (
	circuitBreakers     = make(map[string]CircuitBreaker)
	circuitBreakersLock sync.RWMutex
)

// GetCircuitBreaker 获取或创建一个熔断器
// 如果熔断器不存在，使用默认配置创建
func GetCircuitBreaker(name string) CircuitBreaker {
	circuitBreakersLock.RLock()
	cb, exists := circuitBreakers[name]
	circuitBreakersLock.RUnlock()

	if exists {
		return cb
	}

	circuitBreakersLock.Lock()
	defer circuitBreakersLock.Unlock()

	// 双重检查
	cb, exists = circuitBreakers[name]
	if exists {
		return cb
	}

	// 创建默认配置的熔断器
	cb = NewCircuitBreaker(&CircuitBreakerConfig{
		Name:                 name,
		FailureThreshold:     5,
		Timeout:              30 * time.Second,
		HalfOpenMaxAttempts:  3,
	})

	circuitBreakers[name] = cb
	return cb
}

// NewCircuitBreaker 创建一个新的熔断器
func NewCircuitBreaker(config *CircuitBreakerConfig) *DefaultCircuitBreaker {
	if config == nil {
		config = &CircuitBreakerConfig{
			FailureThreshold:     5,
			Timeout:              30 * time.Second,
			HalfOpenMaxAttempts:  3,
		}
	}

	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 5
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}
	if config.HalfOpenMaxAttempts <= 0 {
		config.HalfOpenMaxAttempts = 3
	}

	return &DefaultCircuitBreaker{
		config:          config,
		state:           StateClosed,
		lastStateChange: time.Now(),
	}
}

// Allow 检查是否允许请求通过
func (cb *DefaultCircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		// 关闭状态，允许请求
		return true

	case StateOpen:
		// 打开状态，检查是否超时需要进入半开状态
		if now.Sub(cb.lastStateChange) >= cb.config.Timeout {
			cb.transitionTo(StateHalfOpen)
			return true
		}
		return false

	case StateHalfOpen:
		// 半开状态，允许有限数量的请求
		if cb.halfOpenAttempts < cb.config.HalfOpenMaxAttempts {
			cb.halfOpenAttempts++
			return true
		}
		return false
	}

	return false
}

// Success 记录一次成功的请求
func (cb *DefaultCircuitBreaker) Success() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0

	switch cb.state {
	case StateClosed:
		// 关闭状态，重置成功计数
		cb.successCount++

	case StateHalfOpen:
		// 半开状态，成功次数增加
		cb.successCount++
		// 如果连续成功次数达到半开状态的最大尝试次数，恢复到关闭状态
		if cb.successCount >= cb.config.HalfOpenMaxAttempts {
			cb.transitionTo(StateClosed)
		}
	}
}

// Failure 记录一次失败的请求
func (cb *DefaultCircuitBreaker) Failure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successCount = 0

	switch cb.state {
	case StateClosed:
		// 关闭状态，增加失败计数
		cb.failureCount++
		// 检查是否超过失败阈值
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.transitionTo(StateOpen)
		}

	case StateHalfOpen:
		// 半开状态，只要有失败就重新打开
		cb.transitionTo(StateOpen)
	}
}

// transitionTo 切换状态
func (cb *DefaultCircuitBreaker) transitionTo(newState CircuitBreakerState) {
	if cb.state == newState {
		return
	}

	oldState := cb.state
	cb.state = newState
	cb.lastStateChange = time.Now()
	cb.failureCount = 0
	cb.successCount = 0
	cb.halfOpenAttempts = 0

	// 调用状态变化回调
	if cb.onStateChange != nil {
		cb.onStateChange(oldState, newState)
	}
}

// State 获取当前状态
func (cb *DefaultCircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Name 获取熔断器名称
func (cb *DefaultCircuitBreaker) Name() string {
	return cb.config.Name
}

// Reset 重置熔断器状态
func (cb *DefaultCircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.transitionTo(StateClosed)
}

// OnStateChange 设置状态变化回调
func (cb *DefaultCircuitBreaker) OnStateChange(callback func(from, to CircuitBreakerState)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onStateChange = callback
}

// Metrics 获取熔断器的统计指标
func (cb *DefaultCircuitBreaker) Metrics() (failureCount, successCount int, lastChange time.Time) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failureCount, cb.successCount, cb.lastStateChange
}

// CircuitBreakerOpenError 熔断器打开错误
var CircuitBreakerOpenError = errors.New("circuit breaker is open")

// ExecuteWithCircuitBreaker 使用熔断器执行函数
// 如果熔断器打开，直接返回错误
func ExecuteWithCircuitBreaker(cb CircuitBreaker, fn func() (interface{}, error)) (interface{}, error) {
	if !cb.Allow() {
		return nil, CircuitBreakerOpenError
	}

	result, err := fn()

	if err != nil {
		cb.Failure()
	} else {
		cb.Success()
	}

	return result, err
}

// String 状态字符串表示
func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}
