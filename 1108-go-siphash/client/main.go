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

	"github.com/siphash-service/common"
)

var serverURL string

func main() {
	urlFlag := flag.String("url", "http://localhost:8500", "SipHash server URL")
	flag.Parse()
	serverURL = strings.TrimRight(*urlFlag, "/")

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "hash":
		cmdHash(cmdArgs)
	case "put":
		cmdPut(cmdArgs)
	case "get":
		cmdGet(cmdArgs)
	case "delete":
		cmdDelete(cmdArgs)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage: siphash-client [options] <command> [arguments]

Commands:
  hash [-variant siphash-2-4|siphash-1-3] <data>
    Compute SipHash of the given data.

  put <key> <value>
    Store a key-value pair in the hash table.

  get <key>
    Retrieve a value from the hash table by key.

  delete <key>
    Delete a key-value pair from the hash table.

Options:
  -url string
    Server URL (default "http://localhost:8500")`)
}

func readStdin() []byte {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil
	}
	return data
}

func cmdHash(args []string) {
	fs := flag.NewFlagSet("hash", flag.ExitOnError)
	variantFlag := fs.String("variant", "siphash-2-4", "SipHash variant: siphash-2-4 or siphash-1-3")
	fs.Parse(args)

	var data []byte
	if fs.NArg() > 0 {
		data = []byte(fs.Arg(0))
	} else {
		data = readStdin()
		if len(data) == 0 {
			fmt.Println("Error: no data provided")
			os.Exit(1)
		}
	}

	req := common.HashRequest{
		Data:    data,
		Variant: common.Variant(*variantFlag),
	}

	var resp common.HashResponse
	if err := callHTTP("/hash", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Hash:  %d\n", resp.Hash)
	fmt.Printf("Hex:   %s\n", resp.Hex)
	fmt.Printf("Variant: %s\n", resp.Variant)
}

func cmdPut(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: siphash-client put <key> [value]")
		os.Exit(1)
	}

	key := []byte(args[0])
	var value []byte
	if len(args) > 1 {
		value = []byte(args[1])
	} else {
		value = readStdin()
	}

	req := common.PutRequest{Key: key, Value: value}
	var resp common.PutResponse
	if err := callHTTP("/put", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("OK")
}

func cmdGet(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: siphash-client get <key>")
		os.Exit(1)
	}

	key := []byte(args[0])
	req := common.GetRequest{Key: key}
	var resp common.GetResponse
	if err := callHTTP("/get", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Found {
		fmt.Println("Key not found")
		os.Exit(1)
	}

	fmt.Printf("%s\n", string(resp.Value))
}

func cmdDelete(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: siphash-client delete <key>")
		os.Exit(1)
	}

	key := []byte(args[0])
	req := common.DeleteRequest{Key: key}
	var resp common.DeleteResponse
	if err := callHTTP("/delete", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp.OK {
		fmt.Println("Deleted")
	} else {
		fmt.Println("Key not found")
	}
}

func callHTTP(path string, reqBody interface{}, respBody interface{}) error {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(serverURL+path, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	if respBody != nil {
		if err := json.Unmarshal(body, respBody); err != nil {
			return fmt.Errorf("failed to parse response: %v", err)
		}
	}

	return nil
}

func drainBody(body io.ReadCloser) {
	io.Copy(io.Discard, body)
	body.Close()
}
