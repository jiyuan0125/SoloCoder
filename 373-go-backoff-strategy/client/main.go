package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"backoff/common"
)

type Config struct {
	ServerAddr   string
	TaskID       string
	TaskType     string
	StrategyType string
	MaxRetries   int
	MaxWait      time.Duration
	InitialWait  time.Duration
	Increment    time.Duration
	Multiplier   float64
	JitterFactor float64
	DelayMs      int
	SucceedOn    int
}

func main() {
	config := parseArgs()

	if err := validateConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		printUsage()
		os.Exit(1)
	}

	req := buildRequest(config)

	client := NewRetryClient(config.ServerAddr)
	resp, err := client.Execute(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Request failed: %v\n", err)
		os.Exit(1)
	}

	printResponse(resp)
}

func parseArgs() *Config {
	config := &Config{}

	flag.StringVar(&config.ServerAddr, "server", "http://localhost:8080", "Server address")
	flag.StringVar(&config.TaskID, "task-id", "", "Task ID (required)")
	flag.StringVar(&config.TaskType, "task-type", common.TaskTypeMockFlaky, "Task type: mock_success, mock_failure, mock_non_retryable, mock_flaky")
	flag.StringVar(&config.StrategyType, "strategy", string(common.StrategyExponential), "Retry strategy: fixed, linear, exponential, exponential-jitter")
	flag.IntVar(&config.MaxRetries, "max-retries", 3, "Maximum number of retries")
	flag.DurationVar(&config.MaxWait, "max-wait", 0, "Maximum wait duration between retries (0 means no limit)")
	flag.DurationVar(&config.InitialWait, "initial-wait", 1*time.Second, "Initial wait duration")
	flag.DurationVar(&config.Increment, "increment", 1*time.Second, "Increment for linear strategy")
	flag.Float64Var(&config.Multiplier, "multiplier", 2.0, "Multiplier for exponential strategies")
	flag.Float64Var(&config.JitterFactor, "jitter-factor", 0.5, "Jitter factor (0-1) for exponential-jitter strategy")
	flag.IntVar(&config.DelayMs, "delay-ms", 0, "Simulated delay in milliseconds for each task attempt")
	flag.IntVar(&config.SucceedOn, "succeed-on", 3, "For mock_flaky: attempt number on which to succeed")

	flag.Parse()

	return config
}

func validateConfig(config *Config) error {
	if config.TaskID == "" {
		return fmt.Errorf("task-id is required")
	}

	if !common.IsValidTaskType(config.TaskType) {
		return fmt.Errorf("invalid task-type: %s", config.TaskType)
	}

	strategyType := common.StrategyType(config.StrategyType)
	if err := strategyType.Validate(); err != nil {
		return err
	}

	if config.MaxRetries < 0 {
		return fmt.Errorf("max-retries must be >= 0")
	}

	if config.InitialWait <= 0 {
		return fmt.Errorf("initial-wait must be > 0")
	}

	if config.MaxWait < 0 {
		return fmt.Errorf("max-wait must be >= 0")
	}

	if config.Multiplier <= 0 {
		return fmt.Errorf("multiplier must be > 0")
	}

	if config.JitterFactor < 0 || config.JitterFactor > 1 {
		return fmt.Errorf("jitter-factor must be between 0 and 1")
	}

	return nil
}

func buildRequest(config *Config) common.ExecuteRequest {
	req := common.ExecuteRequest{
		TaskID:   config.TaskID,
		TaskType: config.TaskType,
		StrategyConfig: common.StrategyConfig{
			Type:         common.StrategyType(config.StrategyType),
			MaxRetries:   config.MaxRetries,
			MaxWait:      config.MaxWait,
			InitialWait:  config.InitialWait,
			Increment:    config.Increment,
			Multiplier:   config.Multiplier,
			JitterFactor: config.JitterFactor,
		},
		TaskPayload: make(map[string]any),
	}

	if config.DelayMs > 0 {
		req.TaskPayload["delay_ms"] = config.DelayMs
	}

	if config.TaskType == common.TaskTypeMockFlaky {
		req.TaskPayload["succeed_on_attempt"] = config.SucceedOn
	}

	return req
}

func printResponse(resp *common.ExecuteResponse) {
	fmt.Printf("Task ID: %s\n", resp.TaskID)
	fmt.Printf("Success: %v\n", resp.Success)
	fmt.Printf("Total Attempts: %d\n", resp.Attempts)

	if len(resp.RetryEvents) > 0 {
		fmt.Println("\nRetry Events:")
		for i, event := range resp.RetryEvents {
			fmt.Printf("  Retry %d: attempt=%d, delay=%v, error=%q\n",
				i+1, event.Attempt, event.Delay, event.Error)
		}
	}

	if resp.Success && resp.Result != nil {
		fmt.Println("\nResult:")
		for k, v := range resp.Result {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}

	if resp.Error != "" {
		fmt.Printf("\nError: %s\n", resp.Error)
	}
}

func printUsage() {
	fmt.Println("\nUsage:")
	fmt.Println("  client -task-id <id> [options]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Println("  # Test exponential strategy with flaky task")
	fmt.Println("  client -task-id test1 -strategy exponential -task-type mock_flaky -succeed-on 3 -max-retries 5")
	fmt.Println("\n  # Test fixed interval strategy")
	fmt.Println("  client -task-id test2 -strategy fixed -task-type mock_success -initial-wait 500ms")
	fmt.Println("\n  # Test non-retryable error")
	fmt.Println("  client -task-id test3 -strategy exponential -task-type mock_non_retryable -max-retries 5")
}
