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

	"tcp-framing/protocol"
)

const serverURL = "http://localhost:8400"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	switch cmd {
	case "config":
		handleConfig()
	case "echo":
		handleEcho()
	default:
		fmt.Printf("unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: client <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  config get                   Get current server configuration")
	fmt.Println("  config set [options]         Update server configuration")
	fmt.Println("  echo [options] <messages>    Send messages to echo endpoint")
	fmt.Println()
	fmt.Println("Config set options:")
	fmt.Println("  -mode string        Framer mode: length_prefix or delimiter")
	fmt.Println("  -header-size int    Header size in bytes: 2 or 4 (for length_prefix mode)")
	fmt.Println("  -byte-order string  Byte order: big or little (for length_prefix mode)")
	fmt.Println("  -max-frame int      Maximum frame size in bytes")
	fmt.Println("  -delimiter string   Delimiter string (for delimiter mode)")
	fmt.Println()
	fmt.Println("Echo options:")
	fmt.Println("  -mode string        Override framer mode for this request")
	fmt.Println("  -messages strings   Messages to send (comma-separated, or pass as positional args)")
}

func handleConfig() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: client config get|set")
		os.Exit(1)
	}
	subCmd := os.Args[2]
	switch subCmd {
	case "get":
		getConfig()
	case "set":
		setConfig()
	default:
		fmt.Printf("unknown config subcommand: %s\n", subCmd)
		os.Exit(1)
	}
}

func getConfig() {
	resp, err := http.Get(serverURL + "/config")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error reading response: %v\n", err)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("server error: %s\n", string(body))
		os.Exit(1)
	}
	var cfg protocol.GetConfigResponse
	if err := json.Unmarshal(body, &cfg); err != nil {
		fmt.Printf("error parsing response: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Current configuration:")
	fmt.Printf("  Mode:         %s\n", cfg.Mode)
	fmt.Printf("  Header Size:  %d bytes\n", cfg.HeaderSize)
	fmt.Printf("  Byte Order:   %s\n", cfg.ByteOrder)
	fmt.Printf("  Max Frame:    %d bytes\n", cfg.MaxFrameSize)
	fmt.Printf("  Delimiter:    %q\n", cfg.Delimiter)
}

func setConfig() {
	fs := flag.NewFlagSet("config set", flag.ExitOnError)
	mode := fs.String("mode", "", "Framer mode: length_prefix or delimiter")
	headerSize := fs.Int("header-size", 0, "Header size in bytes: 2 or 4")
	byteOrder := fs.String("byte-order", "", "Byte order: big or little")
	maxFrame := fs.Int("max-frame", 0, "Maximum frame size in bytes")
	delimiter := fs.String("delimiter", "", "Delimiter string")
	fs.Parse(os.Args[3:])
	req := protocol.ConfigRequest{}
	if *mode != "" {
		req.Mode = protocol.FramerMode(*mode)
	}
	if *headerSize != 0 {
		req.HeaderSize = *headerSize
	}
	if *byteOrder != "" {
		req.ByteOrder = protocol.ByteOrder(*byteOrder)
	}
	if *maxFrame != 0 {
		req.MaxFrameSize = *maxFrame
	}
	if *delimiter != "" {
		req.Delimiter = *delimiter
	}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("error encoding request: %v\n", err)
		os.Exit(1)
	}
	resp, err := http.Post(serverURL+"/config", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error reading response: %v\n", err)
		os.Exit(1)
	}
	var respData protocol.ConfigResponse
	if err := json.Unmarshal(respBody, &respData); err == nil {
		if respData.Success {
			fmt.Printf("Success: %s\n", respData.Message)
		} else {
			fmt.Printf("Error: %s\n", respData.Message)
			os.Exit(1)
		}
	} else {
		if resp.StatusCode == http.StatusOK {
			fmt.Println("Success: config updated")
		} else {
			fmt.Printf("Error: %s\n", string(respBody))
			os.Exit(1)
		}
	}
}

func handleEcho() {
	fs := flag.NewFlagSet("echo", flag.ExitOnError)
	mode := fs.String("mode", "", "Override framer mode for this request")
	messagesFlag := fs.String("messages", "", "Messages to send (comma-separated)")
	fs.Parse(os.Args[2:])
	var messages []string
	if *messagesFlag != "" {
		messages = strings.Split(*messagesFlag, ",")
	}
	positional := fs.Args()
	if len(positional) > 0 {
		messages = append(messages, positional...)
	}
	if len(messages) == 0 {
		fmt.Println("error: no messages provided")
		fmt.Println("usage: client echo [options] <message1> <message2> ...")
		os.Exit(1)
	}
	msgBytes := make([][]byte, len(messages))
	for i, m := range messages {
		msgBytes[i] = []byte(m)
	}
	req := protocol.EchoRequest{
		Messages: msgBytes,
	}
	if *mode != "" {
		req.Mode = protocol.FramerMode(*mode)
	}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("error encoding request: %v\n", err)
		os.Exit(1)
	}
	resp, err := http.Post(serverURL+"/echo", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error reading response: %v\n", err)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("server error: %s\n", string(respBody))
		os.Exit(1)
	}
	var respData protocol.EchoResponse
	if err := json.Unmarshal(respBody, &respData); err != nil {
		fmt.Printf("error parsing response: %v\n", err)
		fmt.Printf("raw response: %s\n", string(respBody))
		os.Exit(1)
	}
	if !respData.Success {
		fmt.Printf("Error: %s\n", respData.Message)
		os.Exit(1)
	}
	fmt.Printf("Success! Received %d messages:\n", len(respData.Messages))
	for i, msg := range respData.Messages {
		fmt.Printf("  [%d] %q\n", i, string(msg))
	}
}
