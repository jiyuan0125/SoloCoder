package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/example/bloomfilter/common"
)

var serverURL string

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Bloom filter server URL")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	var err error
	switch cmd {
	case "add":
		err = cmdAdd(args)
	case "check":
		err = cmdCheck(args)
	case "stats":
		err = cmdStats()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Bloom Filter Client

Usage:
  client [flags] command [arguments]

Commands:
  add <item>       Add an item to the bloom filter
  check <item>     Check if an item exists
  stats            Get bloom filter statistics

Flags:
  -server string   Server URL (default "http://localhost:8080")
`)
}

func httpPost(endpoint string, body interface{}, resp interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	res, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("server returned %d: %s", res.StatusCode, string(errBody))
	}

	if resp != nil {
		return json.NewDecoder(res.Body).Decode(resp)
	}
	return nil
}

func httpGet(endpoint string, resp interface{}) error {
	res, err := http.Get(endpoint)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("server returned %d: %s", res.StatusCode, string(errBody))
	}

	return json.NewDecoder(res.Body).Decode(resp)
}

func cmdAdd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: add <item>")
	}

	item := strings.Join(args, " ")
	req := common.AddRequest{Item: item}
	endpoint := strings.TrimRight(serverURL, "/") + "/bloom/add"

	if err := httpPost(endpoint, req, nil); err != nil {
		return err
	}

	fmt.Printf("Added: %s\n", item)
	return nil
}

func cmdCheck(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: check <item>")
	}

	item := strings.Join(args, " ")
	endpoint := strings.TrimRight(serverURL, "/") + "/bloom/check?item=" + url.QueryEscape(item)

	var resp common.CheckResponse
	if err := httpGet(endpoint, &resp); err != nil {
		return err
	}

	if resp.Exists {
		fmt.Printf("Item exists (possibly false positive): %s\n", item)
	} else {
		fmt.Printf("Item definitely does not exist: %s\n", item)
	}
	return nil
}

func cmdStats() error {
	endpoint := strings.TrimRight(serverURL, "/") + "/bloom/stats"

	var resp common.StatsResponse
	if err := httpGet(endpoint, &resp); err != nil {
		return err
	}

	fmt.Printf("Bloom Filter Statistics:\n")
	fmt.Printf("  Element Count: %d\n", resp.Count)
	fmt.Printf("  Capacity:      %d\n", resp.Capacity)
	fmt.Printf("  Current FPR:   %.6f (%.2f%%)\n", resp.CurrentFPR, resp.CurrentFPR*100)
	fmt.Printf("  Bit Usage:     %.2f%%\n", resp.BitUsage*100)
	return nil
}
