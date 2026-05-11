package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"splaytree/common"
	"strconv"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

func (c *Client) doPost(endpoint string, reqBody interface{}, respBody interface{}) error {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + endpoint
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	if respBody != nil && len(body) > 0 {
		if err := json.Unmarshal(body, respBody); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *Client) Put(key, value int64) error {
	req := common.PutRequest{Key: key, Value: value}
	var resp common.PutResponse
	if err := c.doPost("/put", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("put failed: %s", resp.Message)
	}
	fmt.Printf("PUT success: key=%d, value=%d\n", key, value)
	return nil
}

func (c *Client) Get(key int64) error {
	req := common.GetRequest{Key: key}
	var resp common.GetResponse
	if err := c.doPost("/get", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("get failed: %s", resp.Message)
	}
	fmt.Printf("GET success: key=%d, value=%d\n", key, resp.Value)
	return nil
}

func (c *Client) Delete(key int64) error {
	req := common.DeleteRequest{Key: key}
	var resp common.DeleteResponse
	if err := c.doPost("/delete", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("delete failed: %s", resp.Message)
	}
	fmt.Printf("DELETE success: key=%d\n", key)
	return nil
}

func (c *Client) Range(l, r int64) error {
	req := common.RangeQueryRequest{Left: l, Right: r}
	var resp common.RangeQueryResponse
	if err := c.doPost("/range/query", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("range query failed: %s", resp.Message)
	}
	fmt.Printf("RANGE [%d, %d]:\n", l, r)
	fmt.Printf("  Sum: %d\n", resp.Sum)
	fmt.Printf("  Keys: %v\n", resp.Keys)
	fmt.Printf("  Values: %v\n", resp.Values)
	return nil
}

func printUsage() {
	fmt.Println("Splay Tree Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client [--server <url>] <command> [args...]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  put <key> <value>    Insert or update a key-value pair")
	fmt.Println("  get <key>            Get value by key")
	fmt.Println("  delete <key>         Delete a key")
	fmt.Println("  range <l> <r>        Query keys in range [l, r]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --server <url>       Server URL (default: http://localhost:8080)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client put 5 100")
	fmt.Println("  client get 5")
	fmt.Println("  client range 1 10")
	fmt.Println("  client delete 5")
}

func main() {
	var serverURL string
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Server URL")
	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(serverURL)
	command := args[0]

	var err error
	switch command {
	case "put":
		if len(args) != 3 {
			fmt.Println("Usage: put <key> <value>")
			os.Exit(1)
		}
		key, err1 := strconv.ParseInt(args[1], 10, 64)
		value, err2 := strconv.ParseInt(args[2], 10, 64)
		if err1 != nil || err2 != nil {
			fmt.Println("Invalid arguments: key and value must be integers")
			os.Exit(1)
		}
		err = client.Put(key, value)

	case "get":
		if len(args) != 2 {
			fmt.Println("Usage: get <key>")
			os.Exit(1)
		}
		key, err1 := strconv.ParseInt(args[1], 10, 64)
		if err1 != nil {
			fmt.Println("Invalid argument: key must be an integer")
			os.Exit(1)
		}
		err = client.Get(key)

	case "delete":
		if len(args) != 2 {
			fmt.Println("Usage: delete <key>")
			os.Exit(1)
		}
		key, err1 := strconv.ParseInt(args[1], 10, 64)
		if err1 != nil {
			fmt.Println("Invalid argument: key must be an integer")
			os.Exit(1)
		}
		err = client.Delete(key)

	case "range":
		if len(args) != 3 {
			fmt.Println("Usage: range <l> <r>")
			os.Exit(1)
		}
		l, err1 := strconv.ParseInt(args[1], 10, 64)
		r, err2 := strconv.ParseInt(args[2], 10, 64)
		if err1 != nil || err2 != nil {
			fmt.Println("Invalid arguments: l and r must be integers")
			os.Exit(1)
		}
		err = client.Range(l, r)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
