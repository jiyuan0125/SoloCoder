package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/concurrent-sortedmap/internal/protocol"
)

const defaultServer = "http://localhost:8800"

type Client struct {
	server string
}

func NewClient(server string) *Client {
	if server == "" {
		server = defaultServer
	}
	return &Client{server: server}
}

func (c *Client) doPost(endpoint string, reqBody interface{}, respBody interface{}) error {
	data, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.server+endpoint, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(respBody)
}

func (c *Client) doGet(endpoint string, respBody interface{}) error {
	resp, err := http.Get(c.server + endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(respBody)
}

func (c *Client) Put(key, value string) error {
	req := protocol.PutRequest{Key: key, Value: value}
	var resp protocol.PutResponse
	if err := c.doPost("/put", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("put failed: %s", resp.Error)
	}
	return nil
}

func (c *Client) Get(key string) (string, bool, error) {
	var resp protocol.GetResponse
	if err := c.doGet("/get?key="+key, &resp); err != nil {
		return "", false, err
	}
	if !resp.Success {
		return "", false, fmt.Errorf("get failed: %s", resp.Error)
	}
	return resp.Value, resp.Exists, nil
}

func (c *Client) Delete(key string) (bool, error) {
	req := protocol.DeleteRequest{Key: key}
	var resp protocol.DeleteResponse
	if err := c.doPost("/delete", req, &resp); err != nil {
		return false, err
	}
	if !resp.Success {
		return false, fmt.Errorf("delete failed: %s", resp.Error)
	}
	return resp.Deleted, nil
}

func (c *Client) Range(start, end string) ([]protocol.KVPair, error) {
	var resp protocol.RangeResponse
	url := "/range"
	if start != "" || end != "" {
		url += "?"
		if start != "" {
			url += "start=" + start
		}
		if end != "" {
			if start != "" {
				url += "&"
			}
			url += "end=" + end
		}
	}
	if err := c.doGet(url, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("range failed: %s", resp.Error)
	}
	return resp.Pairs, nil
}

func (c *Client) Size() (int64, error) {
	var resp protocol.SizeResponse
	if err := c.doGet("/size", &resp); err != nil {
		return 0, err
	}
	if !resp.Success {
		return 0, fmt.Errorf("size failed: %s", resp.Error)
	}
	return resp.Size, nil
}

func cmdPut(client *Client, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: put <key> <value>")
		os.Exit(1)
	}
	if err := client.Put(args[0], args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	fmt.Println("OK")
}

func cmdGet(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: get <key>")
		os.Exit(1)
	}
	value, exists, err := client.Get(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	if exists {
		fmt.Println(value)
	} else {
		fmt.Println("(not found)")
	}
}

func cmdDelete(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: delete <key>")
		os.Exit(1)
	}
	deleted, err := client.Delete(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	if deleted {
		fmt.Println("deleted")
	} else {
		fmt.Println("not found")
	}
}

func cmdRange(client *Client, args []string) {
	var start, end string
	if len(args) >= 1 {
		start = args[0]
	}
	if len(args) >= 2 {
		end = args[1]
	}
	pairs, err := client.Range(start, end)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	for _, p := range pairs {
		fmt.Printf("%s\t%s\n", p.Key, p.Value)
	}
}

func cmdSize(client *Client, args []string) {
	size, err := client.Size()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	fmt.Println(size)
}

func cmdBench(client *Client, args []string) {
	fs := flag.NewFlagSet("bench", flag.ExitOnError)
	concurrency := fs.Int("c", 10, "number of concurrent workers")
	iterations := fs.Int("n", 1000, "number of operations per worker")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	var wg sync.WaitGroup
	start := time.Now()
	errCount := 0
	var errMu sync.Mutex

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < *iterations; j++ {
				key := fmt.Sprintf("bench_%d_%d", id, j)
				value := fmt.Sprintf("value_%d", j)
				if err := client.Put(key, value); err != nil {
					errMu.Lock()
					errCount++
					errMu.Unlock()
					continue
				}
				if _, _, err := client.Get(key); err != nil {
					errMu.Lock()
					errCount++
					errMu.Unlock()
				}
			}
		}(i)
	}
	wg.Wait()

	duration := time.Since(start)
	totalOps := int64(*concurrency) * int64(*iterations) * 2
	fmt.Printf("Benchmark completed in %v\n", duration)
	fmt.Printf("Total operations: %d\n", totalOps)
	fmt.Printf("Throughput: %.2f ops/sec\n", float64(totalOps)/duration.Seconds())
	if errCount > 0 {
		fmt.Printf("Errors: %d\n", errCount)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: client <command> [args]")
		fmt.Fprintln(os.Stderr, "Commands: put, get, delete, range, size, bench")
		os.Exit(1)
	}

	server := os.Getenv("SERVER_URL")
	if server == "" {
		server = defaultServer
	}

	client := NewClient(server)
	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "put":
		cmdPut(client, args)
	case "get":
		cmdGet(client, args)
	case "delete":
		cmdDelete(client, args)
	case "range":
		cmdRange(client, args)
	case "size":
		cmdSize(client, args)
	case "bench":
		cmdBench(client, args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}
