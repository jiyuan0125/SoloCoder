package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"punycode-idn/pkg/api"
)

const defaultServer = "http://localhost:8310"

func getServerURL() string {
	server := os.Getenv("SERVER_URL")
	if server != "" {
		return server
	}
	return defaultServer
}

func doEncode(server, domain string) error {
	reqBody, err := json.Marshal(api.EncodeRequest{Domain: domain})
	if err != nil {
		return err
	}

	resp, err := http.Post(server+"/encode", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result api.EncodeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Error != "" {
		return fmt.Errorf("%s", result.Error)
	}

	fmt.Println(result.Domain)
	return nil
}

func doDecode(server, domain string) error {
	reqBody, err := json.Marshal(api.DecodeRequest{Domain: domain})
	if err != nil {
		return err
	}

	resp, err := http.Post(server+"/decode", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result api.DecodeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Error != "" {
		return fmt.Errorf("%s", result.Error)
	}

	fmt.Println(result.Domain)
	return nil
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage: punycode-client <encode|decode> <domain>")
	fmt.Fprintln(os.Stderr, "  -server string")
	fmt.Fprintln(os.Stderr, "        server URL (default \"http://localhost:8310\")")
	fmt.Fprintln(os.Stderr, "  or use SERVER_URL environment variable")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	if cmd != "encode" && cmd != "decode" {
		printUsage()
		os.Exit(1)
	}

	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	server := fs.String("server", getServerURL(), "server URL")
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) != 1 {
		printUsage()
		os.Exit(1)
	}

	domain := remaining[0]

	var err error
	switch cmd {
	case "encode":
		err = doEncode(*server, domain)
	case "decode":
		err = doDecode(*server, domain)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
