package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"murmurhash3/pkg/api"
)

type clientConfig struct {
	serverURL string
}

func newClient(serverURL string) *clientConfig {
	return &clientConfig{serverURL: serverURL}
}

func (c *clientConfig) postJSON(path string, req, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpResp, err := http.Post(c.serverURL+path, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server error: %d - %s", httpResp.StatusCode, string(respBody))
	}

	if resp != nil {
		if err := json.Unmarshal(respBody, resp); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}

func cmdHash(args []string) {
	fs := flag.NewFlagSet("hash", flag.ExitOnError)
	seedFlag := fs.Uint("seed", 0, "hash seed")
	hashTypeFlag := fs.String("type", "32", "hash type: 32 or 128")
	serverFlag := fs.String("server", "http://localhost:8410", "server URL")

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if fs.NArg() < 1 {
		fmt.Println("Usage: hash [flags] <data>")
		fs.PrintDefaults()
		os.Exit(1)
	}

	data := fs.Arg(0)
	seed := uint32(*seedFlag)
	hashType := *hashTypeFlag
	client := newClient(*serverFlag)

	if hashType == "128" {
		req := api.HashRequest{Data: data, Seed: seed}
		var resp api.Hash128Response
		if err := client.postJSON("/hash/128", req, &resp); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("Hash128: high=0x%016x, low=0x%016x\n", resp.High, resp.Low)
		fmt.Printf("Hex: %s\n", resp.Hex)
		fmt.Printf("Seed: %d\n", resp.Seed)
	} else {
		req := api.HashRequest{Data: data, Seed: seed}
		var resp api.Hash32Response
		if err := client.postJSON("/hash/32", req, &resp); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("Hash32: %d\n", resp.Hash)
		fmt.Printf("Hex: %s\n", resp.Hex)
		fmt.Printf("Seed: %d\n", resp.Seed)
	}
}

func generateTestData(count int) []string {
	data := make([]string, count)
	for i := 0; i < count; i++ {
		data[i] = "log-entry-" + strconv.Itoa(i) + "-test-data-" + strconv.Itoa(i*13)
	}
	return data
}

func cmdBenchmark(args []string) {
	fs := flag.NewFlagSet("benchmark", flag.ExitOnError)
	seedFlag := fs.Uint("seed", 0, "hash seed")
	countFlag := fs.Int("count", 10000, "number of test entries")
	hashTypeFlag := fs.String("type", "32", "hash type: 32 or 128")
	bucketFlag := fs.Int("buckets", 256, "number of buckets")
	serverFlag := fs.String("server", "http://localhost:8410", "server URL")

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	seed := uint32(*seedFlag)
	count := *countFlag
	hashType := *hashTypeFlag
	bucketCount := *bucketFlag
	client := newClient(*serverFlag)

	fmt.Printf("Generating %d test entries...\n", count)
	testData := generateTestData(count)

	fmt.Printf("Running distribution analysis (seed=%d, buckets=%d)...\n", seed, bucketCount)

	req := api.DistributionRequest{
		Data:        testData,
		Seed:        seed,
		HashType:    hashType,
		BucketCount: bucketCount,
	}

	var resp api.DistributionReport
	if err := client.postJSON("/distribution", req, &resp); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("\n=== Distribution Report ===")
	fmt.Printf("Sample Count:   %d\n", resp.SampleCount)
	fmt.Printf("Bucket Count:   %d\n", resp.BucketCount)
	fmt.Printf("Hash Type:      %s\n", hashType)
	fmt.Printf("Seed:           %d\n", seed)
	fmt.Println("---")
	fmt.Printf("Min Hash:       %d (0x%016x)\n", resp.Min, resp.Min)
	fmt.Printf("Max Hash:       %d (0x%016x)\n", resp.Max, resp.Max)
	fmt.Printf("Mean:           %.2f\n", resp.Mean)
	fmt.Printf("Std Dev:        %.2f\n", resp.StdDev)
	fmt.Printf("Variance:       %.2f\n", resp.Variance)
	fmt.Println("---")
	fmt.Printf("Chi-Square:     %.4f\n", resp.ChiSquare)
	fmt.Printf("Uniform Score:  %.4f (%.1f%%)\n", resp.UniformScore, resp.UniformScore*100)

	if resp.UniformScore >= 0.8 {
		fmt.Println("\n✓ Excellent distribution!")
	} else if resp.UniformScore >= 0.6 {
		fmt.Println("\n✓ Good distribution.")
	} else if resp.UniformScore >= 0.4 {
		fmt.Println("\n⚠ Moderate distribution.")
	} else {
		fmt.Println("\n✗ Poor distribution - try a different seed.")
	}
}

func printUsage() {
	fmt.Println("MurmurHash3 Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  hash      Compute hash for a single input")
	fmt.Println("  benchmark Run distribution benchmark")
	fmt.Println()
	fmt.Println("Use 'client <command> -h' for command-specific help.")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "hash":
		cmdHash(args)
	case "benchmark":
		cmdBenchmark(args)
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
