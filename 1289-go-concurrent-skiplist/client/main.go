package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/example/concurrent-skiplist/common"
)

const DefaultServerAddr = "http://localhost:8516"

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Insert(key string, value interface{}) error {
	reqBody := common.InsertRequest{
		Key:   key,
		Value: value,
	}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := c.client.Post(c.baseURL+"/api/insert", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var respBody common.InsertResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("decode response failed: %v, body: %s", err, string(body))
	}

	if !respBody.Success {
		return fmt.Errorf("insert failed: %s", respBody.Message)
	}

	fmt.Printf("Inserted: %s = %v\n", key, value)
	return nil
}

func (c *Client) Get(key string) error {
	resp, err := c.client.Get(fmt.Sprintf("%s/api/get?key=%s", c.baseURL, key))
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var respBody common.GetResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("decode response failed: %v, body: %s", err, string(body))
	}

	if respBody.Success {
		fmt.Printf("Got: %s = %v\n", key, respBody.Value)
	} else {
		fmt.Printf("Key not found: %s\n", key)
	}
	return nil
}

func (c *Client) Delete(key string) error {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/delete?key=%s", c.baseURL, key), nil)
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var respBody common.DeleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("decode response failed: %v, body: %s", err, string(body))
	}

	if respBody.Success {
		fmt.Printf("Deleted: %s\n", key)
	} else {
		fmt.Printf("Key not found: %s\n", key)
	}
	return nil
}

func (c *Client) Range(start, end string) error {
	resp, err := c.client.Get(fmt.Sprintf("%s/api/range?start=%s&end=%s", c.baseURL, start, end))
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var respBody common.RangeQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("decode response failed: %v, body: %s", err, string(body))
	}

	if respBody.Success {
		fmt.Printf("Range [%s, %s] - %d results:\n", start, end, len(respBody.Data))
		for _, kv := range respBody.Data {
			fmt.Printf("  %s = %v\n", kv.Key, kv.Value)
		}
	} else {
		fmt.Printf("Range query failed: %s\n", respBody.Message)
	}
	return nil
}

func (c *Client) SwitchStrategy(strategy string) error {
	reqBody := common.SwitchStrategyRequest{
		Strategy: strategy,
	}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := c.client.Post(c.baseURL+"/api/strategy", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var respBody common.SwitchStrategyResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("decode response failed: %v, body: %s", err, string(body))
	}

	if respBody.Success {
		fmt.Println(respBody.Message)
	} else {
		return fmt.Errorf("switch strategy failed: %s", respBody.Message)
	}
	return nil
}

func (c *Client) GetStats() error {
	resp, err := c.client.Get(c.baseURL + "/api/stats")
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var respBody common.GetStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("decode response failed: %v, body: %s", err, string(body))
	}

	if respBody.Success {
		stats := respBody.Stats
		fmt.Println("=== Performance Statistics ===")
		fmt.Printf("Current Strategy: %s\n", stats.CurrentStrategy)
		fmt.Printf("Total Operations: %d\n", stats.TotalOperations)
		fmt.Printf("Success Operations: %d\n", stats.SuccessOperations)
		fmt.Printf("Failed Operations: %d\n", stats.FailedOperations)
		fmt.Printf("Conflict Count: %d\n", stats.ConflictCount)
		fmt.Printf("Average Wait Time: %.4f ms\n", stats.AvgWaitTime)
		fmt.Printf("Insert Count: %d\n", stats.InsertCount)
		fmt.Printf("Delete Count: %d\n", stats.DeleteCount)
		fmt.Printf("Get Count: %d\n", stats.GetCount)
		fmt.Printf("Range Count: %d\n", stats.RangeCount)
	} else {
		return fmt.Errorf("get stats failed: %s", respBody.Message)
	}
	return nil
}

func (c *Client) StressTest(operations, concurrency int) error {
	reqBody := common.StressTestRequest{
		Operations:  operations,
		Concurrency: concurrency,
	}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := c.client.Post(c.baseURL+"/api/stress", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var respBody common.StressTestResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("decode response failed: %v, body: %s", err, string(body))
	}

	if respBody.Success {
		fmt.Println("=== Stress Test Result ===")
		fmt.Printf("Operations: %d\n", respBody.Operations)
		fmt.Printf("Duration: %d ms\n", respBody.DurationMs)
		fmt.Printf("Throughput: %.2f ops/sec\n", respBody.Throughput)
	} else {
		return fmt.Errorf("stress test failed: %s", respBody.Message)
	}
	return nil
}

func (c *Client) BenchmarkStrategies(operations, concurrency int) error {
	fmt.Println("=== Benchmark: Global Lock Strategy ===")
	if err := c.SwitchStrategy("global_lock"); err != nil {
		return err
	}
	
	if err := c.StressTest(operations, concurrency); err != nil {
		return err
	}
	if err := c.GetStats(); err != nil {
		return err
	}

	fmt.Println("\n=== Benchmark: Node Level Lock Strategy ===")
	if err := c.SwitchStrategy("node_level_lock"); err != nil {
		return err
	}
	
	if err := c.StressTest(operations, concurrency); err != nil {
		return err
	}
	if err := c.GetStats(); err != nil {
		return err
	}

	return nil
}

func printUsage() {
	fmt.Println("Concurrent SkipList Client")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  client [flags] <command> [args...]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  insert <key> <value>    Insert or update a key-value pair")
	fmt.Println("  get <key>               Get value by key")
	fmt.Println("  delete <key>            Delete a key")
	fmt.Println("  range <start> <end>     Range query [start, end]")
	fmt.Println("  strategy <type>         Switch strategy: global_lock or node_level_lock")
	fmt.Println("  stats                   Get performance statistics")
	fmt.Println("  stress <ops> <conc>     Run stress test with N operations and C concurrency")
	fmt.Println("  benchmark <ops> <conc>  Compare both strategies")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  -server <url>           Server URL (default: http://localhost:8516)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  client insert key1 value1")
	fmt.Println("  client get key1")
	fmt.Println("  client delete key1")
	fmt.Println("  client range a z")
	fmt.Println("  client strategy node_level_lock")
	fmt.Println("  client stats")
	fmt.Println("  client stress 10000 20")
	fmt.Println("  client benchmark 10000 20")
}

func main() {
	serverFlag := flag.String("server", os.Getenv("SERVER_URL"), "Server URL")
	flag.Parse()

	if *serverFlag == "" {
		*serverFlag = DefaultServerAddr
	}

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		return
	}

	client := NewClient(*serverFlag)
	command := args[0]

	var err error
	switch strings.ToLower(command) {
	case "insert":
		if len(args) < 3 {
			err = fmt.Errorf("usage: insert <key> <value>")
		} else {
			err = client.Insert(args[1], args[2])
		}
	case "get":
		if len(args) < 2 {
			err = fmt.Errorf("usage: get <key>")
		} else {
			err = client.Get(args[1])
		}
	case "delete":
		if len(args) < 2 {
			err = fmt.Errorf("usage: delete <key>")
		} else {
			err = client.Delete(args[1])
		}
	case "range":
		if len(args) < 3 {
			err = fmt.Errorf("usage: range <start> <end>")
		} else {
			err = client.Range(args[1], args[2])
		}
	case "strategy":
		if len(args) < 2 {
			err = fmt.Errorf("usage: strategy <global_lock|node_level_lock>")
		} else {
			err = client.SwitchStrategy(args[1])
		}
	case "stats":
		err = client.GetStats()
	case "stress":
		operations := 10000
		concurrency := 10
		if len(args) >= 2 {
			operations, _ = strconv.Atoi(args[1])
		}
		if len(args) >= 3 {
			concurrency, _ = strconv.Atoi(args[2])
		}
		err = client.StressTest(operations, concurrency)
	case "benchmark":
		operations := 10000
		concurrency := 10
		if len(args) >= 2 {
			operations, _ = strconv.Atoi(args[1])
		}
		if len(args) >= 3 {
			concurrency, _ = strconv.Atoi(args[2])
		}
		err = client.BenchmarkStrategies(operations, concurrency)
	default:
		printUsage()
		return
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
