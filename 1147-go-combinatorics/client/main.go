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

	"combinatorics/api"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	op := args[0]
	var req *api.CombinatoricsRequest
	var err error

	switch op {
	case "C":
		req, err = parseCombinationArgs(args[1:])
	case "Cmod":
		req, err = parseCombinationModArgs(args[1:])
	case "P":
		req, err = parsePermutationArgs(args[1:])
	case "Pmod":
		req, err = parsePermutationModArgs(args[1:])
	case "Pdup":
		req, err = parsePermutationDupArgs(args[1:], false)
	case "PdupMod":
		req, err = parsePermutationDupArgs(args[1:], true)
	default:
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		printUsage()
		os.Exit(1)
	}

	req.Operation = api.Operation(op)
	resp, err := sendRequest(*serverURL, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Println(resp.Result)
}

func printUsage() {
	fmt.Println("Usage: comb [command] [args]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  C n k         - Calculate combination C(n,k)")
	fmt.Println("  Cmod n k mod  - Calculate combination C(n,k) mod mod")
	fmt.Println("  P n k         - Calculate permutation P(n,k)")
	fmt.Println("  Pmod n k mod  - Calculate permutation P(n,k) mod mod")
	fmt.Println("  Pdup c1 c2... - Calculate permutation with duplicates")
	fmt.Println("  PdupMod mod c1 c2... - Calculate permutation with duplicates mod mod")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  -server URL   - Server URL (default: http://localhost:8080)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  comb C 100 50")
	fmt.Println("  comb P 10 3")
	fmt.Println("  comb Cmod 1000000 500000 998244353")
}

func parseCombinationArgs(args []string) (*api.CombinatoricsRequest, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("C command requires 2 arguments: n and k")
	}
	n, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid n: %v", err)
	}
	k, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid k: %v", err)
	}
	return &api.CombinatoricsRequest{N: n, K: k}, nil
}

func parseCombinationModArgs(args []string) (*api.CombinatoricsRequest, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("Cmod command requires 3 arguments: n, k and mod")
	}
	n, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid n: %v", err)
	}
	k, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid k: %v", err)
	}
	mod, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid mod: %v", err)
	}
	return &api.CombinatoricsRequest{N: n, K: k, Mod: mod}, nil
}

func parsePermutationArgs(args []string) (*api.CombinatoricsRequest, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("P command requires 2 arguments: n and k")
	}
	n, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid n: %v", err)
	}
	k, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid k: %v", err)
	}
	return &api.CombinatoricsRequest{N: n, K: k}, nil
}

func parsePermutationModArgs(args []string) (*api.CombinatoricsRequest, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("Pmod command requires 3 arguments: n, k and mod")
	}
	n, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid n: %v", err)
	}
	k, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid k: %v", err)
	}
	mod, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid mod: %v", err)
	}
	return &api.CombinatoricsRequest{N: n, K: k, Mod: mod}, nil
}

func parsePermutationDupArgs(args []string, withMod bool) (*api.CombinatoricsRequest, error) {
	if withMod {
		if len(args) < 2 {
			return nil, fmt.Errorf("PdupMod command requires at least 2 arguments: mod and at least one count")
		}
		mod, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid mod: %v", err)
		}
		counts, err := parseInt64List(args[1:])
		if err != nil {
			return nil, err
		}
		return &api.CombinatoricsRequest{Mod: mod, Counts: counts}, nil
	}
	if len(args) < 1 {
		return nil, fmt.Errorf("Pdup command requires at least 1 argument: at least one count")
	}
	counts, err := parseInt64List(args)
	if err != nil {
		return nil, err
	}
	return &api.CombinatoricsRequest{Counts: counts}, nil
}

func parseInt64List(args []string) ([]int64, error) {
	result := make([]int64, len(args))
	for i, arg := range args {
		n, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid argument %d: %v", i+1, err)
		}
		result[i] = n
	}
	return result, nil
}

func sendRequest(serverURL string, req *api.CombinatoricsRequest) (*api.CombinatoricsResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var result api.CombinatoricsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &result, nil
}
