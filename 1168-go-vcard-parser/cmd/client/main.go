package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"vcard-parser/api"
)

func getServerURL() string {
	url := os.Getenv("VCARD_SERVER_URL")
	if url != "" {
		return strings.TrimSuffix(url, "/")
	}
	flagURL := flag.String("server", "http://localhost:8080", "vCard parser server URL")
	flag.Parse()
	return strings.TrimSuffix(*flagURL, "/")
}

func usage() {
	fmt.Println("Usage: vcard-client [options] <vcf-file>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -server string    vCard parser server URL (default: http://localhost:8080)")
	fmt.Println("                    Can also be set via VCARD_SERVER_URL environment variable")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  vcard-client contacts.vcf")
	fmt.Println("  vcard-client -server http://192.168.1.100:8080 contacts.vcf")
	fmt.Println("  VCARD_SERVER_URL=http://192.168.1.100:8080 vcard-client contacts.vcf")
}

func main() {
	flag.Usage = usage

	serverURL := getServerURL()
	args := flag.Args()

	if len(args) < 1 {
		usage()
		os.Exit(1)
	}

	vcfFile := args[0]

	vcfContent, err := ioutil.ReadFile(vcfFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read vcf file: %v\n", err)
		os.Exit(1)
	}

	reqBody, err := json.Marshal(api.ParseRequest{
		VCardText: string(vcfContent),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/parse", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: server returned status %d\n", resp.StatusCode)
		fmt.Fprintf(os.Stderr, "Response: %s\n", string(respBody))
		os.Exit(1)
	}

	var result api.ParseResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse response: %v\n", err)
		fmt.Fprintf(os.Stderr, "Response: %s\n", string(respBody))
		os.Exit(1)
	}

	if !result.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Error)
		os.Exit(1)
	}

	prettyJSON, err := json.MarshalIndent(result.Contacts, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to format JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(prettyJSON))
}
