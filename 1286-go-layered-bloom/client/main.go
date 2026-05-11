package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"layered-bloom/common"
	"layered-bloom/core"
)

var baseURL string

func makeRequest(method, path string, body interface{}, target interface{}) error {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return err
		}
	}
	url := baseURL + path
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if target != nil {
		return json.Unmarshal(data, target)
	}
	return nil
}

func cmdInsert(args []string) error {
	fs := flag.NewFlagSet("insert", flag.ExitOnError)
	fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: client insert <element>")
	}
	element := fs.Arg(0)
	var resp common.InsertResponse
	err := makeRequest("POST", "/insert", common.InsertRequest{Element: element}, &resp)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("insert failed: %s", resp.Message)
	}
	if resp.WasExisting {
		fmt.Printf("Element exists in layer %d (duplicate, not inserted)\n", resp.LayerIndex)
	} else {
		fmt.Printf("Inserted successfully into layer %d\n", resp.LayerIndex)
	}
	return nil
}

func cmdQuery(args []string) error {
	fs := flag.NewFlagSet("query", flag.ExitOnError)
	fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: client query <element>")
	}
	element := fs.Arg(0)
	var resp common.QueryResponse
	err := makeRequest("GET", "/query?element="+element, nil, &resp)
	if err != nil {
		return err
	}
	if resp.Exists {
		fmt.Println("Element MAY exist (positive)")
	} else {
		fmt.Println("Element does NOT exist (negative)")
	}
	return nil
}

func cmdDelete(args []string) error {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: client delete <element>")
	}
	element := fs.Arg(0)
	var resp common.DeleteResponse
	err := makeRequest("POST", "/delete", common.DeleteRequest{Element: element}, &resp)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("delete failed: %s", resp.Message)
	}
	if resp.Removed {
		fmt.Println("Removed successfully")
	} else {
		fmt.Println("Element not found, nothing removed")
	}
	return nil
}

func cmdStats(_ []string) error {
	var resp common.StatsResponse
	err := makeRequest("GET", "/stats", nil, &resp)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("stats failed: %s", resp.Message)
	}
	stats := resp.Stats
	fmt.Printf("Layered Bloom Filter Stats:\n")
	fmt.Printf("  Total: %d / %d elements, %d layers\n", stats.TotalCount, stats.TotalCapacity, stats.LayerCount)
	for _, ls := range stats.Layers {
		warn := ""
		if ls.HasWarning {
			warn = " [FPR WARNING]"
		}
		full := ""
		if ls.IsFull {
			full = " [FULL]"
		}
		fmt.Printf("  Layer %d: %d/%d, hash=%d, fpr=%.6f/%.6f%s%s\n",
			ls.Index, ls.Count, ls.Capacity, ls.HashFunctions,
			ls.CurrentFPR, ls.TargetFPR, warn, full)
	}
	return nil
}

func cmdReset(_ []string) error {
	var resp common.ResetResponse
	err := makeRequest("POST", "/reset", nil, &resp)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("reset failed: %s", resp.Message)
	}
	fmt.Println(resp.Message)
	return nil
}

func cmdConfig(args []string) error {
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	layersFlag := fs.String("layers", "", "Comma-separated layers: capacity:hash:fpr,... (e.g., 1000:8:0.001,5000:6:0.005)")
	autoExpand := fs.Bool("auto-expand", true, "Auto-expand when all layers full")
	warnFactor := fs.Float64("warn-factor", 2.0, "Warning FPR factor over target")
	maxLayers := fs.Int("max-auto-layers", 10, "Maximum auto-expand layers")
	fs.Parse(args)
	if *layersFlag == "" {
		return fmt.Errorf("usage: client config -layers <spec> [-auto-expand] [-warn-factor] [-max-auto-layers]")
	}
	layerSpecs := strings.Split(*layersFlag, ",")
	layers := make([]common.LayerConfig, 0, len(layerSpecs))
	for _, spec := range layerSpecs {
		parts := strings.Split(spec, ":")
		if len(parts) != 3 {
			return fmt.Errorf("invalid layer spec: %s (format: capacity:hash:fpr)", spec)
		}
		var capacity int
		var hash int
		var fpr float64
		_, err := fmt.Sscanf(parts[0], "%d", &capacity)
		if err != nil {
			return fmt.Errorf("invalid capacity in %s: %v", spec, err)
		}
		_, err = fmt.Sscanf(parts[1], "%d", &hash)
		if err != nil {
			return fmt.Errorf("invalid hash count in %s: %v", spec, err)
		}
		_, err = fmt.Sscanf(parts[2], "%f", &fpr)
		if err != nil {
			return fmt.Errorf("invalid fpr in %s: %v", spec, err)
		}
		layers = append(layers, common.LayerConfig{
			Capacity:      capacity,
			HashFunctions: hash,
			TargetFPR:     fpr,
		})
	}
	req := common.ConfigRequest{
		Layers:           layers,
		AutoExpandOnFull: *autoExpand,
		WarningFPRFactor: *warnFactor,
		MaxAutoLayers:    *maxLayers,
	}
	var resp common.ConfigResponse
	err := makeRequest("POST", "/config", req, &resp)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("config failed: %s", resp.Message)
	}
	fmt.Println(resp.Message)
	fmt.Printf("  Layers: %d\n", len(layers))
	for i, l := range layers {
		fmt.Printf("    Layer %d: capacity=%d, hash=%d, fpr=%.6f\n",
			i, l.Capacity, l.HashFunctions, l.TargetFPR)
	}
	fmt.Printf("  Auto-expand: %v, max: %d\n", *autoExpand, *maxLayers)
	return nil
}

func readLinesFromFile(filename string) ([]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result, nil
}

func cmdBatchInsert(args []string) error {
	fs := flag.NewFlagSet("batch-insert", flag.ExitOnError)
	fileFlag := fs.String("file", "", "File containing elements (one per line)")
	fs.Parse(args)
	var elements []string
	if *fileFlag != "" {
		var err error
		elements, err = readLinesFromFile(*fileFlag)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}
	} else if fs.NArg() > 0 {
		elements = fs.Args()
	} else {
		return fmt.Errorf("usage: client batch-insert -file <filename> or <elem1> <elem2> ...")
	}
	if len(elements) == 0 {
		return fmt.Errorf("no elements to insert")
	}
	var resp common.BatchInsertResponse
	err := makeRequest("POST", "/batch/insert", common.BatchInsertRequest{Elements: elements}, &resp)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("batch insert failed")
	}
	fmt.Printf("Batch insert: total=%d, inserted=%d, existing=%d, failed=%d\n",
		resp.Total, resp.Inserted, resp.Existing, resp.Failed)
	for _, r := range resp.Results {
		if !r.Success {
			fmt.Printf("  FAIL: %s\n", r.Element)
		} else if r.WasExisting {
			fmt.Printf("  EXISTS: %s (layer %d)\n", r.Element, r.LayerIndex)
		} else {
			fmt.Printf("  INSERT: %s (layer %d)\n", r.Element, r.LayerIndex)
		}
	}
	return nil
}

func cmdBatchQuery(args []string) error {
	fs := flag.NewFlagSet("batch-query", flag.ExitOnError)
	fileFlag := fs.String("file", "", "File containing elements (one per line)")
	showOnly := fs.String("show-only", "all", "Show only: exists, notexists, or all")
	fs.Parse(args)
	var elements []string
	if *fileFlag != "" {
		var err error
		elements, err = readLinesFromFile(*fileFlag)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}
	} else if fs.NArg() > 0 {
		elements = fs.Args()
	} else {
		return fmt.Errorf("usage: client batch-query -file <filename> or <elem1> <elem2> ...")
	}
	if len(elements) == 0 {
		return fmt.Errorf("no elements to query")
	}
	var resp common.BatchQueryResponse
	err := makeRequest("POST", "/batch/query", common.BatchQueryRequest{Elements: elements}, &resp)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("batch query failed")
	}
	fmt.Printf("Batch query: total=%d, exists=%d, notexists=%d\n",
		resp.Total, resp.Exists, resp.NotExists)
	for _, r := range resp.Results {
		status := "EXISTS"
		if !r.Exists {
			status = "NOTEXISTS"
		}
		show := false
		switch *showOnly {
		case "all":
			show = true
		case "exists":
			show = r.Exists
		case "notexists":
			show = !r.Exists
		}
		if show {
			fmt.Printf("  %s: %s\n", status, r.Element)
		}
	}
	return nil
}

func cmdBatchDelete(args []string) error {
	fs := flag.NewFlagSet("batch-delete", flag.ExitOnError)
	fileFlag := fs.String("file", "", "File containing elements (one per line)")
	fs.Parse(args)
	var elements []string
	if *fileFlag != "" {
		var err error
		elements, err = readLinesFromFile(*fileFlag)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}
	} else if fs.NArg() > 0 {
		elements = fs.Args()
	} else {
		return fmt.Errorf("usage: client batch-delete -file <filename> or <elem1> <elem2> ...")
	}
	if len(elements) == 0 {
		return fmt.Errorf("no elements to delete")
	}
	var resp common.BatchDeleteResponse
	err := makeRequest("POST", "/batch/delete", common.BatchDeleteRequest{Elements: elements}, &resp)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("batch delete failed")
	}
	fmt.Printf("Batch delete: total=%d, removed=%d, notfound=%d\n",
		resp.Total, resp.Removed, resp.NotFound)
	for _, r := range resp.Results {
		status := "REMOVED"
		if !r.Removed {
			status = "NOTFOUND"
		}
		fmt.Printf("  %s: %s\n", status, r.Element)
	}
	return nil
}

func cmdDefaultConfig(_ []string) error {
	cfg := core.DefaultConfig()
	fmt.Println("Default configuration:")
	fmt.Printf("  Auto-expand: %v, max: %d, warn-factor: %.2f\n",
		cfg.AutoExpandOnFull, cfg.MaxAutoLayers, cfg.WarningFPRFactor)
	fmt.Printf("  Layers:\n")
	for i, l := range cfg.Layers {
		fmt.Printf("    Layer %d: capacity=%d, hash=%d, fpr=%.6f\n",
			i, l.Capacity, l.HashFunctions, l.TargetFPR)
	}
	return nil
}

func printUsage() {
	fmt.Println(`Layered Bloom Filter Client

Usage: client [global-options] <command> [command-options] [args]

Global options:
  -server <url>   Server base URL (default: http://localhost:8513)

Commands:
  insert <elem>            Insert an element
  query <elem>             Check if element exists
  delete <elem>            Delete an element
  stats                    Show filter statistics
  reset                    Reset the filter
  config                   Reconfigure filter (see -help)
  batch-insert             Insert multiple elements
  batch-query              Query multiple elements
  batch-delete             Delete multiple elements
  default-config           Show default configuration

Examples:
  client -server http://localhost:9090 insert "http://example.com"
  client query "test"
  client stats
  client batch-query -file urls.txt -show-only notexists`)
}

func main() {
	fs := flag.NewFlagSet("global", flag.ExitOnError)
	serverFlag := fs.String("server", "http://localhost:8513", "Server base URL")
	fs.Usage = printUsage
	fs.Parse(os.Args[1:])
	baseURL = *serverFlag
	args := fs.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}
	cmd := args[0]
	rest := args[1:]
	var err error
	switch cmd {
	case "insert":
		err = cmdInsert(rest)
	case "query":
		err = cmdQuery(rest)
	case "delete":
		err = cmdDelete(rest)
	case "stats":
		err = cmdStats(rest)
	case "reset":
		err = cmdReset(rest)
	case "config":
		err = cmdConfig(rest)
	case "batch-insert":
		err = cmdBatchInsert(rest)
	case "batch-query":
		err = cmdBatchQuery(rest)
	case "batch-delete":
		err = cmdBatchDelete(rest)
	case "default-config":
		err = cmdDefaultConfig(rest)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
