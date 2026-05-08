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
	"strings"

	"go-bitmap-index/api"
)

const defaultServer = "http://localhost:8080"

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return &Client{baseURL: baseURL}
}

func (c *Client) request(method, path string, body interface{}, resp interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errResp api.ErrorResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("server error: %d", httpResp.StatusCode)
		}
		return fmt.Errorf("%s", errResp.Error)
	}

	if resp != nil {
		return json.NewDecoder(httpResp.Body).Decode(resp)
	}
	return nil
}

func (c *Client) CreateIndex(name string, size uint64) error {
	return c.request("POST", "/index/create", &api.CreateIndexRequest{Name: name, Size: size}, nil)
}

func (c *Client) DeleteIndex(name string) error {
	return c.request("POST", "/index/delete", &api.DeleteIndexRequest{Name: name}, nil)
}

func (c *Client) ListIndices() ([]api.IndexInfo, error) {
	var resp api.ListIndicesResponse
	if err := c.request("POST", "/index/list", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Indices, nil
}

func (c *Client) SetBit(index string, bit uint64) error {
	return c.request("POST", "/bit/set", &api.SetBitRequest{Index: index, Bit: bit}, nil)
}

func (c *Client) ClearBit(index string, bit uint64) error {
	return c.request("POST", "/bit/clear", &api.ClearBitRequest{Index: index, Bit: bit}, nil)
}

func (c *Client) GetBit(index string, bit uint64) (bool, error) {
	var resp api.GetBitResponse
	if err := c.request("POST", "/bit/get", &api.GetBitRequest{Index: index, Bit: bit}, &resp); err != nil {
		return false, err
	}
	return resp.Value, nil
}

func (c *Client) Count(index string) (uint64, error) {
	var resp api.CountResponse
	if err := c.request("POST", "/count", &api.CountRequest{Index: index}, &resp); err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func (c *Client) Indices(index string) ([]uint64, error) {
	var resp api.IndicesResponse
	if err := c.request("POST", "/indices", &api.IndicesRequest{Index: index}, &resp); err != nil {
		return nil, err
	}
	return resp.Indices, nil
}

func (c *Client) Operation(op api.Operation, indices []string, result string) (uint64, error) {
	var resp api.OperationResponse
	if err := c.request("POST", "/op", &api.OperationRequest{Operation: op, Indices: indices, Result: result}, &resp); err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: bitmap-cli [flags] command [arguments]

Flags:
  -server <url>    Server URL (default: %s)

Commands:
  create <name> <size>           Create a new bitmap index
  delete <name>                  Delete a bitmap index
  list                           List all indices
  set <index> <bit>              Set a bit
  clear <index> <bit>            Clear a bit
  get <index> <bit>              Get a bit value
  count <index>                  Count set bits
  indices <index>                List indices of set bits
  and <indices...>               AND operation
  or <indices...>                OR operation
  xor <indices...>               XOR operation
  not <index>                    NOT operation
  andnot <index1> <index2>       ANDNOT operation
  op and|or|xor|not|andnot <indices...> [--result <name>]
                                 Generic operation with optional result storage
`, defaultServer)
}

func main() {
	server := flag.String("server", defaultServer, "server URL")
	result := flag.String("result", "", "store result to a new index")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	client := NewClient(*server)

	var err error
	switch cmd {
	case "create":
		if len(args) < 2 {
			usage()
			os.Exit(1)
		}
		size, e := strconv.ParseUint(args[1], 10, 64)
		if e != nil {
			fmt.Fprintf(os.Stderr, "Invalid size: %v\n", e)
			os.Exit(1)
		}
		err = client.CreateIndex(args[0], size)
	case "delete":
		if len(args) < 1 {
			usage()
			os.Exit(1)
		}
		err = client.DeleteIndex(args[0])
	case "list":
		var indices []api.IndexInfo
		indices, err = client.ListIndices()
		if err == nil {
			fmt.Printf("Indices (%d):\n", len(indices))
			for _, idx := range indices {
				fmt.Printf("  %s: size=%d, count=%d\n", idx.Name, idx.Size, idx.Count)
			}
			os.Exit(0)
		}
	case "set":
		if len(args) < 2 {
			usage()
			os.Exit(1)
		}
		bit, e := strconv.ParseUint(args[1], 10, 64)
		if e != nil {
			fmt.Fprintf(os.Stderr, "Invalid bit: %v\n", e)
			os.Exit(1)
		}
		err = client.SetBit(args[0], bit)
	case "clear":
		if len(args) < 2 {
			usage()
			os.Exit(1)
		}
		bit, e := strconv.ParseUint(args[1], 10, 64)
		if e != nil {
			fmt.Fprintf(os.Stderr, "Invalid bit: %v\n", e)
			os.Exit(1)
		}
		err = client.ClearBit(args[0], bit)
	case "get":
		if len(args) < 2 {
			usage()
			os.Exit(1)
		}
		bit, e := strconv.ParseUint(args[1], 10, 64)
		if e != nil {
			fmt.Fprintf(os.Stderr, "Invalid bit: %v\n", e)
			os.Exit(1)
		}
		var value bool
		value, err = client.GetBit(args[0], bit)
		if err == nil {
			fmt.Println(value)
			os.Exit(0)
		}
	case "count":
		if len(args) < 1 {
			usage()
			os.Exit(1)
		}
		var count uint64
		count, err = client.Count(args[0])
		if err == nil {
			fmt.Println(count)
			os.Exit(0)
		}
	case "indices":
		if len(args) < 1 {
			usage()
			os.Exit(1)
		}
		var bits []uint64
		bits, err = client.Indices(args[0])
		if err == nil {
			for _, b := range bits {
				fmt.Println(b)
			}
			os.Exit(0)
		}
	case "and", "or", "xor", "not", "andnot":
		if len(args) < 1 {
			usage()
			os.Exit(1)
		}
		var op api.Operation
		switch cmd {
		case "and":
			op = api.OpAnd
		case "or":
			op = api.OpOr
		case "xor":
			op = api.OpXor
		case "not":
			op = api.OpNot
		case "andnot":
			op = api.OpAndNot
		}
		var count uint64
		count, err = client.Operation(op, args, *result)
		if err == nil {
			fmt.Println(count)
			os.Exit(0)
		}
	case "op":
		if len(args) < 2 {
			usage()
			os.Exit(1)
		}
		var op api.Operation
		switch strings.ToLower(args[0]) {
		case "and":
			op = api.OpAnd
		case "or":
			op = api.OpOr
		case "xor":
			op = api.OpXor
		case "not":
			op = api.OpNot
		case "andnot":
			op = api.OpAndNot
		default:
			fmt.Fprintf(os.Stderr, "Unknown operation: %s\n", args[0])
			os.Exit(1)
		}
		var count uint64
		count, err = client.Operation(op, args[1:], *result)
		if err == nil {
			fmt.Println(count)
			os.Exit(0)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("OK")
}
