package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"lsm-storage/pkg/api"
)

var serverURL string

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		return
	}

	cmd := args[0]
	switch cmd {
	case "put":
		if len(args) < 3 {
			fmt.Println("usage: client put <key> <value>")
			return
		}
		putCmd(args[1], args[2])
	case "get":
		if len(args) < 2 {
			fmt.Println("usage: client get <key>")
			return
		}
		getCmd(args[1])
	case "delete":
		if len(args) < 2 {
			fmt.Println("usage: client delete <key>")
			return
		}
		deleteCmd(args[1])
	case "scan":
		var start, end string
		if len(args) >= 2 {
			start = args[1]
		}
		if len(args) >= 3 {
			end = args[2]
		}
		scanCmd(start, end)
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("usage: client [-server url] <command> [args]")
	fmt.Println("commands:")
	fmt.Println("  put <key> <value>")
	fmt.Println("  get <key>")
	fmt.Println("  delete <key>")
	fmt.Println("  scan [start] [end]")
}

func putCmd(key, value string) {
	req := api.PutRequest{Key: key, Value: value}
	data, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/put", "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var res api.PutResponse
	json.Unmarshal(body, &res)
	if !res.Success {
		fmt.Printf("error: %s\n", res.Error)
		os.Exit(1)
	}
	fmt.Println("ok")
}

func getCmd(key string) {
	req := api.GetRequest{Key: key}
	data, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/get", "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var res api.GetResponse
	json.Unmarshal(body, &res)
	if !res.Success {
		fmt.Printf("error: %s\n", res.Error)
		os.Exit(1)
	}
	if !res.Found {
		fmt.Println("not found")
		return
	}
	fmt.Println(res.Value)
}

func deleteCmd(key string) {
	req := api.DeleteRequest{Key: key}
	data, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/delete", "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var res api.DeleteResponse
	json.Unmarshal(body, &res)
	if !res.Success {
		fmt.Printf("error: %s\n", res.Error)
		os.Exit(1)
	}
	fmt.Println("ok")
}

func scanCmd(start, end string) {
	req := api.ScanRequest{Start: start, End: end}
	data, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/scan", "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var res api.ScanResponse
	json.Unmarshal(body, &res)
	if !res.Success {
		fmt.Printf("error: %s\n", res.Error)
		os.Exit(1)
	}
	for _, p := range res.Pairs {
		fmt.Printf("%s => %s\n", p.Key, p.Value)
	}
}
