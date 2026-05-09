package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/solocoder/openpgp-parser/api"
)

const defaultServer = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	server := os.Getenv("PGP_PARSER_SERVER")
	if server == "" {
		server = defaultServer
	}

	switch os.Args[1] {
	case "parse":
		handleParse(server, os.Args[2:])
	case "keyinfo":
		handleKeyInfo(server, os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("OpenPGP Parser Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client parse [flags]      Parse PGP ASCII Armor from stdin")
	fmt.Println("  client keyinfo [flags]    Extract key info from public key in stdin")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -server URL       Server URL (default: http://localhost:8080)")
	fmt.Println("                    Can also be set via PGP_PARSER_SERVER env var")
	fmt.Println("  -raw              Output raw JSON response")
}

func handleParse(server string, args []string) {
	flagSet := flag.NewFlagSet("parse", flag.ExitOnError)
	rawOutput := flagSet.Bool("raw", false, "output raw JSON response")
	flagSet.StringVar(&server, "server", server, "server URL")
	flagSet.Parse(args)

	armor, err := readStdin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	resp, err := callParseAPI(server, armor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *rawOutput {
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(data))
		return
	}

	printParseResult(resp)
}

func handleKeyInfo(server string, args []string) {
	flagSet := flag.NewFlagSet("keyinfo", flag.ExitOnError)
	rawOutput := flagSet.Bool("raw", false, "output raw JSON response")
	flagSet.StringVar(&server, "server", server, "server URL")
	flagSet.Parse(args)

	armor, err := readStdin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	resp, err := callKeyInfoAPI(server, armor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *rawOutput {
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(data))
		return
	}

	printKeyInfoResult(resp)
}

func readStdin() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func callParseAPI(server, armor string) (*api.ParseResponse, error) {
	req := api.ParseRequest{Armor: armor}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	url := strings.TrimRight(server, "/") + "/parse"
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to call server: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var result api.ParseResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &result, nil
}

func callKeyInfoAPI(server, armor string) (*api.KeyInfoResponse, error) {
	req := api.KeyInfoRequest{Armor: armor}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	url := strings.TrimRight(server, "/") + "/keyinfo"
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to call server: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var result api.KeyInfoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &result, nil
}

func printParseResult(resp *api.ParseResponse) {
	if !resp.Success {
		fmt.Println("Error:", resp.Error)
		return
	}

	if resp.Armor != nil {
		fmt.Println("=== Armor Information ===")
		fmt.Println("  Type:", resp.Armor.Type)
		fmt.Printf("  Payload Size: %d bytes\n", resp.Armor.PayloadSize)
		if len(resp.Armor.Headers) > 0 {
			fmt.Println("  Headers:")
			for k, v := range resp.Armor.Headers {
				fmt.Printf("    %s: %s\n", k, v)
			}
		}
		fmt.Printf("  Checksum: exists=%v, valid=%v\n", resp.Armor.ChecksumExists, resp.Armor.ChecksumValid)
		fmt.Println()
	}

	fmt.Println("=== Packet Structure ===")
	for i, packet := range resp.Packets {
		fmt.Printf("Packet %d:\n", i+1)
		fmt.Printf("  Tag: %s (%d)\n", packet.Tag, packet.TagValue)
		fmt.Printf("  Format: %s\n", formatBool(packet.IsNewFormat, "New", "Old"))
		fmt.Printf("  Length: %d bytes\n", packet.Length)
		fmt.Printf("  Indefinite: %v\n", packet.IsIndefinite)

		if len(packet.Subpackets) > 0 {
			fmt.Printf("  Subpackets (%d):\n", len(packet.Subpackets))
			for j, sp := range packet.Subpackets {
				fmt.Printf("    [%d] Type=%d, Length=%d, Data=%s\n",
					j, sp.Type, sp.Length, hex.EncodeToString(sp.Data))
			}
		}

		if packet.BodyPreview != "" {
			fmt.Printf("  Body Preview: %s\n", packet.BodyPreview)
		}
		fmt.Println()
	}
}

func printKeyInfoResult(resp *api.KeyInfoResponse) {
	if !resp.Success {
		fmt.Println("Error:", resp.Error)
		return
	}

	if resp.PrimaryKey != nil {
		fmt.Println("=== Primary Key ===")
		fmt.Println("  Version:", resp.PrimaryKey.Version)
		fmt.Println("  Algorithm:", resp.PrimaryKey.Algorithm)
		fmt.Println("  Creation Time:", resp.PrimaryKey.CreationTime)
		fmt.Println("  Key ID:", resp.PrimaryKey.KeyID)
		fmt.Println("  Fingerprint:", resp.PrimaryKey.Fingerprint)
		fmt.Println()
	}

	if len(resp.UserIDs) > 0 {
		fmt.Println("=== User IDs ===")
		for _, uid := range resp.UserIDs {
			fmt.Printf("  %s\n", uid)
		}
		fmt.Println()
	}

	if len(resp.Subkeys) > 0 {
		fmt.Printf("=== Subkeys (%d) ===\n", len(resp.Subkeys))
		for i, sk := range resp.Subkeys {
			fmt.Printf("Subkey %d:\n", i+1)
			fmt.Println("  Version:", sk.Version)
			fmt.Println("  Algorithm:", sk.Algorithm)
			fmt.Println("  Creation Time:", sk.CreationTime)
			fmt.Println("  Key ID:", sk.KeyID)
			fmt.Println("  Fingerprint:", sk.Fingerprint)
			fmt.Println()
		}
	}
}

func formatBool(value bool, trueStr, falseStr string) string {
	if value {
		return trueStr
	}
	return falseStr
}
