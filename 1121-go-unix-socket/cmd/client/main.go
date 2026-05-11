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

	"unixsocket-fdpass/api"
)

var serverAddr = flag.String("server", "http://127.0.0.1:8402", "server address")

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	switch cmd {
	case "send":
		doSend()
	case "health":
		doHealth()
	case "status":
		doStatus()
	case "restart":
		doRestart()
	case "history":
		doHistory()
	default:
		fmt.Printf("unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: client [flags] <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  send [msg]     Send a request to server")
	fmt.Println("  health         Check server health")
	fmt.Println("  status         Show server status")
	fmt.Println("  restart        Trigger hot restart")
	fmt.Println("  history        Show restart history")
	fmt.Println("")
	fmt.Println("Flags:")
	flag.PrintDefaults()
}

func doSend() {
	msg := "hello"
	if len(flag.Args()) > 1 {
		msg = flag.Arg(1)
	}

	url := fmt.Sprintf("%s/echo?msg=%s", *serverAddr, url.QueryEscape(msg))
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)

	var result api.EchoResponse
	if err := json.Unmarshal(body, &result); err == nil {
		fmt.Printf("PID:     %d\n", result.PID)
		fmt.Printf("Message: %s\n", result.Message)
	} else {
		fmt.Printf("Body:    %s\n", strings.TrimSpace(string(body)))
	}
}

func doHealth() {
	url := fmt.Sprintf("%s/health", *serverAddr)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("health check failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)

	var result api.HealthResponse
	if err := json.Unmarshal(body, &result); err == nil {
		fmt.Printf("Health:  %s\n", result.Status)
		fmt.Printf("PID:     %d\n", result.PID)
	} else {
		fmt.Printf("Body:    %s\n", strings.TrimSpace(string(body)))
	}
}

func doStatus() {
	url := fmt.Sprintf("%s/status", *serverAddr)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("status check failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)

	var result api.StatusResponse
	if err := json.Unmarshal(body, &result); err == nil {
		fmt.Printf("PID:          %d\n", result.PID)
		fmt.Printf("ListenAddr:   %s\n", result.ListenAddr)
		fmt.Printf("IsPrimary:    %v\n", result.IsPrimary)
		fmt.Printf("RestartCount: %d\n", result.RestartCount)
	} else {
		fmt.Printf("Body:    %s\n", strings.TrimSpace(string(body)))
	}
}

func doRestart() {
	url := fmt.Sprintf("%s/restart", *serverAddr)
	resp, err := http.Post(url, "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		fmt.Printf("restart request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Body:   %s\n", strings.TrimSpace(string(body)))
}

func doHistory() {
	url := fmt.Sprintf("%s/history", *serverAddr)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("history request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)

	var result api.RestartHistoryResponse
	if err := json.Unmarshal(body, &result); err == nil {
		if len(result.History) == 0 {
			fmt.Println("No restart history")
		} else {
			for i, h := range result.History {
				fmt.Printf("Restart #%d:\n", i+1)
				fmt.Printf("  OldPID:     %d\n", h.OldPID)
				fmt.Printf("  NewPID:     %d\n", h.NewPID)
				fmt.Printf("  Successful: %v\n", h.Successful)
			}
		}
	} else {
		fmt.Printf("Body:    %s\n", strings.TrimSpace(string(body)))
	}
}
