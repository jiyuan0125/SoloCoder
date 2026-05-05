// Package main 实现了一个演示用的HTTP客户端
// 用于演示重试库的各种功能
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/example/retrybackoff/common"
	"github.com/example/retrybackoff/retry"
)

// Client 定义客户端结构
type Client struct {
	// 重试HTTP客户端
	httpClient *retry.HTTPClient
	// 服务端地址
	serverURL string
}

// NewClient 创建一个新的客户端
func NewClient(serverURL string, maxAttempts int, initialDelay time.Duration, strategy retry.BackoffStrategy, enableJitter bool) *Client {
	// 创建带重试的HTTP客户端
	httpClient := retry.NewHTTPClient(
		serverURL,
		retry.WithMaxAttempts(maxAttempts),
		retry.WithInitialDelay(initialDelay),
		retry.WithMaxDelay(5*time.Second),
		retry.WithBackoffStrategy(strategy),
		retry.WithJitter(enableJitter),
		retry.WithTotalTimeout(30*time.Second),
		// 添加重试回调
		retry.WithOnRetryCallback(func(attempt int, lastErr error, nextDelay time.Duration) {
			log.Printf("[Retry Callback] Attempt %d failed: %v, next retry in %v", attempt, lastErr, nextDelay)
		}),
	)

	return &Client{
		httpClient: httpClient,
		serverURL:  serverURL,
	}
}

// DemoHealthCheck 演示健康检查
func (c *Client) DemoHealthCheck(ctx context.Context) error {
	log.Println("\n=== Demo: Health Check ===")

	var response common.Response
	err := c.httpClient.GetWithResponse(ctx, "/api/health", &response)
	if err != nil {
		log.Printf("Health check failed: %v", err)
		return err
	}

	log.Printf("Health check successful:")
	log.Printf("  ID: %s", response.ID)
	log.Printf("  Success: %v", response.Success)
	log.Printf("  Message: %s", response.Message)

	return nil
}

// DemoEcho 演示回显接口
func (c *Client) DemoEcho(ctx context.Context) error {
	log.Println("\n=== Demo: Echo ===")

	request := common.Request{
		ID:      "echo-" + time.Now().Format("20060102150405"),
		Message: "Hello, Retry Library!",
	}

	var response common.Response
	err := c.httpClient.PostWithResponse(ctx, "/api/echo", request, &response)
	if err != nil {
		log.Printf("Echo failed: %v", err)
		return err
	}

	log.Printf("Echo successful:")
	log.Printf("  ID: %s", response.ID)
	log.Printf("  Success: %v", response.Success)
	log.Printf("  Message: %s", response.Message)

	return nil
}

// DemoFlakyService 演示调用不稳定服务
func (c *Client) DemoFlakyService(ctx context.Context, numRequests int) error {
	log.Println("\n=== Demo: Flaky Service ===")
	log.Printf("Sending %d requests to flaky service...", numRequests)

	successCount := 0
	failureCount := 0

	for i := 0; i < numRequests; i++ {
		request := common.Request{
			ID:      fmt.Sprintf("flaky-%d-%s", i+1, time.Now().Format("20060102150405")),
			Message: fmt.Sprintf("Test message %d", i+1),
		}

		var response common.Response
		err := c.httpClient.PostWithResponse(ctx, "/api/flaky", request, &response)
		if err != nil {
			log.Printf("Request %d failed after retries: %v", i+1, err)
			failureCount++
		} else {
			log.Printf("Request %d succeeded", i+1)
			successCount++
		}

		// 稍微延迟一下
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("\nResults:")
	log.Printf("  Successful: %d", successCount)
	log.Printf("  Failed: %d", failureCount)
	log.Printf("  Success rate: %.2f%%", float64(successCount)/float64(numRequests)*100)

	return nil
}

// DemoWithInterceptor 演示使用拦截器
func (c *Client) DemoWithInterceptor(ctx context.Context) error {
	log.Println("\n=== Demo: Interceptors ===")

	// 创建指标拦截器
	attemptCount := 0
	successCount := 0
	failureCount := 0

	metricsInterceptor := &retry.MetricsInterceptor{
		OnAttempt: func(ctx *retry.RetryContext) {
			attemptCount++
			log.Printf("[Metrics] Attempt %d/%d", ctx.Attempt, ctx.MaxAttempts)
		},
		OnSuccess: func(ctx *retry.RetryContext) {
			successCount++
			log.Printf("[Metrics] Attempt %d succeeded after %d attempts", ctx.Attempt, ctx.Attempt)
		},
		OnFailure: func(ctx *retry.RetryContext, err error) {
			failureCount++
			log.Printf("[Metrics] Attempt %d failed: %v", ctx.Attempt, err)
		},
	}

	// 创建带拦截器的客户端
	clientWithInterceptor := retry.NewHTTPClient(
		c.serverURL,
		retry.WithMaxAttempts(5),
		retry.WithInitialDelay(100*time.Millisecond),
		retry.WithBackoffStrategy(retry.BackoffExponential),
		retry.WithInterceptors(metricsInterceptor),
	)

	request := common.Request{
		ID:      "interceptor-" + time.Now().Format("20060102150405"),
		Message: "Testing interceptors",
	}

	var response common.Response
	err := clientWithInterceptor.PostWithResponse(ctx, "/api/flaky", request, &response)
	if err != nil {
		log.Printf("Request failed: %v", err)
	} else {
		log.Printf("Request succeeded")
	}

	log.Printf("\nInterceptor Metrics:")
	log.Printf("  Total attempts: %d", attemptCount)
	log.Printf("  Success count: %d", successCount)
	log.Printf("  Failure count: %d", failureCount)

	return nil
}

// DemoCircuitBreaker 演示熔断器模式
func (c *Client) DemoCircuitBreaker(ctx context.Context) error {
	log.Println("\n=== Demo: Circuit Breaker ===")

	// 创建带熔断器的客户端
	cbConfig := &retry.CircuitBreakerConfig{
		Name:                 "demo-service",
		FailureThreshold:     3,
		Timeout:              10 * time.Second,
		HalfOpenMaxAttempts:  2,
	}

	clientWithCB := retry.NewHTTPClient(
		c.serverURL,
		retry.WithMaxAttempts(1), // 禁用重试，让熔断器快速触发
		retry.WithCircuitBreaker(cbConfig),
	)

	// 获取熔断器实例以便监控
	cb := retry.GetCircuitBreaker("demo-service")

	// 发送多个请求触发熔断器
	log.Println("Sending requests to trigger circuit breaker...")

	for i := 0; i < 10; i++ {
		request := common.Request{
			ID:      fmt.Sprintf("cb-%d-%s", i+1, time.Now().Format("20060102150405")),
			Message: fmt.Sprintf("Circuit breaker test %d", i+1),
		}

		var response common.Response
		err := clientWithCB.PostWithResponse(ctx, "/api/flaky", request, &response)

		state := cb.State()
		if err != nil {
			log.Printf("Request %d: State=%s, Error=%v", i+1, state, err)
		} else {
			log.Printf("Request %d: State=%s, Success", i+1, state)
		}

		// 如果熔断器打开，等待一段时间让它进入半开状态
		if state == retry.StateOpen {
			log.Println("Circuit breaker is OPEN, waiting 5 seconds for half-open state...")
			time.Sleep(5 * time.Second)
		} else {
			time.Sleep(200 * time.Millisecond)
		}
	}

	return nil
}

// DemoConcurrentRequests 演示并发请求
func (c *Client) DemoConcurrentRequests(ctx context.Context, numGoroutines int, numRequestsPerGoroutine int) error {
	log.Println("\n=== Demo: Concurrent Requests ===")
	log.Printf("Starting %d goroutines, each sending %d requests...", numGoroutines, numRequestsPerGoroutine)

	var wg sync.WaitGroup
	successChan := make(chan int, numGoroutines*numRequestsPerGoroutine)
	failureChan := make(chan int, numGoroutines*numRequestsPerGoroutine)

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for i := 0; i < numRequestsPerGoroutine; i++ {
				request := common.Request{
					ID:      fmt.Sprintf("concurrent-%d-%d-%s", goroutineID, i, time.Now().Format("20060102150405")),
					Message: fmt.Sprintf("Goroutine %d, Request %d", goroutineID, i),
				}

				var response common.Response
				err := c.httpClient.PostWithResponse(ctx, "/api/flaky", request, &response)
				if err != nil {
					failureChan <- 1
				} else {
					successChan <- 1
				}
			}
		}(g)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(successChan)
	close(failureChan)

	// 统计结果
	successCount := 0
	failureCount := 0

	for range successChan {
		successCount++
	}
	for range failureChan {
		failureCount++
	}

	total := successCount + failureCount
	log.Printf("\nConcurrent Results:")
	log.Printf("  Total requests: %d", total)
	log.Printf("  Successful: %d", successCount)
	log.Printf("  Failed: %d", failureCount)
	log.Printf("  Success rate: %.2f%%", float64(successCount)/float64(total)*100)

	return nil
}

// DemoBackoffStrategies 演示不同的退避策略
func (c *Client) DemoBackoffStrategies(ctx context.Context) error {
	log.Println("\n=== Demo: Backoff Strategies ===")

	strategies := []struct {
		name     string
		strategy retry.BackoffStrategy
	}{
		{"Fixed", retry.BackoffFixed},
		{"Linear", retry.BackoffLinear},
		{"Exponential", retry.BackoffExponential},
	}

	for _, s := range strategies {
		log.Printf("\n--- Testing %s Backoff Strategy ---", s.name)

		// 创建一个会失败的函数来演示退避策略
		attemptTimes := make([]time.Time, 0)
		startTime := time.Now()

		// 使用重试库直接演示退避
		config := retry.DefaultConfig()
		config.MaxAttempts = 5
		config.InitialDelay = 100 * time.Millisecond
		config.MaxDelay = 2 * time.Second
		config.BackoffStrategy = s.strategy
		config.EnableJitter = true
		config.OnRetry = func(attempt int, lastErr error, nextDelay time.Duration) {
			elapsed := time.Since(startTime)
			log.Printf("  Attempt %d failed, next retry in %v (elapsed: %v)", attempt, nextDelay, elapsed)
			attemptTimes = append(attemptTimes, time.Now())
		}

		// 创建一个总是失败的函数
		failFn := func(ctx context.Context) (interface{}, error) {
			return nil, fmt.Errorf("simulated failure for %s strategy", s.name)
		}

		// 执行（会失败）
		_, err := retry.DoRetry(ctx, failFn, config)
		if err != nil {
			log.Printf("  All %d attempts failed as expected: %v", config.MaxAttempts, err)
		}
	}

	return nil
}

func main() {
	// 解析命令行参数
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	maxAttempts := flag.Int("max-attempts", 5, "Maximum retry attempts")
	initialDelay := flag.Duration("initial-delay", 100*time.Millisecond, "Initial retry delay")
	backoffStrategy := flag.String("backoff", "exponential", "Backoff strategy: fixed, linear, exponential")
	enableJitter := flag.Bool("jitter", true, "Enable jitter")
	demo := flag.String("demo", "all", "Demo to run: health, echo, flaky, interceptor, circuitbreaker, concurrent, backoff, all")
	numRequests := flag.Int("requests", 10, "Number of requests for flaky demo")
	numGoroutines := flag.Int("goroutines", 5, "Number of goroutines for concurrent demo")
	requestsPerGoroutine := flag.Int("requests-per-goroutine", 5, "Requests per goroutine for concurrent demo")

	flag.Parse()

	// 解析退避策略
	var strategy retry.BackoffStrategy
	switch *backoffStrategy {
	case "fixed":
		strategy = retry.BackoffFixed
	case "linear":
		strategy = retry.BackoffLinear
	case "exponential":
		strategy = retry.BackoffExponential
	default:
		log.Printf("Unknown backoff strategy: %s, using exponential", *backoffStrategy)
		strategy = retry.BackoffExponential
	}

	// 创建上下文，支持信号取消
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\nReceived shutdown signal, stopping...")
		cancel()
	}()

	// 创建客户端
	client := NewClient(*serverURL, *maxAttempts, *initialDelay, strategy, *enableJitter)

	log.Println("========================================")
	log.Println("Retry Library Demo Client")
	log.Println("========================================")
	log.Printf("Server URL: %s", *serverURL)
	log.Printf("Max Attempts: %d", *maxAttempts)
	log.Printf("Initial Delay: %v", *initialDelay)
	log.Printf("Backoff Strategy: %s", *backoffStrategy)
	log.Printf("Enable Jitter: %v", *enableJitter)
	log.Println("========================================")

	// 运行指定的演示
	var err error
	switch *demo {
	case "health":
		err = client.DemoHealthCheck(ctx)
	case "echo":
		err = client.DemoEcho(ctx)
	case "flaky":
		err = client.DemoFlakyService(ctx, *numRequests)
	case "interceptor":
		err = client.DemoWithInterceptor(ctx)
	case "circuitbreaker":
		err = client.DemoCircuitBreaker(ctx)
	case "concurrent":
		err = client.DemoConcurrentRequests(ctx, *numGoroutines, *requestsPerGoroutine)
	case "backoff":
		err = client.DemoBackoffStrategies(ctx)
	case "all":
		// 运行所有演示
		if err = client.DemoHealthCheck(ctx); err == nil {
			if err = client.DemoEcho(ctx); err == nil {
				if err = client.DemoFlakyService(ctx, *numRequests); err == nil {
					if err = client.DemoWithInterceptor(ctx); err == nil {
						if err = client.DemoBackoffStrategies(ctx); err == nil {
							err = client.DemoConcurrentRequests(ctx, *numGoroutines, *requestsPerGoroutine)
						}
					}
				}
			}
		}
	default:
		log.Printf("Unknown demo: %s", *demo)
		log.Println("Available demos: health, echo, flaky, interceptor, circuitbreaker, concurrent, backoff, all")
		os.Exit(1)
	}

	if err != nil {
		log.Printf("\nDemo completed with error: %v", err)
		os.Exit(1)
	}

	log.Println("\n========================================")
	log.Println("Demo completed successfully!")
	log.Println("========================================")
}
