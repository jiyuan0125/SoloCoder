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

	"sharded-map/common"
)

const DefaultServerURL = "http://localhost:8217"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	serverURL := getServerURL()

	var err error
	switch cmd {
	case "put":
		err = cmdPut(serverURL, args)
	case "get":
		err = cmdGet(serverURL, args)
	case "delete":
		err = cmdDelete(serverURL, args)
	case "keys":
		err = cmdKeys(serverURL, args)
	case "stats":
		err = cmdStats(serverURL, args)
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func getServerURL() string {
	if envURL := os.Getenv("SERVER_URL"); envURL != "" {
		return envURL
	}
	return DefaultServerURL
}

func printUsage() {
	fmt.Println("Usage: client <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  put <key> <value>    Store a key-value pair")
	fmt.Println("  get <key>            Retrieve a value by key")
	fmt.Println("  delete <key>         Delete a key-value pair")
	fmt.Println("  keys                 List all keys")
	fmt.Println("  stats                Show shard statistics")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  SERVER_URL           Server URL (default: http://localhost:8217)")
}

func cmdPut(serverURL string, args []string) error {
	fs := flag.NewFlagSet("put", flag.ExitOnError)
	fs.Parse(args)
	rest := fs.Args()

	if len(rest) != 2 {
		return fmt.Errorf("usage: client put <key> <value>")
	}

	key, value := rest[0], rest[1]
	req := common.PutRequest{Key: key, Value: value}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(serverURL+"/put", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	fmt.Printf("OK: stored %s = %s\n", key, value)
	return nil
}

func cmdGet(serverURL string, args []string) error {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	fs.Parse(args)
	rest := fs.Args()

	if len(rest) != 1 {
		return fmt.Errorf("usage: client get <key>")
	}

	key := rest[0]
	resp, err := http.Get(fmt.Sprintf("%s/get?key=%s", serverURL, key))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result common.GetResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Found {
		fmt.Printf("Not found: %s\n", key)
		return nil
	}

	fmt.Printf("%s\n", result.Value)
	return nil
}

func cmdDelete(serverURL string, args []string) error {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	fs.Parse(args)
	rest := fs.Args()

	if len(rest) != 1 {
		return fmt.Errorf("usage: client delete <key>")
	}

	key := rest[0]
	req := common.DeleteRequest{Key: key}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(serverURL+"/delete", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	fmt.Printf("OK: deleted %s\n", key)
	return nil
}

func cmdKeys(serverURL string, args []string) error {
	_ = args
	resp, err := http.Get(serverURL + "/keys")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result common.KeysResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if len(result.Keys) == 0 {
		fmt.Println("(no keys)")
		return nil
	}

	for _, k := range result.Keys {
		fmt.Println(k)
	}
	return nil
}

func cmdStats(serverURL string, args []string) error {
	_ = args
	resp, err := http.Get(serverURL + "/stats")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result common.StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	fmt.Printf("Total elements: %d\n", result.TotalCount)
	fmt.Printf("Shard count:    %d\n", result.ShardCount)
	fmt.Println()
	fmt.Println("Shard statistics:")
	fmt.Println("-----------------")
	for _, s := range result.Shards {
		marker := ""
		if s.IsUneven {
			marker = " [UNEVEN]"
		}
		fmt.Printf("  Shard %3d: %d%s\n", s.ShardID, s.Count, marker)
	}
	return nil
}
