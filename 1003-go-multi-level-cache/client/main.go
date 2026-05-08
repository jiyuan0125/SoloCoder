package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"multilevelcache/common"
)

var serverURL string

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Cache server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	cmd := strings.ToLower(args[0])
	switch cmd {
	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: cache get <key>")
			os.Exit(1)
		}
		handleGet(args[1])
	case "set":
		if len(args) < 3 {
			fmt.Println("Usage: cache set <key> <value> [memory-ttl] [file-ttl]")
			os.Exit(1)
		}
		memTTL := parseDuration(args, 3, 10*time.Minute)
		fileTTL := parseDuration(args, 4, 24*time.Hour)
		handleSet(args[1], args[2], memTTL, fileTTL)
	case "delete":
		if len(args) < 2 {
			fmt.Println("Usage: cache delete <key> or cache delete --prefix <prefix>")
			os.Exit(1)
		}
		if args[1] == "--prefix" {
			if len(args) < 3 {
				fmt.Println("Usage: cache delete --prefix <prefix>")
				os.Exit(1)
			}
			handleDeletePrefix(args[2])
		} else {
			handleDelete(args[1])
		}
	case "clear":
		handleClear()
	case "stats":
		handleStats()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func parseDuration(args []string, idx int, defaultVal time.Duration) time.Duration {
	if len(args) > idx {
		d, err := time.ParseDuration(args[idx])
		if err == nil {
			return d
		}
	}
	return defaultVal
}

func printUsage() {
	fmt.Println("Usage: cache <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  get <key>                 Get value by key")
	fmt.Println("  set <key> <value> [ttl]   Set value with optional TTL")
	fmt.Println("  delete <key>              Delete by key")
	fmt.Println("  delete --prefix <prefix>  Delete by prefix")
	fmt.Println("  clear                     Clear all cache")
	fmt.Println("  stats                     Show cache statistics")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --server <url>            Cache server URL (default: http://localhost:8080)")
}

func handleGet(key string) {
	req := common.GetRequest{Key: key}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	respBody, err := httpPost(serverURL+"/get", body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var resp common.GetResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if !resp.Found {
		fmt.Printf("Key '%s' not found\n", key)
		return
	}

	valueJSON, _ := json.MarshalIndent(resp.Value, "", "  ")
	fmt.Printf("Key: %s\n", resp.Key)
	fmt.Printf("Value: %s\n", valueJSON)
	if !resp.Expire.IsZero() {
		fmt.Printf("Expire: %s\n", resp.Expire.Format(time.RFC3339))
	}
}

func handleSet(key, value string, memTTL, fileTTL time.Duration) {
	var val interface{}
	if err := json.Unmarshal([]byte(value), &val); err != nil {
		val = value
	}

	req := common.SetRequest{
		Key:       key,
		Value:     val,
		MemoryTTL: memTTL,
		FileTTL:   fileTTL,
	}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	respBody, err := httpPost(serverURL+"/set", body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var resp common.SetResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Set key '%s' successfully\n", key)
	} else {
		fmt.Printf("Failed to set key '%s'\n", key)
	}
}

func handleDelete(key string) {
	req := common.DeleteRequest{Key: key}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	respBody, err := httpPost(serverURL+"/delete", body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var resp common.DeleteResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if resp.Deleted > 0 {
		fmt.Printf("Deleted %d entry(ies)\n", resp.Deleted)
	} else {
		fmt.Printf("Key '%s' not found\n", key)
	}
}

func handleDeletePrefix(prefix string) {
	req := common.DeleteRequest{Prefix: prefix}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	respBody, err := httpPost(serverURL+"/delete", body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var resp common.DeleteResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Deleted %d entry(ies) with prefix '%s'\n", resp.Deleted, prefix)
}

func handleClear() {
	req := common.ClearRequest{}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	respBody, err := httpPost(serverURL+"/clear", body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var resp common.ClearResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Println("Cache cleared successfully")
	} else {
		fmt.Println("Failed to clear cache")
	}
}

func handleStats() {
	respBody, err := httpPost(serverURL+"/stats", nil)
	if err != nil {
		respBody, err = httpGet(serverURL + "/stats")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	var resp common.StatsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Cache Statistics ===")
	fmt.Printf("Memory Hits:      %d\n", resp.MemoryHits)
	fmt.Printf("Memory Misses:    %d\n", resp.MemoryMisses)
	fmt.Printf("File Hits:        %d\n", resp.FileHits)
	fmt.Printf("File Misses:      %d\n", resp.FileMisses)
	fmt.Printf("Hit Rate:         %.2f%%\n", resp.HitRate*100)
	fmt.Println()
	fmt.Printf("Memory Entries:   %d / %d (%.1f%%)\n",
		resp.MemoryCount, resp.MemoryCapacity, resp.MemoryUsage*100)
	fmt.Printf("File Entries:     %d\n", resp.FileCount)
}

func httpPost(url string, body []byte) ([]byte, error) {
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: %s - %s", resp.Status, string(respBody))
	}

	return respBody, nil
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: %s - %s", resp.Status, string(respBody))
	}

	return respBody, nil
}
