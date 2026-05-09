package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"bloomfilter/api"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverAddr := flag.String("server", "http://localhost:8080", "Server address")
	flag.CommandLine.Parse(os.Args[2:])

	cmd := os.Args[1]

	switch cmd {
	case "add":
		if flag.NArg() < 1 {
			fmt.Println("Usage: client add <item>")
			os.Exit(1)
		}
		item := flag.Arg(0)
		if err := handleAdd(*serverAddr, item); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "delete":
		if flag.NArg() < 1 {
			fmt.Println("Usage: client delete <item>")
			os.Exit(1)
		}
		item := flag.Arg(0)
		if err := handleDelete(*serverAddr, item); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "query":
		if flag.NArg() < 1 {
			fmt.Println("Usage: client query <item>")
			os.Exit(1)
		}
		item := flag.Arg(0)
		if err := handleQuery(*serverAddr, item); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "info":
		if err := handleInfo(*serverAddr); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Counting Bloom Filter Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client [-server address] <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  add <item>      Add an item to the filter")
	fmt.Println("  delete <item>     Delete an item from the filter")
	fmt.Println("  query <item>      Query if an item may exist")
	fmt.Println("  info              Show filter information")
}

func handleAdd(server, item string) error {
	reqBody := api.AddRequest{Item: item}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(server+"/add", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return handleError(resp.Body)
	}

	var result api.AddResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Success {
		fmt.Printf("Item '%s' added successfully\n", item)
	}

	return nil
}

func handleDelete(server, item string) error {
	reqBody := api.DeleteRequest{Item: item}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(server+"/delete", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return handleError(resp.Body)
	}

	var result api.DeleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Success {
		fmt.Printf("Item '%s' deleted (if existed)\n", item)
	}

	return nil
}

func handleQuery(server, item string) error {
	reqBody := api.QueryRequest{Item: item}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(server+"/query", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return handleError(resp.Body)
	}

	var result api.QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.MayContain {
		fmt.Printf("Item '%s' may exist in the filter\n", item)
	} else {
		fmt.Printf("Item '%s' definitely not in the filter\n", item)
	}

	return nil
}

func handleInfo(server string) error {
	resp, err := http.Get(server + "/info")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return handleError(resp.Body)
	}

	var info api.InfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return err
	}

	fmt.Println("Bloom Filter Information:")
	fmt.Printf("  Hash Functions:    %d\n", info.NumHashes)
	fmt.Printf("  Counter Bits:       %d\n", info.CounterBits)
	fmt.Printf("  Capacity:          %d\n", info.Capacity)
	fmt.Printf("  Inserted:          %d\n", info.Inserted)
	fmt.Printf("  Non-zero Counters: %d\n", info.NonZeroCounters)
	fmt.Printf("  Overflow Count:    %d\n", info.OverflowCount)

	return nil
}

func handleError(body io.Reader) error {
	var errResp api.ErrorResponse
	if err := json.NewDecoder(body).Decode(&errResp); err != nil {
		return fmt.Errorf("request failed with unknown error")
	}
	return fmt.Errorf("%s", errResp.Error)
}
