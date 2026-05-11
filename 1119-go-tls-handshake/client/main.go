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

	"tls-simulator/common"
)

func main() {
	var (
		serverURL       = flag.String("server", "http://localhost:8700", "TLS handshake simulator server URL")
		clientSuites    = flag.String("client-suites", "", "Comma-separated list of client cipher suite IDs (e.g., 0xC02F,0xC030)")
		serverSuites    = flag.String("server-suites", "", "Comma-separated list of server cipher suite IDs")
		sni             = flag.String("sni", "", "Server Name Indication")
		alpn            = flag.String("alpn", "", "Comma-separated list of ALPN protocols (e.g., h2,http/1.1)")
		resumeSessionID = flag.String("resume", "", "Session ID to resume (hex encoded)")
		listSuites      = flag.Bool("list", false, "List all supported cipher suites")
		verbose         = flag.Bool("verbose", false, "Show detailed message dump")
		jsonOutput      = flag.Bool("json", false, "Output as JSON")
	)
	flag.Parse()
	if *listSuites {
		listCipherSuites(*serverURL, *jsonOutput)
		return
	}
	simulateHandshake(*serverURL, *clientSuites, *serverSuites, *sni, *alpn, *resumeSessionID, *verbose, *jsonOutput)
}

func listCipherSuites(serverURL string, jsonOutput bool) {
	resp, err := http.Get(fmt.Sprintf("%s/cipher-suites", serverURL))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Server returned error: %s\n", string(body))
		os.Exit(1)
	}
	if jsonOutput {
		fmt.Println(string(body))
		return
	}
	var result common.CipherSuitesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Supported Cipher Suites:")
	fmt.Println("========================")
	for _, cs := range result.CipherSuites {
		fmt.Printf("\n  %s (%s)\n", cs.ID, cs.Name)
		fmt.Printf("    Key Exchange: %s\n", cs.KeyExchange)
		fmt.Printf("    Encryption:   %s\n", cs.Encryption)
		fmt.Printf("    MAC:          %s\n", cs.MAC)
	}
}

func simulateHandshake(serverURL, clientSuites, serverSuites, sni, alpn, resumeSessionID string, verbose, jsonOutput bool) {
	req := common.HandshakeRequest{
		Verbose: verbose,
	}
	if clientSuites != "" {
		req.ClientCipherSuites = parseCommaList(clientSuites)
	}
	if serverSuites != "" {
		req.ServerCipherSuites = parseCommaList(serverSuites)
	}
	if sni != "" {
		req.SNI = sni
	}
	if alpn != "" {
		req.ALPNProtocols = parseCommaList(alpn)
	}
	if resumeSessionID != "" {
		req.ResumeSessionID = resumeSessionID
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling request: %v\n", err)
		os.Exit(1)
	}
	resp, err := http.Post(
		fmt.Sprintf("%s/handshake", serverURL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Server returned error: %s\n", string(body))
		os.Exit(1)
	}
	if jsonOutput {
		fmt.Println(string(body))
		return
	}
	var result common.HandshakeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}
	printResult(&result, verbose)
}

func parseCommaList(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func printResult(result *common.HandshakeResponse, verbose bool) {
	fmt.Println("TLS 1.2 Handshake Simulation")
	fmt.Println("=============================")
	if result.Success {
		fmt.Println("\n✓ Handshake completed successfully!")
	} else {
		fmt.Printf("\n✗ Handshake failed: %s\n", result.Error)
	}
	if result.IsSessionResume {
		fmt.Println("\n  Note: Using session resumption (abbreviated handshake)")
	}
	if result.NegotiatedSuite != nil {
		fmt.Println("\nNegotiated Cipher Suite:")
		fmt.Printf("  ID:          %s\n", result.NegotiatedSuite.ID)
		fmt.Printf("  Name:        %s\n", result.NegotiatedSuite.Name)
		fmt.Printf("  Key Exchange: %s\n", result.NegotiatedSuite.KeyExchange)
		fmt.Printf("  Encryption:   %s\n", result.NegotiatedSuite.Encryption)
		fmt.Printf("  MAC:          %s\n", result.NegotiatedSuite.MAC)
	}
	if result.SessionID != "" {
		fmt.Printf("\nSession ID: %s\n", result.SessionID)
	}
	fmt.Println("\nHandshake Steps:")
	fmt.Println("----------------")
	for i, step := range result.Steps {
		fmt.Printf("  %s\n", step)
		_ = i
	}
	if len(result.ClientExtensions) > 0 {
		fmt.Println("\nClient Extensions:")
		for _, ext := range result.ClientExtensions {
			fmt.Printf("  - %s\n", ext)
		}
	}
	if verbose && result.FullMessageDump != "" {
		fmt.Println("\n\nDetailed Message Dump:")
		fmt.Println("=====================")
		fmt.Println(result.FullMessageDump)
	}
}
