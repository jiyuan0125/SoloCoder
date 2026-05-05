package retry

import (
	"context"
	"time"
)

// RetryableFunc 定义可重试的函数类型
type RetryableFunc func(ctx context.Context) (interface{}, error)

// Retry 执行带重试机制的函数调用
// ctx: 上下文，可以用于取消操作
// fn: 需要执行的函数
// opts: 配置选项
func Retry(ctx context.Context, fn RetryableFunc, opts ...Option) (interface{}, error) {
	// 应用配置
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	return DoRetry(ctx, fn, config)
}

// DoRetry 使用指定配置执行重试
func DoRetry(ctx context.Context, fn RetryableFunc, config *Config) (interface{}, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 设置总超时
	var cancel context.CancelFunc
	if config.TotalTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, config.TotalTimeout)
		defer cancel()
	}

	startTime := time.Now()
	backoffCalculator := GetBackoffCalculator(config.BackoffStrategy)

	// 获取熔断器（如果配置了）
	var cb CircuitBreaker
	if config.CircuitBreaker != nil {
		cb = GetCircuitBreaker(config.CircuitBreaker.Name)
		// 确保熔断器配置一致
		if defaultCb, ok := cb.(*DefaultCircuitBreaker); ok {
			defaultCb.config = config.CircuitBreaker
		}
	}

	var lastErr error
	var result interface{}

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// 检查熔断器
		if cb != nil && !cb.Allow() {
			return nil, CircuitBreakerOpenError
		}

		// 构建重试上下文
		retryCtx := &RetryContext{
			Attempt:     attempt,
			MaxAttempts: config.MaxAttempts,
			StartTime:   startTime,
			ElapsedTime: time.Since(startTime),
			LastError:   lastErr,
			CustomData:  make(map[string]interface{}),
		}

		// 执行前置拦截器
		shouldContinue := true
		for _, interceptor := range config.Interceptors {
			if !interceptor.BeforeRetry(retryCtx) {
				shouldContinue = false
				break
			}
		}

		if !shouldContinue {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, nil
		}

		// 执行实际函数
		result, lastErr = fn(ctx)

		// 执行后置拦截器
		for i := len(config.Interceptors) - 1; i >= 0; i-- {
			config.Interceptors[i].AfterRetry(retryCtx, result, lastErr)
		}

		// 更新熔断器状态
		if cb != nil {
			if lastErr != nil {
				cb.Failure()
			} else {
				cb.Success()
			}
		}

		// 检查是否成功
		if lastErr == nil {
			return result, nil
		}

		// 检查是否是最后一次尝试
		if attempt >= config.MaxAttempts {
			break
		}

		// 检查是否应该重试
		shouldRetry := false
		for _, condition := range config.RetryConditions {
			if condition(lastErr, attempt) {
				shouldRetry = true
				break
			}
		}

		if !shouldRetry {
			break
		}

		// 计算下次重试的延迟
		nextDelay := backoffCalculator(attempt, config)
		retryCtx.NextDelay = nextDelay

		// 调用重试回调
		if config.OnRetry != nil {
			config.OnRetry(attempt, lastErr, nextDelay)
		}

		// 检查总超时
		if config.TotalTimeout > 0 {
			elapsed := time.Since(startTime)
			if elapsed+nextDelay > config.TotalTimeout {
				// 剩余时间不足以完成这次重试，直接失败
				break
			}
		}

		// 等待后重试
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(nextDelay):
			// 继续下一次重试
		}
	}

	return nil, lastErr
}

// RetryWithResult 执行带重试机制的函数调用，并返回指定类型的结果
// 这是一个泛型函数，需要Go 1.18+支持
func RetryWithResult[T any](ctx context.Context, fn func(ctx context.Context) (T, error), opts ...Option) (T, error) {
	var zero T

	// 包装成 RetryableFunc
	wrappedFn := func(ctx context.Context) (interface{}, error) {
		return fn(ctx)
	}

	result, err := Retry(ctx, wrappedFn, opts...)
	if err != nil {
		return zero, err
	}

	// 类型断言
	if typedResult, ok := result.(T); ok {
		return typedResult, nil
	}

	return zero, nil
}

// Do 是 Retry 的简化版本，使用默认配置
func Do(ctx context.Context, fn RetryableFunc) (interface{}, error) {
	return Retry(ctx, fn)
}

// Must 执行重试，如果失败则panic
func Must(ctx context.Context, fn RetryableFunc, opts ...Option) interface{} {
	result, err := Retry(ctx, fn, opts...)
	if err != nil {
		panic(err)
	}
	return result
}

// LoggingInterceptor 实现了一个简单的日志拦截器
type LoggingInterceptor struct {
	// Logger 日志函数，如果为nil则不输出日志
	Logger func(message string)
}

// BeforeRetry 实现 Interceptor 接口
func (i *LoggingInterceptor) BeforeRetry(ctx *RetryContext) bool {
	if i.Logger != nil {
		i.Logger(formatLogMessage(ctx, "before retry"))
	}
	return true
}

// AfterRetry 实现 Interceptor 接口
func (i *LoggingInterceptor) AfterRetry(ctx *RetryContext, result interface{}, err error) {
	if i.Logger != nil {
		if err != nil {
			i.Logger(formatLogMessage(ctx, "after retry - failed: "+err.Error()))
		} else {
			i.Logger(formatLogMessage(ctx, "after retry - succeeded"))
		}
	}
}

func formatLogMessage(ctx *RetryContext, message string) string {
	return "[Retry] Attempt " + string(rune(ctx.Attempt)) + "/" + string(rune(ctx.MaxAttempts)) + " - " + message
}

// MetricsInterceptor 实现了一个收集指标的拦截器
type MetricsInterceptor struct {
	// OnAttempt 每次尝试时调用
	OnAttempt func(ctx *RetryContext)
	// OnSuccess 成功时调用
	OnSuccess func(ctx *RetryContext)
	// OnFailure 失败时调用
	OnFailure func(ctx *RetryContext, err error)
}

// BeforeRetry 实现 Interceptor 接口
func (i *MetricsInterceptor) BeforeRetry(ctx *RetryContext) bool {
	if i.OnAttempt != nil {
		i.OnAttempt(ctx)
	}
	return true
}

// AfterRetry 实现 Interceptor 接口
func (i *MetricsInterceptor) AfterRetry(ctx *RetryContext, result interface{}, err error) {
	if err != nil {
		if i.OnFailure != nil {
			i.OnFailure(ctx, err)
		}
	} else {
		if i.OnSuccess != nil {
			i.OnSuccess(ctx)
		}
	}
}
