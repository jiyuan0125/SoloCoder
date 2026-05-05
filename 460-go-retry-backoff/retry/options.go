// Package retry 提供了灵活的重试机制，用于处理分布式系统中的临时故障
package retry

import (
	"errors"
	"net/http"
	"time"
)

// Config 定义重试的配置选项
type Config struct {
	// MaxAttempts 最大重试次数（包括首次尝试）
	MaxAttempts int
	// InitialDelay 初始重试延迟
	InitialDelay time.Duration
	// MaxDelay 最大重试延迟
	MaxDelay time.Duration
	// BackoffStrategy 退避策略
	BackoffStrategy BackoffStrategy
	// EnableJitter 是否启用抖动
	EnableJitter bool
	// TotalTimeout 整个重试过程的总超时
	TotalTimeout time.Duration
	// RetryConditions 重试条件判断函数列表
	RetryConditions []RetryCondition
	// OnRetry 重试回调函数
	OnRetry OnRetryCallback
	// Interceptors 重试拦截器列表
	Interceptors []Interceptor
	// CircuitBreaker 熔断器配置
	CircuitBreaker *CircuitBreakerConfig
}

// BackoffStrategy 定义退避策略类型
type BackoffStrategy int

const (
	// BackoffFixed 固定间隔退避策略
	BackoffFixed BackoffStrategy = iota
	// BackoffLinear 线性递增退避策略
	BackoffLinear
	// BackoffExponential 指数退避策略
	BackoffExponential
)

// RetryCondition 定义重试条件判断函数类型
// 返回true表示应该重试
type RetryCondition func(err error, attempt int) bool

// OnRetryCallback 定义重试回调函数类型
// attempt: 第几次重试（从1开始）
// lastErr: 上次尝试的错误
// nextDelay: 下次重试的等待间隔
type OnRetryCallback func(attempt int, lastErr error, nextDelay time.Duration)

// Interceptor 定义重试拦截器接口
// 可以在每次重试前后执行自定义逻辑
type Interceptor interface {
	// BeforeRetry 在重试之前执行
	// 返回true表示继续执行，返回false表示取消重试
	BeforeRetry(ctx *RetryContext) bool
	// AfterRetry 在重试之后执行
	AfterRetry(ctx *RetryContext, result interface{}, err error)
}

// RetryContext 定义重试上下文，包含当前重试的相关信息
type RetryContext struct {
	// Attempt 当前尝试次数（从1开始）
	Attempt int
	// MaxAttempts 最大尝试次数
	MaxAttempts int
	// StartTime 开始时间
	StartTime time.Time
	// ElapsedTime 已流逝时间
	ElapsedTime time.Duration
	// LastError 上次错误
	LastError error
	// NextDelay 下次重试的延迟
	NextDelay time.Duration
	// RequestInfo 请求相关信息（可选）
	RequestInfo *RequestInfo
	// CustomData 自定义数据，拦截器可以使用
	CustomData map[string]interface{}
}

// RequestInfo 定义HTTP请求相关信息
type RequestInfo struct {
	// Method HTTP方法
	Method string
	// URL 请求URL
	URL string
	// StatusCode HTTP状态码（如果有响应）
	StatusCode int
	// Response HTTP响应（如果有）
	Response *http.Response
}

// CircuitBreakerConfig 定义熔断器配置
type CircuitBreakerConfig struct {
	// Name 熔断器名称，用于标识不同的服务
	Name string
	// FailureThreshold 连续失败阈值，超过后熔断器打开
	FailureThreshold int
	// Timeout 熔断器打开后的超时时间，超时后进入半开状态
	Timeout time.Duration
	// HalfOpenMaxAttempts 半开状态下的最大尝试次数
	HalfOpenMaxAttempts int
}

// CircuitBreakerState 定义熔断器状态
type CircuitBreakerState int

const (
	// StateClosed 熔断器关闭状态，允许请求通过
	StateClosed CircuitBreakerState = iota
	// StateOpen 熔断器打开状态，拒绝请求
	StateOpen
	// StateHalfOpen 熔断器半开状态，允许少量请求通过进行探测
	StateHalfOpen
)

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxAttempts:      3,
		InitialDelay:     100 * time.Millisecond,
		MaxDelay:         5 * time.Second,
		BackoffStrategy:  BackoffExponential,
		EnableJitter:     true,
		TotalTimeout:     30 * time.Second,
		RetryConditions:  []RetryCondition{DefaultRetryCondition},
		CircuitBreaker:   nil,
	}
}

// Option 定义配置选项函数类型
type Option func(*Config)

// WithMaxAttempts 设置最大重试次数
func WithMaxAttempts(maxAttempts int) Option {
	return func(c *Config) {
		if maxAttempts > 0 {
			c.MaxAttempts = maxAttempts
		}
	}
}

// WithInitialDelay 设置初始重试延迟
func WithInitialDelay(delay time.Duration) Option {
	return func(c *Config) {
		if delay > 0 {
			c.InitialDelay = delay
		}
	}
}

// WithMaxDelay 设置最大重试延迟
func WithMaxDelay(delay time.Duration) Option {
	return func(c *Config) {
		if delay > 0 {
			c.MaxDelay = delay
		}
	}
}

// WithBackoffStrategy 设置退避策略
func WithBackoffStrategy(strategy BackoffStrategy) Option {
	return func(c *Config) {
		c.BackoffStrategy = strategy
	}
}

// WithJitter 启用或禁用抖动
func WithJitter(enable bool) Option {
	return func(c *Config) {
		c.EnableJitter = enable
	}
}

// WithTotalTimeout 设置总超时
func WithTotalTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		if timeout > 0 {
			c.TotalTimeout = timeout
		}
	}
}

// WithRetryConditions 设置重试条件
func WithRetryConditions(conditions ...RetryCondition) Option {
	return func(c *Config) {
		c.RetryConditions = conditions
	}
}

// WithOnRetryCallback 设置重试回调
func WithOnRetryCallback(callback OnRetryCallback) Option {
	return func(c *Config) {
		c.OnRetry = callback
	}
}

// WithInterceptors 设置拦截器
func WithInterceptors(interceptors ...Interceptor) Option {
	return func(c *Config) {
		c.Interceptors = interceptors
	}
}

// WithCircuitBreaker 设置熔断器配置
func WithCircuitBreaker(config *CircuitBreakerConfig) Option {
	return func(c *Config) {
		c.CircuitBreaker = config
	}
}

// DefaultRetryCondition 默认重试条件
// 只重试网络错误，不重试业务错误
func DefaultRetryCondition(err error, attempt int) bool {
	if err == nil {
		return false
	}

	// 检查是否是HTTP错误
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		// 只重试5xx和429状态码
		return httpErr.StatusCode >= 500 || httpErr.StatusCode == 429
	}

	// 检查是否是网络错误
	return IsNetworkError(err)
}

// HTTPError 定义HTTP错误
type HTTPError struct {
	StatusCode int
	Message    string
}

// Error 实现error接口
func (e *HTTPError) Error() string {
	return e.Message
}
