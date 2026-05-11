package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"lsm-tree/common"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) post(path string, body interface{}, result interface{}) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	
	resp, err := c.http.Post(c.baseURL+path, "application/json", bytes.NewReader(buf))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) get(path string, params url.Values, result interface{}) error {
	url := c.baseURL + path
	if len(params) > 0 {
		url += "?" + params.Encode()
	}
	
	resp, err := c.http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) Put(key, value string) error {
	req := common.PutRequest{Key: key, Value: value}
	var resp common.PutResponse
	if err := c.post("/put", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *Client) Get(key string) (string, bool, error) {
	params := url.Values{}
	params.Set("key", key)
	
	var resp common.GetResponse
	if err := c.get("/get", params, &resp); err != nil {
		return "", false, err
	}
	if resp.Error != "" {
		return "", false, fmt.Errorf(resp.Error)
	}
	return resp.Value, resp.Exists, nil
}

func (c *Client) Delete(key string) error {
	req := common.DeleteRequest{Key: key}
	var resp common.DeleteResponse
	if err := c.post("/delete", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *Client) Range(start, end string) ([]common.KVPair, error) {
	params := url.Values{}
	if start != "" {
		params.Set("start", start)
	}
	if end != "" {
		params.Set("end", end)
	}
	
	var resp common.RangeResponse
	if err := c.get("/range", params, &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf(resp.Error)
	}
	return resp.Results, nil
}

func (c *Client) Compact() error {
	var resp common.CompactResponse
	if err := c.post("/compact", struct{}{}, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *Client) SwitchStrategy(strategy common.CompactionStrategy) error {
	req := common.SwitchStrategyRequest{Strategy: strategy}
	var resp common.SwitchStrategyResponse
	if err := c.post("/strategy", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *Client) Stats() (*common.StatsResponse, error) {
	var resp common.StatsResponse
	if err := c.get("/stats", url.Values{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Config() (*common.ConfigResponse, error) {
	var resp common.ConfigResponse
	if err := c.get("/config", url.Values{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) BatchPut(pairs []common.KVPair) (int, error) {
	req := common.BatchPutRequest{Pairs: pairs}
	var resp common.BatchPutResponse
	if err := c.post("/batch", req, &resp); err != nil {
		return 0, err
	}
	if resp.Error != "" {
		return resp.SuccessCount, fmt.Errorf(resp.Error)
	}
	return resp.SuccessCount, nil
}

func printUsage() {
	fmt.Println("LSM Tree Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client [flags] <command> [arguments]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -server string    Server URL (default: http://localhost:8514)")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  put <key> <value>                    Put a key-value pair")
	fmt.Println("  get <key>                            Get a value by key")
	fmt.Println("  delete <key>                         Delete a key")
	fmt.Println("  range <start> <end>                  Range query [start, end]")
	fmt.Println("  import <file.csv|file.json>          Import data from CSV or JSON")
	fmt.Println("  compact                              Trigger manual compaction")
	fmt.Println("  strategy <size-tiered|leveled>       Switch compaction strategy")
	fmt.Println("  stats                                Show SSTable and write amplification stats")
	fmt.Println("  config                               Show current configuration")
}

func main() {
	serverURL := flag.String("server", "http://localhost:8514", "Server URL")
	flag.Usage = printUsage
	flag.Parse()
	
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}
	
	client := NewClient(*serverURL)
	
	switch args[0] {
	case "put":
		if len(args) < 3 {
			fmt.Println("Usage: put <key> <value>")
			os.Exit(1)
		}
		key := args[1]
		value := strings.Join(args[2:], " ")
		if err := client.Put(key, value); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Put: %s -> %s\n", key, value)
		
	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: get <key>")
			os.Exit(1)
		}
		key := args[1]
		value, exists, err := client.Get(key)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if exists {
			fmt.Printf("%s\n", value)
		} else {
			fmt.Println("(not found)")
		}
		
	case "delete":
		if len(args) < 2 {
			fmt.Println("Usage: delete <key>")
			os.Exit(1)
		}
		key := args[1]
		if err := client.Delete(key); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Deleted: %s\n", key)
		
	case "range":
		start := ""
		end := ""
		if len(args) >= 2 {
			start = args[1]
		}
		if len(args) >= 3 {
			end = args[2]
		}
		results, err := client.Range(start, end)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if len(results) == 0 {
			fmt.Println("(no results)")
		} else {
			for _, pair := range results {
				fmt.Printf("%s -> %s\n", pair.Key, pair.Value)
			}
		}
		
	case "import":
		if len(args) < 2 {
			fmt.Println("Usage: import <file.csv|file.json>")
			os.Exit(1)
		}
		filename := args[1]
		pairs, err := readImportFile(filename)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Importing %d entries...\n", len(pairs))
		count, err := client.BatchPut(pairs)
		if err != nil {
			fmt.Printf("Import error (partial): %v\n", err)
		}
		fmt.Printf("Imported %d/%d entries\n", count, len(pairs))
		
	case "compact":
		fmt.Println("Triggering compaction...")
		if err := client.Compact(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Compaction completed")
		
	case "strategy":
		if len(args) < 2 {
			fmt.Println("Usage: strategy <size-tiered|leveled>")
			os.Exit(1)
		}
		strategy := common.CompactionStrategy(args[1])
		if strategy != common.SizeTieredStrategy && strategy != common.LeveledStrategy {
			fmt.Println("Invalid strategy. Use 'size-tiered' or 'leveled'")
			os.Exit(1)
		}
		fmt.Printf("Switching to %s strategy...\n", strategy)
		if err := client.SwitchStrategy(strategy); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Switched to %s strategy\n", strategy)
		
	case "stats":
		stats, err := client.Stats()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Strategy: %s\n", stats.Strategy)
		fmt.Printf("Write Amplification: %.2f\n", stats.WriteAmplification)
		fmt.Printf("User Writes: %d bytes\n", stats.UserWrites)
		fmt.Printf("Total Writes: %d bytes\n", stats.TotalWrites)
		fmt.Println()
		fmt.Println("Levels:")
		if len(stats.Levels) == 0 {
			fmt.Println("  (no SSTables)")
		} else {
			for _, level := range stats.Levels {
				fmt.Printf("  Level %d: %d files, %d bytes\n",
					level.Level, level.FileCount, level.TotalSize)
			}
		}
		
	case "config":
		config, err := client.Config()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("MemTable Size: %d bytes\n", config.Current.MemTableSize)
		fmt.Printf("Strategy: %s\n", config.Current.Strategy)
		fmt.Printf("Data Directory: %s\n", config.Current.DataDir)
		
	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func readImportFile(filename string) ([]common.KVPair, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		return readCSV(file)
	}
	return readJSON(file)
}

func readCSV(r io.Reader) ([]common.KVPair, error) {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	
	pairs := make([]common.KVPair, 0)
	for i, record := range records {
		if i == 0 && len(record) >= 2 &&
			(strings.EqualFold(record[0], "key") || strings.EqualFold(record[1], "value")) {
			continue
		}
		if len(record) >= 2 {
			pairs = append(pairs, common.KVPair{
				Key:   record[0],
				Value: record[1],
			})
		}
	}
	
	return pairs, nil
}

func readJSON(r io.Reader) ([]common.KVPair, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	
	var pairs []common.KVPair
	if err := json.Unmarshal(data, &pairs); err == nil {
		return pairs, nil
	}
	
	var mapData map[string]string
	if err := json.Unmarshal(data, &mapData); err != nil {
		return nil, err
	}
	
	pairs = make([]common.KVPair, 0, len(mapData))
	for k, v := range mapData {
		pairs = append(pairs, common.KVPair{Key: k, Value: v})
	}
	
	return pairs, nil
}
