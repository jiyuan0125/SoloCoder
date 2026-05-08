package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"ringbuffer/common"
)

const defaultServer = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	server := os.Getenv("RINGBUFFER_SERVER")
	if server == "" {
		server = defaultServer
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "create":
		err = cmdCreate(server, args)
	case "write":
		err = cmdWrite(server, args)
	case "read":
		err = cmdRead(server, args)
	case "status":
		err = cmdStatus(server, args)
	case "close":
		err = cmdClose(server, args)
	case "list":
		err = cmdList(server, args)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("RingBuffer Client - A command line tool for ring buffer server")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create <capacity>           Create a new ring buffer")
	fmt.Println("  write <id> <data|file>      Write data to buffer (use @file to read from file)")
	fmt.Println("  read <id> <length> [file]   Read data from buffer (save to file if specified)")
	fmt.Println("  status <id>                 Check buffer status")
	fmt.Println("  close <id>                  Close buffer")
	fmt.Println("  list                        List all buffers")
	fmt.Println("  help                        Show this help")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  RINGBUFFER_SERVER           Server URL (default: http://localhost:8080)")
}

func cmdCreate(server string, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: client create <capacity>")
	}

	capacity, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid capacity: %v", err)
	}

	req := common.CreateRequest{Capacity: capacity}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(server+"/create", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", string(data))
	}

	var result common.CreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Error != "" {
		return fmt.Errorf(result.Error)
	}

	fmt.Printf("Created buffer: %s (capacity: %d)\n", result.ID, result.Capacity)
	return nil
}

func cmdWrite(server string, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: client write <id> <data|@file>")
	}

	id := args[0]
	dataStr := args[1]

	var data []byte
	if strings.HasPrefix(dataStr, "@") {
		filename := dataStr[1:]
		var err error
		data, err = ioutil.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("read file error: %v", err)
		}
	} else {
		data = []byte(dataStr)
	}

	timeout := int64(-1)
	if len(args) >= 3 {
		t, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid timeout: %v", err)
		}
		timeout = t
	}

	req := common.WriteRequest{
		ID:      id,
		Data:    data,
		Timeout: timeout,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(server+"/write", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", string(data))
	}

	var result common.WriteResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Error != "" {
		return fmt.Errorf(result.Error)
	}

	fmt.Printf("Written %d bytes\n", result.Written)
	return nil
}

func cmdRead(server string, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: client read <id> <length> [file] [timeout]")
	}

	id := args[0]
	length, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid length: %v", err)
	}

	var outFile string
	if len(args) >= 3 {
		outFile = args[2]
	}

	timeout := int64(-1)
	if len(args) >= 4 {
		t, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid timeout: %v", err)
		}
		timeout = t
	}

	req := common.ReadRequest{
		ID:      id,
		Length:  length,
		Timeout: timeout,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(server+"/read", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", string(data))
	}

	var result common.ReadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Error != "" {
		return fmt.Errorf(result.Error)
	}

	if result.EOF && result.Read == 0 {
		fmt.Println("EOF")
		return nil
	}

	if outFile != "" {
		if err := ioutil.WriteFile(outFile, result.Data, 0644); err != nil {
			return fmt.Errorf("write file error: %v", err)
		}
		fmt.Printf("Read %d bytes, saved to %s\n", result.Read, outFile)
	} else {
		fmt.Printf("Read %d bytes: %s\n", result.Read, string(result.Data))
	}

	return nil
}

func cmdStatus(server string, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: client status <id>")
	}

	id := args[0]
	req := common.StatusRequest{ID: id}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(server+"/status", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", string(data))
	}

	var result common.StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Error != "" {
		return fmt.Errorf(result.Error)
	}

	fmt.Printf("Buffer: %s\n", result.ID)
	fmt.Printf("  Capacity: %d\n", result.Capacity)
	fmt.Printf("  Size:     %d\n", result.Size)
	fmt.Printf("  Closed:   %v\n", result.Closed)
	return nil
}

func cmdClose(server string, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: client close <id>")
	}

	id := args[0]
	req := common.CloseRequest{ID: id}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(server+"/close", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", string(data))
	}

	var result common.CloseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Error != "" {
		return fmt.Errorf(result.Error)
	}

	fmt.Printf("Closed buffer: %s\n", id)
	return nil
}

func cmdList(server string, args []string) error {
	resp, err := http.Post(server+"/list", "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", string(data))
	}

	var result common.ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Error != "" {
		return fmt.Errorf(result.Error)
	}

	if len(result.Buffers) == 0 {
		fmt.Println("No buffers")
		return nil
	}

	fmt.Println("Buffers:")
	for _, b := range result.Buffers {
		fmt.Printf("  %s: capacity=%d, size=%d\n", b.ID, b.Capacity, b.Size)
	}
	return nil
}
