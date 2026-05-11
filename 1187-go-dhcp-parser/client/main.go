package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/dhcp-parser/api"
)

const defaultServerURL = "http://localhost:8205"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	serverURL := defaultServerURL
	if envURL := os.Getenv("DHCP_CLIENT_URL"); envURL != "" {
		serverURL = envURL
	}
	var err error
	switch cmd {
	case "parse":
		err = cmdParse(serverURL, os.Args[2:])
	case "build":
		err = cmdBuild(serverURL, os.Args[2:])
	case "pool":
		err = cmdPool(serverURL, os.Args[2:])
	case "leases":
		err = cmdLeases(serverURL, os.Args[2:])
	case "dora":
		err = cmdDORA(serverURL, os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("DHCP Client CLI")
	fmt.Println()
	fmt.Println("Usage: dhcp-client <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  parse <hex-string>          Parse a DHCP packet from hex string")
	fmt.Println("  build <type> <mac> [opts]   Build a DHCP packet")
	fmt.Println("                              type: discover, offer, request, ack, nak, release")
	fmt.Println("  pool                        Show IP pool status")
	fmt.Println("  pool add <start> <end>      Add IP range to pool")
	fmt.Println("  pool remove <start> <end>   Remove IP range from pool")
	fmt.Println("  leases                      List all active leases")
	fmt.Println("  leases mac <mac>            Find lease by MAC address")
	fmt.Println("  leases ip <ip>              Find lease by IP address")
	fmt.Println("  dora <mac>                  Simulate full DORA process")
	fmt.Println("  help                        Show this help message")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  DHCP_CLIENT_URL             Server URL (default: http://localhost:8205)")
}

func postJSON(url string, body interface{}, result interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if result != nil {
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("failed to parse response: %w, body: %s", err, string(data))
		}
	}
	return nil
}

func getJSON(url string, result interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if result != nil {
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("failed to parse response: %w, body: %s", err, string(data))
		}
	}
	return nil
}

func cmdParse(serverURL string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: dhcp-client parse <hex-string>")
	}
	hexStr := args[0]
	hexStr = strings.ReplaceAll(hexStr, " ", "")
	hexStr = strings.ReplaceAll(hexStr, ":", "")
	hexStr = strings.ReplaceAll(hexStr, "-", "")
	req := api.ParseRequest{Hex: hexStr}
	var resp api.ParseResponse
	if err := postJSON(serverURL+"/parse", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	printPacketInfo(resp.Packet)
	return nil
}

func printPacketInfo(p api.PacketInfo) {
	fmt.Println("=== DHCP Packet ===")
	fmt.Printf("  OP:           %d (%s)\n", p.Op, p.OpName)
	fmt.Printf("  HTYPE:        %d\n", p.Htype)
	fmt.Printf("  HLEN:         %d\n", p.Hlen)
	fmt.Printf("  HOPS:         %d\n", p.Hops)
	fmt.Printf("  XID:          0x%08x (%d)\n", p.Xid, p.Xid)
	fmt.Printf("  SECS:         %d\n", p.Secs)
	fmt.Printf("  FLAGS:        0x%04x", p.Flags)
	if p.Broadcast {
		fmt.Print(" (BROADCAST)")
	}
	fmt.Println()
	fmt.Printf("  CIADDR:       %s\n", p.Ciaddr)
	fmt.Printf("  YIADDR:       %s\n", p.Yiaddr)
	fmt.Printf("  SIADDR:       %s\n", p.Siaddr)
	fmt.Printf("  GIADDR:       %s\n", p.Giaddr)
	fmt.Printf("  CHADDR:       %s\n", p.Chaddr)
	if p.MessageName != "" {
		fmt.Printf("  Message Type: %s (%d)\n", p.MessageName, p.MessageType)
	}
	fmt.Printf("  Options:      %d options\n", len(p.Options))
	for _, opt := range p.Options {
		fmt.Printf("    Option %3d: len=%d, value=0x%s\n", opt.Code, opt.Length, opt.Value)
	}
}

func cmdBuild(serverURL string, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: dhcp-client build <type> <mac> [opts]\n  type: discover, offer, request, ack, nak, release")
	}
	msgType := strings.ToLower(args[0])
	mac := args[1]
	var msgCode byte
	var opCode byte
	switch msgType {
	case "discover":
		msgCode = 1
		opCode = 1
	case "offer":
		msgCode = 2
		opCode = 2
	case "request":
		msgCode = 3
		opCode = 1
	case "ack":
		msgCode = 5
		opCode = 2
	case "nak":
		msgCode = 6
		opCode = 2
	case "release":
		msgCode = 7
		opCode = 1
	default:
		return fmt.Errorf("unknown message type: %s", msgType)
	}
	req := api.BuildRequest{
		Op:          opCode,
		Htype:       1,
		Hlen:        6,
		Chaddr:      mac,
		MessageType: msgCode,
		Broadcast:   true,
	}
	var resp api.BuildResponse
	if err := postJSON(serverURL+"/build", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	fmt.Printf("=== Built DHCP %s Packet ===\n", strings.ToUpper(msgType))
	fmt.Printf("Length: %d bytes\n", len(resp.Bytes))
	fmt.Printf("Hex: %s\n", resp.Hex)
	fmt.Println()
	fmt.Println("Formatted:")
	for i := 0; i < len(resp.Hex); i += 32 {
		end := i + 32
		if end > len(resp.Hex) {
			end = len(resp.Hex)
		}
		line := resp.Hex[i:end]
		withSpaces := ""
		for j := 0; j < len(line); j += 2 {
			if j > 0 {
				withSpaces += " "
			}
			withSpaces += line[j : j+2]
		}
		fmt.Printf("  %s\n", withSpaces)
	}
	return nil
}

func cmdPool(serverURL string, args []string) error {
	if len(args) == 0 {
		var resp api.PoolStatusResponse
		if err := getJSON(serverURL+"/pool", &resp); err != nil {
			return err
		}
		if !resp.Success {
			return fmt.Errorf(resp.Error)
		}
		fmt.Println("=== IP Pool Status ===")
		fmt.Printf("  Total:     %d\n", resp.Total)
		fmt.Printf("  Available: %d\n", resp.Available)
		fmt.Printf("  In Use:    %d\n", resp.InUse)
		return nil
	}
	switch args[0] {
	case "add":
		if len(args) < 3 {
			return fmt.Errorf("usage: dhcp-client pool add <start> <end>")
		}
		req := api.IPRangeRequest{StartIP: args[1], EndIP: args[2]}
		var resp api.IPRangeResponse
		if err := postJSON(serverURL+"/pool/add", req, &resp); err != nil {
			return err
		}
		if !resp.Success {
			return fmt.Errorf(resp.Error)
		}
		fmt.Println("IP range added successfully")
		return nil
	case "remove":
		if len(args) < 3 {
			return fmt.Errorf("usage: dhcp-client pool remove <start> <end>")
		}
		req := api.IPRangeRequest{StartIP: args[1], EndIP: args[2]}
		var resp api.IPRangeResponse
		if err := postJSON(serverURL+"/pool/remove", req, &resp); err != nil {
			return err
		}
		if !resp.Success {
			return fmt.Errorf(resp.Error)
		}
		fmt.Println("IP range removed successfully")
		return nil
	default:
		return fmt.Errorf("unknown pool subcommand: %s", args[0])
	}
}

func cmdLeases(serverURL string, args []string) error {
	if len(args) == 0 {
		var resp api.LeaseListResponse
		if err := getJSON(serverURL+"/leases", &resp); err != nil {
			return err
		}
		if !resp.Success {
			return fmt.Errorf(resp.Error)
		}
		fmt.Printf("=== Active Leases (%d) ===\n", len(resp.Leases))
		for i, l := range resp.Leases {
			fmt.Printf("\n--- Lease %d ---\n", i+1)
			printLease(l)
		}
		if len(resp.Leases) == 0 {
			fmt.Println("  No active leases")
		}
		return nil
	}
	if len(args) < 2 {
		return fmt.Errorf("usage: dhcp-client leases <mac|ip> <value>")
	}
	switch args[0] {
	case "mac":
		var resp api.LeaseResponse
		if err := getJSON(serverURL+"/leases/by-mac?mac="+args[1], &resp); err != nil {
			return err
		}
		if !resp.Success {
			return fmt.Errorf(resp.Error)
		}
		fmt.Println("=== Lease by MAC ===")
		printLease(resp.Lease)
		return nil
	case "ip":
		var resp api.LeaseResponse
		if err := getJSON(serverURL+"/leases/by-ip?ip="+args[1], &resp); err != nil {
			return err
		}
		if !resp.Success {
			return fmt.Errorf(resp.Error)
		}
		fmt.Println("=== Lease by IP ===")
		printLease(resp.Lease)
		return nil
	default:
		return fmt.Errorf("unknown leases subcommand: %s", args[0])
	}
}

func printLease(l api.LeaseInfo) {
	fmt.Printf("  MAC:          %s\n", l.MAC)
	fmt.Printf("  IP:           %s\n", l.IP)
	fmt.Printf("  Server IP:    %s\n", l.ServerIP)
	fmt.Printf("  Start Time:   %s\n", l.StartTime.Local())
	fmt.Printf("  Duration:     %d seconds\n", l.Duration)
	fmt.Printf("  Expire Time:  %s\n", l.ExpireTime.Local())
	fmt.Printf("  T1 (Renew):   %s\n", l.T1Time.Local())
	fmt.Printf("  T2 (Rebind):  %s\n", l.T2Time.Local())
	fmt.Printf("  Status:       ")
	if l.IsExpired {
		fmt.Print("EXPIRED")
	} else if l.IsT2Reached {
		fmt.Print("T2 REACHED")
	} else if l.IsT1Reached {
		fmt.Print("T1 REACHED")
	} else {
		fmt.Print("ACTIVE")
	}
	fmt.Println()
}

func cmdDORA(serverURL string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: dhcp-client dora <mac>")
	}
	req := api.DORARequest{MAC: args[0]}
	var resp api.DORAResponse
	if err := postJSON(serverURL+"/dora", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	fmt.Println("=== DORA Process Completed ===")
	fmt.Printf("  Assigned IP:  %s\n", resp.AssignedIP)
	if resp.Lease != nil {
		fmt.Printf("  MAC Address:  %s\n", resp.Lease.MAC)
		fmt.Printf("  Lease Start:  %s\n", resp.Lease.StartTime.Local())
		fmt.Printf("  Lease End:    %s\n", resp.Lease.ExpireTime.Local())
		fmt.Printf("  Duration:     %d seconds\n", resp.Lease.Duration)
	}
	fmt.Println()
	fmt.Println("The client has been assigned an IP address via DORA:")
	fmt.Println("  1. DISCOVER  -> Client broadcasts for DHCP servers")
	fmt.Println("  2. OFFER     -> Server offers an IP address")
	fmt.Println("  3. REQUEST   -> Client requests the offered IP")
	fmt.Println("  4. ACK       -> Server acknowledges and grants the lease")
	return nil
}

func init() {
	_ = hex.EncodeToString
}
