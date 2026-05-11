package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go-lockfree-queue/api"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	defaultServerAddr = "http://localhost:8403"
)

type Client struct {
	serverAddr string
	httpClient *http.Client
}

func NewClient(serverAddr string) *Client {
	return &Client{
		serverAddr: serverAddr,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) postJSON(path string, body interface{}, resp interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.serverAddr+path, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("server returned %d: %s", httpResp.StatusCode, string(body))
	}

	if resp != nil {
		return json.NewDecoder(httpResp.Body).Decode(resp)
	}
	return nil
}

func (c *Client) getJSON(path string, resp interface{}) error {
	httpResp, err := c.httpClient.Get(c.serverAddr + path)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("server returned %d: %s", httpResp.StatusCode, string(body))
	}

	return json.NewDecoder(httpResp.Body).Decode(resp)
}

func (c *Client) Bench(producers, messagesPerProducer int) error {
	req := api.BenchRequest{
		Producers:           producers,
		MessagesPerProducer: messagesPerProducer,
	}

	var resp api.BenchResponse
	if err := c.postJSON("/bench", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("bench failed: %s", resp.Message)
	}

	r := resp.Result
	fmt.Println("=== Benchmark Results ===")
	fmt.Printf("Total Messages:  %d\n", r.TotalMessages)
	fmt.Printf("Producers:       %d\n", r.Producers)
	fmt.Printf("Duration:        %d ms\n", r.DurationMs)
	fmt.Printf("Throughput:      %.2f msg/s\n", r.Throughput)
	fmt.Println()
	fmt.Println("=== Latency (ms) ===")
	fmt.Printf("Min:    %.6f\n", r.LatencyMin)
	fmt.Printf("Avg:    %.6f\n", r.LatencyAvg)
	fmt.Printf("P50:    %.6f\n", r.LatencyP50)
	fmt.Printf("P95:    %.6f\n", r.LatencyP95)
	fmt.Printf("P99:    %.6f\n", r.LatencyP99)
	fmt.Printf("Max:    %.6f\n", r.LatencyMax)

	return nil
}

func (c *Client) Verify(messageCount int, useLock bool) error {
	messages := make([]string, messageCount)
	for i := 0; i < messageCount; i++ {
		messages[i] = fmt.Sprintf("verify-msg-%06d", i)
	}

	req := api.VerifyRequest{
		Messages: messages,
		UseLock:  useLock,
	}

	var resp api.VerifyResponse
	if err := c.postJSON("/verify", req, &resp); err != nil {
		return err
	}

	queueType := "lock-free"
	if useLock {
		queueType = "locked"
	}

	fmt.Printf("=== Verify (%s) ===\n", queueType)
	fmt.Printf("Messages:      %d\n", messageCount)
	fmt.Printf("Expected size: %d\n", len(resp.Expected))
	fmt.Printf("Actual size:   %d\n", len(resp.Actual))
	fmt.Printf("Success:       %v\n", resp.Success)

	if resp.Message != "" {
		fmt.Printf("Message:       %s\n", resp.Message)
	}

	if !resp.Success {
		os.Exit(1)
	}

	return nil
}

func (c *Client) Stats() error {
	var resp api.StatsResponse
	if err := c.getJSON("/stats", &resp); err != nil {
		return err
	}

	fmt.Println("=== Queue Statistics ===")
	fmt.Printf("Depth:           %d\n", resp.Depth)
	fmt.Printf("Enqueue Total:   %d\n", resp.EnqueueTotal)
	fmt.Printf("Processed Total: %d\n", resp.ProcessedTotal)
	fmt.Println()
	fmt.Println("=== CAS Operations ===")
	fmt.Printf("Success:         %d\n", resp.CASSuccess)
	fmt.Printf("Failure:         %d\n", resp.CASFailure)
	if resp.CASSuccess+resp.CASFailure > 0 {
		successRate := float64(resp.CASSuccess) / float64(resp.CASSuccess+resp.CASFailure) * 100
		fmt.Printf("Success Rate:    %.2f%%\n", successRate)
	}
	fmt.Println()
	fmt.Println("=== Memory ===")
	fmt.Printf("Allocs Total:    %d\n", resp.AllocsTotal)
	fmt.Printf("Pool Hits:       %d\n", resp.PoolHits)
	fmt.Printf("Pool Misses:     %d\n", resp.PoolMisses)
	if resp.PoolHits+resp.PoolMisses > 0 {
		hitRate := float64(resp.PoolHits) / float64(resp.PoolHits+resp.PoolMisses) * 100
		fmt.Printf("Pool Hit Rate:   %.2f%%\n", hitRate)
	}

	return nil
}

func printUsage() {
	fmt.Println("Usage: client [global-flags] <command> [command-flags]")
	fmt.Println()
	fmt.Println("Global Flags:")
	fmt.Println("  -server string   Server address (default \"http://localhost:8403\")")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  bench    Run performance benchmark")
	fmt.Println("  verify   Verify queue correctness")
	fmt.Println("  stats    Show queue statistics")
	fmt.Println()
	fmt.Println("Bench Flags:")
	fmt.Println("  -producers int       Number of producer goroutines (default 4)")
	fmt.Println("  -messages int        Messages per producer (default 1000)")
	fmt.Println()
	fmt.Println("Verify Flags:")
	fmt.Println("  -count int           Number of messages to verify (default 10000)")
	fmt.Println("  -lock                Use locked queue instead of lock-free")
}

func main() {
	globalFlags := flag.NewFlagSet("global", flag.ExitOnError)
	serverAddr := globalFlags.String("server", defaultServerAddr, "Server address")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	globalArgs := []string{}
	cmdIdx := 1
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "-server" {
			if i+1 < len(os.Args) {
				*serverAddr = os.Args[i+1]
				i++
			}
			continue
		}
		if os.Args[i] == "-h" || os.Args[i] == "-help" || os.Args[i] == "--help" {
			printUsage()
			os.Exit(0)
		}
		if os.Args[i][0] != '-' {
			cmdIdx = i
			break
		}
		globalArgs = append(globalArgs, os.Args[i])
	}

	_ = globalFlags.Parse(globalArgs)

	client := NewClient(*serverAddr)

	cmd := os.Args[cmdIdx]
	args := os.Args[cmdIdx+1:]

	switch cmd {
	case "bench":
		benchFlags := flag.NewFlagSet("bench", flag.ExitOnError)
		producers := benchFlags.Int("producers", 4, "Number of producer goroutines")
		messages := benchFlags.Int("messages", 1000, "Messages per producer")
		benchFlags.Parse(args)

		if err := client.Bench(*producers, *messages); err != nil {
			fmt.Fprintf(os.Stderr, "bench error: %v\n", err)
			os.Exit(1)
		}

	case "verify":
		verifyFlags := flag.NewFlagSet("verify", flag.ExitOnError)
		count := verifyFlags.Int("count", 10000, "Number of messages to verify")
		useLock := verifyFlags.Bool("lock", false, "Use locked queue")
		verifyFlags.Parse(args)

		if err := client.Verify(*count, *useLock); err != nil {
			fmt.Fprintf(os.Stderr, "verify error: %v\n", err)
			os.Exit(1)
		}

	case "stats":
		if err := client.Stats(); err != nil {
			fmt.Fprintf(os.Stderr, "stats error: %v\n", err)
			os.Exit(1)
		}

	case "-h", "-help", "--help", "help":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
