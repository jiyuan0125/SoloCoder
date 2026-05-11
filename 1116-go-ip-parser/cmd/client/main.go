package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go-ip-parser/pkg/common"
	"net/http"
	"os"
	"strings"
)

const defaultServer = "http://localhost:8303"

type command func(ip string, serverURL string) error

func main() {
	serverURL := flag.String("server", defaultServer, "IP parser server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := strings.ToLower(args[0])
	ip := args[1]

	var err error
	switch cmd {
	case "parse":
		err = parseCommand(ip, *serverURL)
	case "format":
		err = formatCommand(ip, *serverURL)
	case "classify":
		err = classifyCommand(ip, *serverURL)
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("IP Parser Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client [--server URL] <command> <ip_address>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  parse    - Parse an IP address and show detailed information")
	fmt.Println("  format   - Format an IP address (show standard and canonical forms)")
	fmt.Println("  classify - Classify an IP address by type")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client parse 127.1")
	fmt.Println("  client parse ::ffff:192.168.1.1")
	fmt.Println("  client format 2001:0db8:85a3:0000:0000:8a2e:0370:7334")
	fmt.Println("  client classify 192.168.1.1")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --server URL   Server URL (default: http://localhost:8303)")
}

func parseCommand(ip, serverURL string) error {
	req := common.ParseRequest{IP: ip}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(serverURL+"/parse", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result common.ParseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("parse failed: %s", result.Error)
	}

	fmt.Println("=== IP Address Parse Result ===")
	fmt.Printf("Original:     %s\n", result.Original)
	fmt.Printf("Version:      %s\n", result.Version)
	fmt.Printf("Standard:     %s\n", result.Standard)
	fmt.Printf("Canonical:    %s\n", result.Canonical)
	fmt.Println()
	fmt.Println("Properties:")
	fmt.Printf("  Loopback:    %v\n", result.IsLoopback)
	fmt.Printf("  Private:     %v\n", result.IsPrivate)
	fmt.Printf("  Multicast:   %v\n", result.IsMulticast)
	fmt.Printf("  LinkLocal:   %v\n", result.IsLinkLocal)
	fmt.Printf("  Unspecified: %v\n", result.IsUnspecified)

	if result.IPv4Class != nil {
		fmt.Println()
		fmt.Println("IPv4 Information:")
		fmt.Printf("  Class:       %s\n", *result.IPv4Class)
		fmt.Printf("  Octets:      %d.%d.%d.%d\n",
			result.IPv4Octets[0], result.IPv4Octets[1],
			result.IPv4Octets[2], result.IPv4Octets[3])
	}

	if result.IPv6Type != nil {
		fmt.Println()
		fmt.Println("IPv6 Information:")
		fmt.Printf("  Type:        %s\n", *result.IPv6Type)
		fmt.Printf("  Groups:      %x:%x:%x:%x:%x:%x:%x:%x\n",
			result.IPv6Groups[0], result.IPv6Groups[1],
			result.IPv6Groups[2], result.IPv6Groups[3],
			result.IPv6Groups[4], result.IPv6Groups[5],
			result.IPv6Groups[6], result.IPv6Groups[7])
	}

	return nil
}

func formatCommand(ip, serverURL string) error {
	req := common.FormatRequest{IP: ip}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(serverURL+"/format", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result common.FormatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("format failed: %s", result.Error)
	}

	fmt.Println("=== IP Address Format Result ===")
	fmt.Printf("Original:  %s\n", result.Original)
	fmt.Printf("Standard:  %s\n", result.Formatted)
	fmt.Printf("Canonical: %s\n", result.Compact)

	return nil
}

func classifyCommand(ip, serverURL string) error {
	req := common.ClassifyRequest{IP: ip}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(serverURL+"/classify", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result common.ClassifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("classify failed: %s", result.Error)
	}

	fmt.Println("=== IP Address Classification ===")
	fmt.Printf("IP:          %s\n", result.IP)
	fmt.Printf("Version:     %s\n", result.Version)
	fmt.Printf("Description: %s\n", result.Description)
	fmt.Println()
	fmt.Println("Properties:")
	fmt.Printf("  Loopback:    %v\n", result.IsLoopback)
	fmt.Printf("  Private:     %v\n", result.IsPrivate)
	fmt.Printf("  Multicast:   %v\n", result.IsMulticast)
	fmt.Printf("  LinkLocal:   %v\n", result.IsLinkLocal)
	fmt.Printf("  Unspecified: %v\n", result.IsUnspecified)

	if result.IPv4Class != nil {
		fmt.Println()
		fmt.Printf("IPv4 Class:  %s\n", *result.IPv4Class)
	}

	if result.IPv6Type != nil {
		fmt.Println()
		fmt.Printf("IPv6 Type:   %s\n", *result.IPv6Type)
	}

	return nil
}
