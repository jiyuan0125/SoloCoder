package retry

import (
	"errors"
	"net"
	"net/http"
	"os"
	"syscall"
)

// IsNetworkError 检查错误是否是网络相关的错误
// 包括：网络超时、连接拒绝、连接重置、DNS解析错误等
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否是超时错误
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
	}

	// 检查具体的系统调用错误
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		switch opErr.Op {
		case "dial", "read", "write", "connect":
			// 检查底层错误
			var sysErr syscall.Errno
			if errors.As(opErr.Err, &sysErr) {
				switch sysErr {
				case syscall.ECONNREFUSED, syscall.ECONNRESET, syscall.ETIMEDOUT,
					syscall.ENETUNREACH, syscall.EHOSTUNREACH, syscall.EPIPE:
					return true
				}
			}

			// 检查DNS解析错误
			var dnsErr *net.DNSError
			if errors.As(opErr.Err, &dnsErr) {
				return true
			}
		}
	}

	// 检查HTTP客户端错误
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		// HTTP错误在DefaultRetryCondition中单独处理
		return false
	}

	// 检查是否是连接池相关错误
	// 某些库会使用这些错误类型
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}

	// 检查连接关闭错误
	if errors.Is(err, net.ErrClosed) {
		return true
	}

	return false
}

// IsHTTPError 检查错误是否是HTTP错误
func IsHTTPError(err error) bool {
	if err == nil {
		return false
	}
	var httpErr *HTTPError
	return errors.As(err, &httpErr)
}

// IsHTTPServerError 检查是否是HTTP 5xx服务器错误
func IsHTTPServerError(err error) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode >= 500 && httpErr.StatusCode < 600
	}
	return false
}

// IsHTTPTooManyRequests 检查是否是HTTP 429请求过多错误
func IsHTTPTooManyRequests(err error) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == http.StatusTooManyRequests
	}
	return false
}

// IsHTTPBadGateway 检查是否是HTTP 502 Bad Gateway错误
func IsHTTPBadGateway(err error) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == http.StatusBadGateway
	}
	return false
}

// IsHTTPServiceUnavailable 检查是否是HTTP 503 Service Unavailable错误
func IsHTTPServiceUnavailable(err error) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == http.StatusServiceUnavailable
	}
	return false
}

// IsHTTPGatewayTimeout 检查是否是HTTP 504 Gateway Timeout错误
func IsHTTPGatewayTimeout(err error) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == http.StatusGatewayTimeout
	}
	return false
}

// IsConnectionRefused 检查是否是连接拒绝错误
func IsConnectionRefused(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		var sysErr syscall.Errno
		if errors.As(opErr.Err, &sysErr) {
			return sysErr == syscall.ECONNREFUSED
		}
	}
	return false
}

// IsConnectionReset 检查是否是连接重置错误
func IsConnectionReset(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		var sysErr syscall.Errno
		if errors.As(opErr.Err, &sysErr) {
			return sysErr == syscall.ECONNRESET
		}
	}
	return false
}

// IsTimeout 检查是否是超时错误
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}

	return false
}

// RetryableError 定义可重试错误接口
// 实现此接口的错误类型可以明确指示是否应该重试
type RetryableError interface {
	error
	IsRetryable() bool
}

// NewRetryableError 创建一个明确标记为可重试的错误
func NewRetryableError(err error) error {
	if err == nil {
		return nil
	}
	return &retryableError{
		err:       err,
		retryable: true,
	}
}

// NewNonRetryableError 创建一个明确标记为不可重试的错误
func NewNonRetryableError(err error) error {
	if err == nil {
		return nil
	}
	return &retryableError{
		err:       err,
		retryable: false,
	}
}

type retryableError struct {
	err       error
	retryable bool
}

func (e *retryableError) Error() string {
	return e.err.Error()
}

func (e *retryableError) Unwrap() error {
	return e.err
}

func (e *retryableError) IsRetryable() bool {
	return e.retryable
}

// IsRetryable 检查错误是否是可重试的
// 优先检查是否实现了RetryableError接口
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否实现了RetryableError接口
	var retryableErr RetryableError
	if errors.As(err, &retryableErr) {
		return retryableErr.IsRetryable()
	}

	// 默认使用网络错误判断
	return IsNetworkError(err)
}

// WithCustomRetryCondition 创建一个自定义重试条件
// 允许调用方根据错误类型和尝试次数决定是否重试
func WithCustomRetryCondition(condition func(err error, attempt int) bool) RetryCondition {
	return func(err error, attempt int) bool {
		if err == nil {
			return false
		}
		return condition(err, attempt)
	}
}

// CombineRetryConditions 组合多个重试条件
// 只要有一个条件返回true，就认为应该重试
func CombineRetryConditions(conditions ...RetryCondition) RetryCondition {
	return func(err error, attempt int) bool {
		for _, condition := range conditions {
			if condition(err, attempt) {
				return true
			}
		}
		return false
	}
}

// RequireAllRetryConditions 要求所有条件都满足才重试
// 只有所有条件都返回true，才认为应该重试
func RequireAllRetryConditions(conditions ...RetryCondition) RetryCondition {
	return func(err error, attempt int) bool {
		for _, condition := range conditions {
			if !condition(err, attempt) {
				return false
			}
		}
		return true
	}
}
