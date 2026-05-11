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

	"github.com/example/ldap-filter/api"
	"github.com/example/ldap-filter/ldapfilter"
)

type attrMap map[string][]string

func (a *attrMap) String() string {
	return fmt.Sprintf("%v", *a)
}

func (a *attrMap) Set(value string) error {
	if *a == nil {
		*a = make(map[string][]string)
	}
	idx := strings.Index(value, "=")
	if idx < 0 {
		return fmt.Errorf("invalid attribute format: %s, expected key=value", value)
	}
	key := value[:idx]
	val := value[idx+1:]
	(*a)[key] = append((*a)[key], val)
	return nil
}

var serverURL = "http://localhost:8502"

func main() {
	if envURL := os.Getenv("LDAP_FILTER_SERVER"); envURL != "" {
		serverURL = envURL
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  parse <filter>          Parse LDAP filter and print AST tree\n")
		fmt.Fprintf(os.Stderr, "  match <filter>          Test filter against attributes\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		fmt.Fprintf(os.Stderr, "  --attr key=value        Attribute for match (can be used multiple times)\n")
		fmt.Fprintf(os.Stderr, "  --server url            Server URL (default: http://localhost:8502, or LDAP_FILTER_SERVER env)\n")
	}

	args := os.Args[1:]
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	cmd := args[0]
	args = args[1:]

	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	var attrs attrMap
	var serverFlag string
	fs.Var(&attrs, "attr", "attribute key=value")
	fs.StringVar(&serverFlag, "server", "", "server URL")

	switch cmd {
	case "parse":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "Error: parse command requires a filter argument")
			os.Exit(1)
		}
		filter := args[0]
		fs.Parse(args[1:])
		if serverFlag != "" {
			serverURL = serverFlag
		}
		runParse(filter)
	case "match":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "Error: match command requires a filter argument")
			os.Exit(1)
		}
		filter := args[0]
		fs.Parse(args[1:])
		if serverFlag != "" {
			serverURL = serverFlag
		}
		runMatch(filter, attrs)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", cmd)
		flag.Usage()
		os.Exit(1)
	}
}

func runParse(filter string) {
	req := api.ParseRequest{Filter: filter}
	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/ldap/parse", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	var apiResp api.ParseResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if !apiResp.Success {
		fmt.Fprintf(os.Stderr, "Parse error: %s\n", apiResp.Error)
		os.Exit(1)
	}

	node, err := ldapfilter.Parse(filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Local parse error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(ldapfilter.PrintTree(node))
}

func runMatch(filter string, attrs attrMap) {
	req := api.MatchRequest{
		Filter:     filter,
		Attributes: attrs,
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/ldap/match", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	var apiResp api.MatchResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if !apiResp.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", apiResp.Error)
		os.Exit(1)
	}

	fmt.Printf("Filter: %s\n", filter)
	fmt.Printf("Attributes: %v\n", map[string][]string(attrs))
	if apiResp.Matched {
		fmt.Println("Result: MATCH")
	} else {
		fmt.Println("Result: NO MATCH")
	}
}
