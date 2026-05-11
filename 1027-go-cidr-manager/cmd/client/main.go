package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/ip-manager/pkg/common"
)

const defaultServerURL = "http://localhost:8104"

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	return &Client{serverURL: serverURL}
}

func (c *Client) post(endpoint string, body interface{}) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	url := fmt.Sprintf("%s%s", c.serverURL, endpoint)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	return data, nil
}

func printHelp() {
	fmt.Println(`IP Manager Client

Usage:
  ipm <command> [options]

Commands:
  parse <cidr>                    Parse a CIDR and show details
  contains <cidr> <ip>            Check if an IP is in a CIDR
  merge <cidr1> <cidr2>...        Merge adjacent CIDR blocks
  pool create <pool_id> <cidr>    Create a new IP pool
  pool status <pool_id>           Show pool status
  pool allocate <pool_id> <prefix> Allocate a subnet from pool
  pool release <pool_id> <cidr>   Release a subnet back to pool
  pool exclude <pool_id> <cidr>   Exclude a CIDR from available pool

Options:
  --server <url>                  Server URL (default: http://localhost:8104)
  --help, -h                      Show this help message

Examples:
  ipm parse 192.168.1.0/24
  ipm contains 192.168.1.0/24 192.168.1.100
  ipm pool create mypool 10.0.0.0/16
  ipm pool allocate mypool 24
  ipm pool release mypool 10.0.0.0/24`)
}

func main() {
	helpFlag := flag.Bool("help", false, "Show help message")
	helpShortFlag := flag.Bool("h", false, "Show help message")
	serverFlag := flag.String("server", defaultServerURL, "Server URL")

	flag.Usage = func() {
		printHelp()
	}

	flag.Parse()

	if *helpFlag || *helpShortFlag {
		printHelp()
		return
	}

	args := flag.Args()
	if len(args) < 1 {
		printHelp()
		os.Exit(1)
	}

	client := NewClient(*serverFlag)

	command := args[0]
	switch command {
	case "parse":
		if len(args) < 2 {
			fmt.Println("Error: parse requires a CIDR argument")
			printHelp()
			os.Exit(1)
		}
		handleParse(client, args[1])

	case "contains":
		if len(args) < 3 {
			fmt.Println("Error: contains requires CIDR and IP arguments")
			printHelp()
			os.Exit(1)
		}
		handleContains(client, args[1], args[2])

	case "merge":
		if len(args) < 3 {
			fmt.Println("Error: merge requires at least two CIDR arguments")
			printHelp()
			os.Exit(1)
		}
		handleMerge(client, args[1:])

	case "pool":
		if len(args) < 2 {
			fmt.Println("Error: pool requires a subcommand")
			printHelp()
			os.Exit(1)
		}
		handlePoolCommand(client, args[1:])

	default:
		fmt.Printf("Error: unknown command '%s'\n", command)
		printHelp()
		os.Exit(1)
	}
}

func handleParse(client *Client, cidr string) {
	req := common.ParseRequest{CIDR: cidr}
	data, err := client.post("/parse", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var resp common.ParseResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		return
	}

	fmt.Printf("CIDR:        %s\n", resp.CIDR)
	fmt.Printf("Network:     %s\n", resp.Network)
	fmt.Printf("Broadcast:   %s\n", resp.Broadcast)
	fmt.Printf("First Usable: %s\n", resp.FirstUsable)
	fmt.Printf("Last Usable:  %s\n", resp.LastUsable)
	fmt.Printf("Prefix:      /%d\n", resp.Prefix)
	fmt.Printf("Size:        %d addresses\n", resp.Size)
	fmt.Printf("Version:     %s\n", resp.Version)
}

func handleContains(client *Client, cidr, ip string) {
	req := common.ContainsRequest{CIDR: cidr, IP: ip}
	data, err := client.post("/contains", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var resp common.ContainsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		return
	}

	if resp.Contained {
		fmt.Printf("IP %s is contained in CIDR %s\n", ip, cidr)
	} else {
		fmt.Printf("IP %s is NOT contained in CIDR %s\n", ip, cidr)
	}
}

func handleMerge(client *Client, cidrs []string) {
	req := common.MergeRequest{CIDRs: cidrs}
	data, err := client.post("/merge", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var resp common.MergeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		return
	}

	fmt.Println("Merged CIDR blocks:")
	for _, c := range resp.Merged {
		fmt.Printf("  %s\n", c)
	}
}

func handlePoolCommand(client *Client, args []string) {
	subcmd := args[0]
	switch subcmd {
	case "create":
		if len(args) < 3 {
			fmt.Println("Error: pool create requires pool_id and CIDR")
			os.Exit(1)
		}
		handlePoolCreate(client, args[1], args[2])

	case "status":
		if len(args) < 2 {
			fmt.Println("Error: pool status requires pool_id")
			os.Exit(1)
		}
		handlePoolStatus(client, args[1])

	case "allocate":
		if len(args) < 3 {
			fmt.Println("Error: pool allocate requires pool_id and prefix")
			os.Exit(1)
		}
		prefix, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Printf("Error: invalid prefix '%s'\n", args[2])
			os.Exit(1)
		}
		handlePoolAllocate(client, args[1], prefix)

	case "release":
		if len(args) < 3 {
			fmt.Println("Error: pool release requires pool_id and CIDR")
			os.Exit(1)
		}
		handlePoolRelease(client, args[1], args[2])

	case "exclude":
		if len(args) < 3 {
			fmt.Println("Error: pool exclude requires pool_id and CIDR")
			os.Exit(1)
		}
		handlePoolExclude(client, args[1], args[2])

	default:
		fmt.Printf("Error: unknown pool subcommand '%s'\n", subcmd)
		os.Exit(1)
	}
}

func handlePoolCreate(client *Client, poolID, cidr string) {
	req := common.CreatePoolRequest{PoolID: poolID, CIDR: cidr}
	data, err := client.post("/pool/create", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var resp common.CreatePoolResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		return
	}

	fmt.Printf("Pool '%s' created successfully\n", resp.PoolID)
}

func handlePoolStatus(client *Client, poolID string) {
	req := common.PoolStatusRequest{PoolID: poolID}
	data, err := client.post("/pool/status", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var resp common.PoolStatusResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		return
	}

	fmt.Printf("Pool: %s\n\n", resp.PoolID)

	fmt.Println("Available:")
	if len(resp.Available) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, c := range resp.Available {
			fmt.Printf("  %s\n", c)
		}
	}

	fmt.Println("\nUsed:")
	if len(resp.Used) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, c := range resp.Used {
			fmt.Printf("  %s\n", c)
		}
	}
}

func handlePoolAllocate(client *Client, poolID string, prefix int) {
	req := common.AllocateRequest{PoolID: poolID, Prefix: prefix}
	data, err := client.post("/pool/allocate", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var resp common.AllocateResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		return
	}

	fmt.Printf("Allocated: %s\n", resp.Allocated)
}

func handlePoolRelease(client *Client, poolID, cidr string) {
	req := common.ReleaseRequest{PoolID: poolID, CIDR: cidr}
	data, err := client.post("/pool/release", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var resp common.ReleaseResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		return
	}

	fmt.Printf("Released %s from pool '%s'\n", cidr, poolID)
}

func handlePoolExclude(client *Client, poolID, cidr string) {
	req := common.ExcludeRequest{PoolID: poolID, CIDR: cidr}
	data, err := client.post("/pool/exclude", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var resp common.ExcludeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		return
	}

	fmt.Printf("Excluded %s from pool '%s'\n", cidr, poolID)
}
