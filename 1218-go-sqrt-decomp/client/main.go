package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"sqrtdecomp/api"
)

const defaultServer = "http://localhost:8080"

func getServerURL() string {
	if url := os.Getenv("SERVER_URL"); url != "" {
		return url
	}
	return defaultServer
}

func httpPost(endpoint string, req interface{}, resp interface{}) error {
	serverURL := getServerURL()
	url := serverURL + endpoint

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpResp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return fmt.Errorf("server error: %s", string(respBody))
		}
		return fmt.Errorf("%s", errResp.Message)
	}

	if err := json.Unmarshal(respBody, resp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client create <size>               - Create array with given size")
	fmt.Println("  client rangeadd <l> <r> <val>      - Add val to elements [l, r]")
	fmt.Println("  client rangesum <l> <r>            - Get sum of elements [l, r]")
	fmt.Println("  client set <index> <val>           - Set value at index")
	fmt.Println("  client get <index>                 - Get value at index")
	fmt.Println("  client dump                        - Dump full array")
	fmt.Println("")
	fmt.Println("Environment:")
	fmt.Println("  SERVER_URL                         - Server URL (default: http://localhost:8080)")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := strings.ToLower(os.Args[1])

	switch cmd {
	case "create":
		if len(os.Args) != 3 {
			fmt.Println("Usage: client create <size>")
			os.Exit(1)
		}
		size, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Printf("Invalid size: %v\n", err)
			os.Exit(1)
		}

		req := api.CreateRequest{Size: size}
		var resp api.CreateResponse
		if err := httpPost("/create", req, &resp); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(resp.Message)

	case "rangeadd":
		if len(os.Args) != 6 {
			fmt.Println("Usage: client rangeadd <l> <r> <val>")
			os.Exit(1)
		}
		l, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Printf("Invalid l: %v\n", err)
			os.Exit(1)
		}
		r, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Printf("Invalid r: %v\n", err)
			os.Exit(1)
		}
		val, err := strconv.ParseInt(os.Args[4], 10, 64)
		if err != nil {
			fmt.Printf("Invalid val: %v\n", err)
			os.Exit(1)
		}

		req := api.RangeAddRequest{L: l, R: r, Val: val}
		var resp api.RangeAddResponse
		if err := httpPost("/rangeadd", req, &resp); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(resp.Message)

	case "rangesum":
		if len(os.Args) != 4 {
			fmt.Println("Usage: client rangesum <l> <r>")
			os.Exit(1)
		}
		l, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Printf("Invalid l: %v\n", err)
			os.Exit(1)
		}
		r, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Printf("Invalid r: %v\n", err)
			os.Exit(1)
		}

		req := api.RangeSumRequest{L: l, R: r}
		var resp api.RangeSumResponse
		if err := httpPost("/rangesum", req, &resp); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(resp.Sum)

	case "set":
		if len(os.Args) != 4 {
			fmt.Println("Usage: client set <index> <val>")
			os.Exit(1)
		}
		index, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Printf("Invalid index: %v\n", err)
			os.Exit(1)
		}
		val, err := strconv.ParseInt(os.Args[3], 10, 64)
		if err != nil {
			fmt.Printf("Invalid val: %v\n", err)
			os.Exit(1)
		}

		req := api.SetRequest{Index: index, Val: val}
		var resp api.SetResponse
		if err := httpPost("/set", req, &resp); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(resp.Message)

	case "get":
		if len(os.Args) != 3 {
			fmt.Println("Usage: client get <index>")
			os.Exit(1)
		}
		index, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Printf("Invalid index: %v\n", err)
			os.Exit(1)
		}

		req := api.GetRequest{Index: index}
		var resp api.GetResponse
		if err := httpPost("/get", req, &resp); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(resp.Value)

	case "dump":
		req := api.DumpRequest{}
		var resp api.DumpResponse
		if err := httpPost("/dump", req, &resp); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%v\n", resp.Data)

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
