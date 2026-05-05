package retry

import (
	"math/rand"
	"time"
)

// BackoffCalculator 定义退避时间计算函数类型
type BackoffCalculator func(attempt int, config *Config) time.Duration

// GetBackoffCalculator 根据退避策略返回对应的时间计算函数
func GetBackoffCalculator(strategy BackoffStrategy) BackoffCalculator {
	switch strategy {
	case BackoffFixed:
		return FixedBackoff
	case BackoffLinear:
		return LinearBackoff
	case BackoffExponential:
		return ExponentialBackoff
	default:
		return ExponentialBackoff
	}
}

// FixedBackoff 固定间隔退避策略
// 每次重试的延迟时间固定为 InitialDelay
func FixedBackoff(attempt int, config *Config) time.Duration {
	delay := config.InitialDelay
	if config.EnableJitter {
		delay = applyJitter(delay, delay)
	}
	return minDuration(delay, config.MaxDelay)
}

// LinearBackoff 线性递增退避策略
// 延迟时间 = InitialDelay * attempt
func LinearBackoff(attempt int, config *Config) time.Duration {
	baseDelay := config.InitialDelay * time.Duration(attempt)
	if config.EnableJitter {
		baseDelay = applyJitter(baseDelay, config.InitialDelay)
	}
	return minDuration(baseDelay, config.MaxDelay)
}

// ExponentialBackoff 指数退避策略
// 延迟时间 = InitialDelay * 2^(attempt-1)
func ExponentialBackoff(attempt int, config *Config) time.Duration {
	baseDelay := config.InitialDelay
	for i := 1; i < attempt; i++ {
		baseDelay *= 2
	}
	if config.EnableJitter {
		baseDelay = applyJitter(baseDelay, config.InitialDelay)
	}
	return minDuration(baseDelay, config.MaxDelay)
}

// applyJitter 为延迟时间添加随机抖动
// baseDelay: 基础延迟时间
// maxJitter: 最大抖动范围（基于基础延迟的百分比，0.1表示10%）
func applyJitter(baseDelay time.Duration, initialDelay time.Duration) time.Duration {
	if baseDelay <= 0 {
		return 0
	}

	// 使用全抖动策略（Full Jitter）
	// 实际延迟 = random(0, baseDelay)
	// 这种策略可以有效防止惊群效应
	jitterRange := float64(baseDelay)
	jitter := time.Duration(rand.Float64() * jitterRange)

	// 确保最小延迟为初始延迟的10%
	minDelay := time.Duration(float64(initialDelay) * 0.1)
	return maxDuration(jitter, minDelay)
}

// EqualJitter 等抖动策略
// 延迟时间 = baseDelay/2 + random(0, baseDelay/2)
// 这种策略保证延迟至少为基础延迟的一半
func EqualJitter(baseDelay time.Duration) time.Duration {
	if baseDelay <= 0 {
		return 0
	}
	halfDelay := baseDelay / 2
	jitter := time.Duration(rand.Float64() * float64(halfDelay))
	return halfDelay + jitter
}

// DecorrelatedJitter 去相关抖动策略
// 延迟时间 = min(maxDelay, random(initialDelay, previousDelay*3))
// 这种策略可以使不同客户端的重试时间更加分散
func DecorrelatedJitter(previousDelay time.Duration, initialDelay time.Duration, maxDelay time.Duration) time.Duration {
	if previousDelay <= 0 {
		previousDelay = initialDelay
	}
	minRange := float64(initialDelay)
	maxRange := float64(previousDelay * 3)
	if maxRange > float64(maxDelay) {
		maxRange = float64(maxDelay)
	}
	if maxRange < minRange {
		maxRange = minRange
	}
	jitter := time.Duration(minRange + rand.Float64()*(maxRange-minRange))
	return minDuration(jitter, maxDelay)
}

// minDuration 返回两个时间中的较小值
func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// maxDuration 返回两个时间中的较大值
func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
