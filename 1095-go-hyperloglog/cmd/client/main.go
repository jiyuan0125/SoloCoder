package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/example/hyperloglog/pkg/api"
)

const defaultServer = "http://localhost:8080"

type client struct {
	serverURL string
}

func newClient(serverURL string) *client {
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "http://" + serverURL
	}
	return &client{serverURL: serverURL}
}

func (c *client) post(endpoint string, req interface{}, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpResp, err := http.Post(c.serverURL+endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return fmt.Errorf("%s", errResp.Message)
		}
		return fmt.Errorf("HTTP error %d: %s", httpResp.StatusCode, string(respBody))
	}

	if resp != nil {
		if err := json.Unmarshal(respBody, resp); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func cmdCreate(args []string) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	name := fs.String("name", "", "HLL instance name (required)")
	precision := fs.Int("precision", 14, "Precision (4-18, default 14)")
	server := fs.String("server", defaultServer, "Server URL")
	fs.Parse(args)

	if *name == "" {
		fmt.Println("Error: -name is required")
		os.Exit(1)
	}

	c := newClient(*server)
	req := api.CreateHLLRequest{
		Name:      *name,
		Precision: *precision,
	}

	var resp api.CreateHLLResponse
	if err := c.post("/api/hll/create", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(resp.Message)
}

func cmdAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	name := fs.String("name", "", "HLL instance name (required)")
	file := fs.String("file", "", "File to read values from (one per line, use - for stdin)")
	batchSize := fs.Int("batch", 1000, "Batch size for upload (default 1000)")
	server := fs.String("server", defaultServer, "Server URL")
	fs.Parse(args)

	if *name == "" {
		fmt.Println("Error: -name is required")
		os.Exit(1)
	}

	if *file == "" {
		fmt.Println("Error: -file is required")
		os.Exit(1)
	}

	var reader *bufio.Reader
	if *file == "-" {
		reader = bufio.NewReader(os.Stdin)
	} else {
		f, err := os.Open(*file)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		reader = bufio.NewReader(f)
	}

	c := newClient(*server)
	var batch []string
	total := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		line = strings.TrimSpace(line)
		if line != "" {
			batch = append(batch, line)
		}

		if len(batch) >= *batchSize || (err == io.EOF && len(batch) > 0) {
			req := api.AddRequest{
				Name:   *name,
				Values: batch,
			}

			var resp api.AddResponse
			if err := c.post("/api/hll/add", req, &resp); err != nil {
				fmt.Printf("Error adding values: %v\n", err)
				os.Exit(1)
			}

			total += len(batch)
			fmt.Printf("Added %d values, total: %d\n", len(batch), total)
			batch = nil
		}

		if err == io.EOF {
			break
		}
	}

	fmt.Printf("Done. Total values added: %d\n", total)
}

func cmdCount(args []string) {
	fs := flag.NewFlagSet("count", flag.ExitOnError)
	name := fs.String("name", "", "HLL instance name (required)")
	server := fs.String("server", defaultServer, "Server URL")
	fs.Parse(args)

	if *name == "" {
		fmt.Println("Error: -name is required")
		os.Exit(1)
	}

	c := newClient(*server)
	req := api.CountRequest{Name: *name}

	var resp api.CountResponse
	if err := c.post("/api/hll/count", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Estimated count: %.2f\n", resp.Count)
}

func cmdMerge(args []string) {
	fs := flag.NewFlagSet("merge", flag.ExitOnError)
	newName := fs.String("new", "", "New HLL instance name (required)")
	sources := fs.String("sources", "", "Comma-separated list of source HLL names (required)")
	server := fs.String("server", defaultServer, "Server URL")
	fs.Parse(args)

	if *newName == "" {
		fmt.Println("Error: -new is required")
		os.Exit(1)
	}

	if *sources == "" {
		fmt.Println("Error: -sources is required")
		os.Exit(1)
	}

	sourceList := strings.Split(*sources, ",")
	for i, s := range sourceList {
		sourceList[i] = strings.TrimSpace(s)
	}

	if len(sourceList) < 2 {
		fmt.Println("Error: At least 2 sources are required")
		os.Exit(1)
	}

	c := newClient(*server)
	req := api.MergeRequest{
		NewName: *newName,
		Sources: sourceList,
	}

	var resp api.MergeResponse
	if err := c.post("/api/hll/merge", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(resp.Message)
}

func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	server := fs.String("server", defaultServer, "Server URL")
	fs.Parse(args)

	c := newClient(*server)
	var resp api.ListResponse
	if err := c.post("/api/hll/list", struct{}{}, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(resp.HLLs) == 0 {
		fmt.Println("No HLL instances found.")
		return
	}

	fmt.Printf("%-20s %-10s %-20s\n", "Name", "Precision", "Estimated Count")
	fmt.Println(strings.Repeat("-", 50))
	for _, info := range resp.HLLs {
		fmt.Printf("%-20s %-10d %-20.2f\n", info.Name, info.Precision, info.Count)
	}
}

func cmdExport(args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	name := fs.String("name", "", "HLL instance name (required)")
	output := fs.String("output", "", "Output file (required)")
	server := fs.String("server", defaultServer, "Server URL")
	fs.Parse(args)

	if *name == "" {
		fmt.Println("Error: -name is required")
		os.Exit(1)
	}

	if *output == "" {
		fmt.Println("Error: -output is required")
		os.Exit(1)
	}

	c := newClient(*server)
	req := api.ExportRequest{Name: *name}

	var resp api.ExportResponse
	if err := c.post("/api/hll/export", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*output, resp.Data, 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Exported %d bytes to %s\n", len(resp.Data), *output)
}

func cmdImport(args []string) {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	name := fs.String("name", "", "New HLL instance name (required)")
	input := fs.String("input", "", "Input file (required)")
	server := fs.String("server", defaultServer, "Server URL")
	fs.Parse(args)

	if *name == "" {
		fmt.Println("Error: -name is required")
		os.Exit(1)
	}

	if *input == "" {
		fmt.Println("Error: -input is required")
		os.Exit(1)
	}

	data, err := os.ReadFile(*input)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	c := newClient(*server)
	req := api.ImportRequest{
		Name: *name,
		Data: data,
	}

	var resp api.ImportResponse
	if err := c.post("/api/hll/import", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(resp.Message)
}

func cmdBenchmark(args []string) {
	fs := flag.NewFlagSet("benchmark", flag.ExitOnError)
	name := fs.String("name", "benchmark", "HLL instance name")
	precision := fs.Int("precision", 14, "Precision (4-18, default 14)")
	count := fs.Int("count", 10000, "Number of unique values to generate (default 10000)")
	server := fs.String("server", defaultServer, "Server URL")
	batchSize := fs.Int("batch", 1000, "Batch size for upload (default 1000)")
	fs.Parse(args)

	fmt.Printf("Benchmarking with %d unique values, precision=%d...\n", *count, *precision)

	rand.Seed(time.Now().UnixNano())
	values := make([]string, *count)
	for i := 0; i < *count; i++ {
		values[i] = strconv.FormatUint(rand.Uint64(), 36)
	}

	c := newClient(*server)

	fmt.Printf("1. Creating HLL instance '%s'...\n", *name)
	createReq := api.CreateHLLRequest{Name: *name, Precision: *precision}
	if err := c.post("/api/hll/create", createReq, &api.CreateHLLResponse{}); err != nil {
		fmt.Printf("Error creating HLL: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("2. Adding values in batches of %d...\n", *batchSize)
	startAdd := time.Now()
	for i := 0; i < *count; i += *batchSize {
		end := i + *batchSize
		if end > *count {
			end = *count
		}

		addReq := api.AddRequest{
			Name:   *name,
			Values: values[i:end],
		}

		if err := c.post("/api/hll/add", addReq, &api.AddResponse{}); err != nil {
			fmt.Printf("Error adding values: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("   Added %d/%d\n", end, *count)
	}
	addDuration := time.Since(startAdd)

	fmt.Printf("3. Getting count estimate...\n")
	countReq := api.CountRequest{Name: *name}
	var countResp api.CountResponse
	if err := c.post("/api/hll/count", countReq, &countResp); err != nil {
		fmt.Printf("Error getting count: %v\n", err)
		os.Exit(1)
	}

	actual := float64(*count)
	estimated := countResp.Count
	errorPercent := (estimated - actual) / actual * 100

	fmt.Println("\n=== Results ===")
	fmt.Printf("Actual unique count:    %d\n", *count)
	fmt.Printf("Estimated count:        %.2f\n", estimated)
	fmt.Printf("Error percentage:       %.4f%%\n", errorPercent)
	fmt.Printf("Add throughput:         %.2f ops/sec\n", float64(*count)/addDuration.Seconds())
	fmt.Printf("Total add time:         %v\n", addDuration)
	numBuckets := 1 << uint(*precision)
	fmt.Printf("Expected error rate:    ~%.1f%%\n", 1.04/math.Sqrt(float64(numBuckets))*100)
}

func printUsage() {
	fmt.Println("HLL Client - Command line interface for HyperLogLog server")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  hllclient <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create     Create a new HLL instance")
	fmt.Println("  add        Add values from a file to an HLL instance")
	fmt.Println("  count      Get the estimated count from an HLL instance")
	fmt.Println("  merge      Merge multiple HLL instances into a new one")
	fmt.Println("  list       List all HLL instances")
	fmt.Println("  export     Export an HLL instance to a file")
	fmt.Println("  import     Import an HLL instance from a file")
	fmt.Println("  benchmark  Run a benchmark test")
	fmt.Println()
	fmt.Println("Use 'hllclient <command> -h' for command-specific help")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "create":
		cmdCreate(args)
	case "add":
		cmdAdd(args)
	case "count":
		cmdCount(args)
	case "merge":
		cmdMerge(args)
	case "list":
		cmdList(args)
	case "export":
		cmdExport(args)
	case "import":
		cmdImport(args)
	case "benchmark":
		cmdBenchmark(args)
	case "-h", "-help", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
