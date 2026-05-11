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
	"strconv"
	"strings"

	"treap-demo/api"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var addr string
	flag.StringVar(&addr, "addr", "http://localhost:8200", "Server address")
	flag.CommandLine.Parse(os.Args[2:])

	cmd := strings.ToLower(os.Args[1])

	switch cmd {
	case "put":
		handlePut(addr, flag.CommandLine.Args())
	case "get":
		handleGet(addr, flag.CommandLine.Args())
	case "delete":
		handleDelete(addr, flag.CommandLine.Args())
	case "prev":
		handlePrev(addr, flag.CommandLine.Args())
	case "next":
		handleNext(addr, flag.CommandLine.Args())
	case "list":
		handleList(addr)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage: client [-addr http://host:port] <command> [args]

Commands:
  put <key> <value>   Insert or update key-value pair
  get <key>           Get value by key
  delete <key>        Delete key-value pair
  prev <key>          Find predecessor (max key < given key)
  next <key>          Find successor (min key > given key)
  list                List all key-value pairs in order`)
}

func handlePut(addr string, args []string) {
	if len(args) != 2 {
		fmt.Println("Usage: client put <key> <value>")
		os.Exit(1)
	}

	key, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Println("Invalid key:", args[0])
		os.Exit(1)
	}
	value, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		fmt.Println("Invalid value:", args[1])
		os.Exit(1)
	}

	body, _ := json.Marshal(api.PutRequest{Key: key, Value: value})
	resp, err := http.Post(addr+"/put", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result api.PutResponse
	_ = json.Unmarshal(respBody, &result)

	if result.Success {
		fmt.Println("OK")
	} else {
		fmt.Println("Error:", result.Error)
		os.Exit(1)
	}
}

func handleGet(addr string, args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: client get <key>")
		os.Exit(1)
	}

	key, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Println("Invalid key:", args[0])
		os.Exit(1)
	}

	resp, err := http.Get(addr + "/get?key=" + url.QueryEscape(strconv.FormatInt(key, 10)))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result api.GetResponse
	_ = json.Unmarshal(respBody, &result)

	if result.Success {
		fmt.Printf("Key: %d, Value: %d\n", result.Key, result.Value)
	} else {
		fmt.Println("Error:", result.Error)
		os.Exit(1)
	}
}

func handleDelete(addr string, args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: client delete <key>")
		os.Exit(1)
	}

	key, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Println("Invalid key:", args[0])
		os.Exit(1)
	}

	req, _ := http.NewRequest("GET", addr+"/delete?key="+url.QueryEscape(strconv.FormatInt(key, 10)), nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result api.DeleteResponse
	_ = json.Unmarshal(respBody, &result)

	if result.Success {
		fmt.Println("OK")
	} else {
		fmt.Println("Error:", result.Error)
		os.Exit(1)
	}
}

func handlePrev(addr string, args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: client prev <key>")
		os.Exit(1)
	}

	key, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Println("Invalid key:", args[0])
		os.Exit(1)
	}

	resp, err := http.Get(addr + "/prev?key=" + url.QueryEscape(strconv.FormatInt(key, 10)))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result api.PrevNextResponse
	_ = json.Unmarshal(respBody, &result)

	if result.Success {
		if result.Exists {
			fmt.Printf("Key: %d, Value: %d\n", result.Key, result.Value)
		} else {
			fmt.Println("(none)")
		}
	} else {
		fmt.Println("Error:", result.Error)
		os.Exit(1)
	}
}

func handleNext(addr string, args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: client next <key>")
		os.Exit(1)
	}

	key, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Println("Invalid key:", args[0])
		os.Exit(1)
	}

	resp, err := http.Get(addr + "/next?key=" + url.QueryEscape(strconv.FormatInt(key, 10)))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result api.PrevNextResponse
	_ = json.Unmarshal(respBody, &result)

	if result.Success {
		if result.Exists {
			fmt.Printf("Key: %d, Value: %d\n", result.Key, result.Value)
		} else {
			fmt.Println("(none)")
		}
	} else {
		fmt.Println("Error:", result.Error)
		os.Exit(1)
	}
}

func handleList(addr string) {
	resp, err := http.Get(addr + "/list")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result api.ListResponse
	_ = json.Unmarshal(respBody, &result)

	if result.Success {
		if len(result.Items) == 0 {
			fmt.Println("(empty)")
			return
		}
		for _, item := range result.Items {
			fmt.Printf("Key: %d, Value: %d\n", item.Key, item.Value)
		}
	} else {
		fmt.Println("Error:", result.Error)
		os.Exit(1)
	}
}
