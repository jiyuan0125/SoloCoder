package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"radixsort/pkg/api"
)

var (
	serverURL   string
	inputFile   string
	outputFile  string
	radix       int
	getStats    bool
	getRadix    bool
	setRadix    bool
	forceStdout bool
)

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:8511", "radixsort server URL")
	flag.StringVar(&inputFile, "in", "", "input file path (one integer per line or JSON array)")
	flag.StringVar(&outputFile, "out", "", "output file path")
	flag.IntVar(&radix, "radix", 0, "set sort radix (power of 2: 2,4,8,16,32,64,128,256)")
	flag.BoolVar(&getStats, "stats", false, "get last sort stats")
	flag.BoolVar(&getRadix, "get-radix", false, "get current radix")
	flag.BoolVar(&setRadix, "set-radix", false, "set radix using -radix flag")
	flag.BoolVar(&forceStdout, "v", false, "verbose output to stdout")
	flag.Parse()

	if setRadix {
		if err := doSetRadix(); err != nil {
			fmt.Fprintf(os.Stderr, "error setting radix: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if getRadix {
		if err := doGetRadix(); err != nil {
			fmt.Fprintf(os.Stderr, "error getting radix: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if getStats {
		if err := doGetStats(); err != nil {
			fmt.Fprintf(os.Stderr, "error getting stats: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if inputFile == "" && flag.NArg() == 0 {
		showHelp()
		os.Exit(0)
	}

	var data []int64
	var err error
	if inputFile != "" {
		data, err = readInputFile(inputFile)
	} else {
		data, err = parseArgs(flag.Args())
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		os.Exit(1)
	}

	resp, err := doSort(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error sorting: %v\n", err)
		os.Exit(1)
	}

	if forceStdout {
		for _, v := range resp.Data {
			fmt.Println(v)
		}
	}

	if outputFile != "" {
		if err := writeOutputFile(outputFile, resp.Data); err != nil {
			fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
			os.Exit(1)
		}
		if !forceStdout {
			fmt.Printf("sorted %d elements; passes=%d ops=%d radix=%d; written to %s\n",
				len(resp.Data), resp.Stats.Passes, resp.Stats.TotalOperations, resp.Stats.Radix, outputFile)
		}
	} else if !forceStdout {
		fmt.Printf("sorted %d elements; passes=%d ops=%d radix=%d\n",
			len(resp.Data), resp.Stats.Passes, resp.Stats.TotalOperations, resp.Stats.Radix)
		for _, v := range resp.Data {
			fmt.Println(v)
		}
	}
}

func showHelp() {
	fmt.Println("radixsort client - command-line client for radix sort service")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client [flags] <numbers...>")
	fmt.Println("  client -in <input> [-out <output>] [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	flag.PrintDefaults()
}

func readInputFile(path string) ([]int64, error) {
	body, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s := strings.TrimSpace(string(body))
	if strings.HasPrefix(s, "[") {
		var arr []int64
		if err := json.Unmarshal([]byte(s), &arr); err == nil {
			return arr, nil
		}
	}
	lines := strings.Split(s, "\n")
	var result []int64
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, ",") {
			parts := strings.Split(line, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				n, err := strconv.ParseInt(p, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid number: %s", p)
				}
				result = append(result, n)
			}
			continue
		}
		n, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", line)
		}
		result = append(result, n)
	}
	return result, nil
}

func parseArgs(args []string) ([]int64, error) {
	var result []int64
	for _, arg := range args {
		n, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", arg)
		}
		result = append(result, n)
	}
	return result, nil
}

func writeOutputFile(path string, data []int64) error {
	var sb strings.Builder
	for _, v := range data {
		sb.WriteString(strconv.FormatInt(v, 10))
		sb.WriteByte('\n')
	}
	return ioutil.WriteFile(path, []byte(sb.String()), 0644)
}

func doSort(data []int64) (*api.SortResponse, error) {
	reqBody, err := json.Marshal(api.SortRequest{Data: data})
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(serverURL+"/api/sort", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result api.SortResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("server error: %s", result.Error)
	}
	return &result, nil
}

func doSetRadix() error {
	reqBody, err := json.Marshal(api.RadixRequest{Radix: radix})
	if err != nil {
		return err
	}
	resp, err := http.Post(serverURL+"/api/radix", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var result api.RadixResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("server error: %s", result.Error)
	}
	fmt.Printf("radix set to %d\n", result.Radix)
	return nil
}

func doGetRadix() error {
	resp, err := http.Get(serverURL + "/api/radix")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var result api.RadixResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("server error: %s", result.Error)
	}
	fmt.Printf("current radix: %d\n", result.Radix)
	return nil
}

func doGetStats() error {
	resp, err := http.Get(serverURL + "/api/stats")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var result api.StatsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("server error: %s", result.Error)
	}
	fmt.Printf("last sort stats:\n")
	fmt.Printf("  array_size: %d\n", result.Stats.ArraySize)
	fmt.Printf("  radix: %d\n", result.Stats.Radix)
	fmt.Printf("  passes: %d\n", result.Stats.Passes)
	fmt.Printf("  total_operations: %d\n", result.Stats.TotalOperations)
	return nil
}
