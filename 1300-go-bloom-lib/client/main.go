package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"bloom/common"
)

type Config struct {
	ServerAddr string
}

func getConfig() *Config {
	serverAddr := "http://localhost:8202"

	if envAddr := os.Getenv("BLOOM_SERVER_ADDR"); envAddr != "" {
		serverAddr = envAddr
	}

	return &Config{
		ServerAddr: serverAddr,
	}
}

func makeRequest(cfg *Config, endpoint string, reqBody, respBody interface{}) error {
	var body io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewBuffer(data)
	}

	resp, err := http.Post(cfg.ServerAddr+endpoint, "application/json", body)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		if err := json.Unmarshal(data, &errResp); err == nil {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(data))
	}

	if respBody != nil {
		if err := json.Unmarshal(data, respBody); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func printUsage() {
	fmt.Println("Bloom Filter CLI - Command Line Interface for Bloom Filter Server")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  bloom-cli <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create      Create a new bloom filter")
	fmt.Println("  add         Add an element to a filter")
	fmt.Println("  contains    Check if an element exists in a filter")
	fmt.Println("  info        Get information about a filter")
	fmt.Println("  clear       Clear a filter")
	fmt.Println("  list        List all filters")
	fmt.Println("  export      Export a filter to base64")
	fmt.Println("  import      Import a filter from base64")
	fmt.Println("  delete      Delete a filter")
	fmt.Println()
	fmt.Println("Server Address Configuration:")
	fmt.Println("  Default: http://localhost:8202")
	fmt.Println("  Environment variable: BLOOM_SERVER_ADDR")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  bloom-cli create -name myfilter -capacity 10000 -fpr 0.001")
	fmt.Println("  bloom-cli add -name myfilter -element \"hello world\"")
	fmt.Println("  bloom-cli contains -name myfilter -element \"hello world\"")
	fmt.Println("  bloom-cli info -name myfilter")
	fmt.Println("  bloom-cli list")
	fmt.Println("  bloom-cli export -name myfilter -output filter_data.txt")
	fmt.Println("  bloom-cli import -name myfilter2 -input filter_data.txt")
}

func cmdCreate(args []string) error {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	name := fs.String("name", "", "Filter name")
	capacity := fs.Uint64("capacity", 0, "Expected capacity (number of elements)")
	fpr := fs.Float64("fpr", 0.01, "Target false positive rate (0 < fpr < 1)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("-name is required")
	}

	cfg := getConfig()
	req := common.CreateFilterRequest{
		Name:              *name,
		ExpectedCapacity:   *capacity,
		FalsePositiveRate: *fpr,
	}

	var resp common.CreateFilterResponse
	if err := makeRequest(cfg, "/create", req, &resp); err != nil {
		return err
	}

	fmt.Printf("Filter '%s' created successfully.\n", resp.Name)
	fmt.Printf("  Bit Array Size: %d bits\n", resp.BitArraySize)
	fmt.Printf("  Hash Count: %d\n", resp.HashCount)
	fmt.Printf("  Expected Capacity: %d elements\n", resp.ExpectedCapacity)

	return nil
}

func cmdAdd(args []string) error {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	name := fs.String("name", "", "Filter name")
	element := fs.String("element", "", "Element to add")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("-name is required")
	}

	cfg := getConfig()
	req := common.AddElementRequest{
		Name:    *name,
		Element: *element,
	}

	var resp common.AddElementResponse
	if err := makeRequest(cfg, "/add", req, &resp); err != nil {
		return err
	}

	fmt.Println(resp.Message)
	return nil
}

func cmdContains(args []string) error {
	fs := flag.NewFlagSet("contains", flag.ExitOnError)
	name := fs.String("name", "", "Filter name")
	element := fs.String("element", "", "Element to check")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("-name is required")
	}

	cfg := getConfig()
	req := common.ContainsElementRequest{
		Name:    *name,
		Element: *element,
	}

	var resp common.ContainsElementResponse
	if err := makeRequest(cfg, "/contains", req, &resp); err != nil {
		return err
	}

	if resp.Contains {
		fmt.Printf("Element MAY exist in filter '%s' (could be false positive)\n", *name)
	} else {
		fmt.Printf("Element definitely NOT in filter '%s'\n", *name)
	}

	return nil
}

func cmdInfo(args []string) error {
	fs := flag.NewFlagSet("info", flag.ExitOnError)
	name := fs.String("name", "", "Filter name")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("-name is required")
	}

	cfg := getConfig()
	req := common.GetFilterInfoRequest{
		Name: *name,
	}

	var resp common.GetFilterInfoResponse
	if err := makeRequest(cfg, "/info", req, &resp); err != nil {
		return err
	}

	fmt.Printf("Filter: %s\n", resp.Name)
	fmt.Printf("  Elements Added: %d\n", resp.ElementsAdded)
	fmt.Printf("  Fill Rate: %.4f%%\n", resp.FillRate*100)
	fmt.Printf("  Current False Positive Rate: %.6f\n", resp.CurrentFalsePositive)
	fmt.Printf("  Target False Positive Rate: %.6f\n", resp.TargetFalsePositive)
	fmt.Printf("  Remaining Capacity: ~%d elements\n", resp.RemainingCapacity)
	fmt.Printf("  Expected Capacity: %d elements\n", resp.ExpectedCapacity)
	fmt.Printf("  Bit Array Size: %d bits\n", resp.BitArraySize)
	fmt.Printf("  Hash Count: %d\n", resp.HashCount)

	return nil
}

func cmdClear(args []string) error {
	fs := flag.NewFlagSet("clear", flag.ExitOnError)
	name := fs.String("name", "", "Filter name")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("-name is required")
	}

	cfg := getConfig()
	req := common.ClearFilterRequest{
		Name: *name,
	}

	var resp common.ClearFilterResponse
	if err := makeRequest(cfg, "/clear", req, &resp); err != nil {
		return err
	}

	fmt.Println(resp.Message)
	return nil
}

func cmdList(args []string) error {
	cfg := getConfig()
	var resp common.ListFiltersResponse
	if err := makeRequest(cfg, "/list", nil, &resp); err != nil {
		return err
	}

	if len(resp.Filters) == 0 {
		fmt.Println("No filters available.")
		return nil
	}

	fmt.Println("Available filters:")
	for i, name := range resp.Filters {
		fmt.Printf("  %d. %s\n", i+1, name)
	}

	return nil
}

func cmdExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	name := fs.String("name", "", "Filter name")
	output := fs.String("output", "", "Output file path (optional, defaults to stdout)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("-name is required")
	}

	cfg := getConfig()
	req := common.ExportFilterRequest{
		Name: *name,
	}

	var resp common.ExportFilterResponse
	if err := makeRequest(cfg, "/export", req, &resp); err != nil {
		return err
	}

	if *output != "" {
		if err := os.WriteFile(*output, []byte(resp.Data), 0644); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
		fmt.Printf("Filter '%s' exported to %s\n", resp.Name, *output)
	} else {
		fmt.Printf("Filter '%s' data:\n%s\n", resp.Name, resp.Data)
	}

	return nil
}

func cmdImport(args []string) error {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	name := fs.String("name", "", "Filter name")
	input := fs.String("input", "", "Input file path")
	data := fs.String("data", "", "Base64 encoded data (alternative to -input)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("-name is required")
	}

	var filterData string
	if *input != "" {
		fileData, err := os.ReadFile(*input)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		filterData = string(fileData)
	} else if *data != "" {
		filterData = *data
	} else {
		return fmt.Errorf("either -input or -data is required")
	}

	cfg := getConfig()
	req := common.ImportFilterRequest{
		Name: *name,
		Data: filterData,
	}

	var resp common.ImportFilterResponse
	if err := makeRequest(cfg, "/import", req, &resp); err != nil {
		return err
	}

	fmt.Println(resp.Message)
	return nil
}

func cmdDelete(args []string) error {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	name := fs.String("name", "", "Filter name")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("-name is required")
	}

	cfg := getConfig()
	req := common.DeleteFilterRequest{
		Name: *name,
	}

	var resp common.DeleteFilterResponse
	if err := makeRequest(cfg, "/delete", req, &resp); err != nil {
		return err
	}

	fmt.Println(resp.Message)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error
	switch command {
	case "create":
		err = cmdCreate(args)
	case "add":
		err = cmdAdd(args)
	case "contains":
		err = cmdContains(args)
	case "info":
		err = cmdInfo(args)
	case "clear":
		err = cmdClear(args)
	case "list":
		err = cmdList(args)
	case "export":
		err = cmdExport(args)
	case "import":
		err = cmdImport(args)
	case "delete":
		err = cmdDelete(args)
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println()
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
