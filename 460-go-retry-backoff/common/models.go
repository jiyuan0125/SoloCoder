// Package common 定义了服务端和客户端共享的请求响应结构体
package common

// Request 定义了客户端发送给服务端的请求结构
type Request struct {
	// ID 请求的唯一标识符
	ID string `json:"id"`
	// Message 请求的消息内容
	Message string `json:"message"`
	// Metadata 附加的元数据
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Response 定义了服务端返回给客户端的响应结构
type Response struct {
	// ID 与请求对应的唯一标识符
	ID string `json:"id"`
	// Success 表示请求是否成功处理
	Success bool `json:"success"`
	// Message 响应的消息内容
	Message string `json:"message"`
	// Data 响应的附加数据
	Data map[string]interface{} `json:"data,omitempty"`
	// Error 错误信息，当Success为false时有效
	Error string `json:"error,omitempty"`
}

// RetryConfig 定义了重试相关的配置选项
type RetryConfig struct {
	// MaxAttempts 最大重试次数（包括首次调用）
	MaxAttempts int `json:"max_attempts"`
	// InitialDelay 初始重试延迟
	InitialDelay string `json:"initial_delay"`
	// MaxDelay 最大重试延迟
	MaxDelay string `json:"max_delay"`
	// BackoffStrategy 退避策略：fixed, linear, exponential
	BackoffStrategy string `json:"backoff_strategy"`
	// EnableJitter 是否启用抖动
	EnableJitter bool `json:"enable_jitter"`
	// TotalTimeout 总超时时间
	TotalTimeout string `json:"total_timeout"`
	// CircuitBreakerThreshold 熔断器连续失败阈值
	CircuitBreakerThreshold int `json:"circuit_breaker_threshold"`
	// CircuitBreakerTimeout 熔断器超时时间
	CircuitBreakerTimeout string `json:"circuit_breaker_timeout"`
}
