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
	"text/tabwriter"

	"go-reflect-tags/pkg/common"
)

const (
	defaultServerURL = "http://localhost:8080"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "register":
		handleRegister()
	case "query":
		handleQuery()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client register [--server URL] < struct.json")
	fmt.Println("  client query [--server URL] <structName> [fieldPath]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  register   Register a struct definition from stdin")
	fmt.Println("  query      Query tag info for a struct field")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --server   Server URL (default: http://localhost:8080)")
}

func handleRegister() {
	flags := flag.NewFlagSet("register", flag.ExitOnError)
	serverURL := flags.String("server", defaultServerURL, "server URL")
	flags.Parse(os.Args[2:])

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read from stdin: %v\n", err)
		os.Exit(1)
	}

	var req common.RegisterRequest
	if err := json.Unmarshal(data, &req); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse JSON: %v\n", err)
		os.Exit(1)
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(*serverURL+"/register", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "server returned error (status %d): %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var result common.RegisterResponse
	json.Unmarshal(respBody, &result)

	if result.Success {
		fmt.Println(result.Message)
	} else {
		fmt.Fprintf(os.Stderr, "error: %s\n", result.Message)
		os.Exit(1)
	}
}

func handleQuery() {
	flags := flag.NewFlagSet("query", flag.ExitOnError)
	serverURL := flags.String("server", defaultServerURL, "server URL")
	flags.Parse(os.Args[2:])

	args := flags.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "error: struct name is required")
		os.Exit(1)
	}

	structName := args[0]
	fieldPath := ""
	if len(args) > 1 {
		fieldPath = args[1]
	}

	req := common.QueryRequest{
		StructName: structName,
		FieldPath:  fieldPath,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(*serverURL+"/query", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "server returned error (status %d): %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var result common.QueryResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse response: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Fprintf(os.Stderr, "error: %s\n", result.Message)
		os.Exit(1)
	}

	printTable(result.Data)
}

func printTable(info *common.FieldTagInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "Field Path:\t", info.FieldPath)
	fmt.Fprintln(w, "Field Type:\t", info.FieldType)
	fmt.Fprintln(w, "")

	fmt.Fprintln(w, "Tag\tValue\t")
	fmt.Fprintln(w, "---\t---\t")

	if info.Json != nil {
		jsonInfo := info.Json
		if jsonInfo.Ignored {
			fmt.Fprintln(w, "json\t(ignored)\t")
		} else {
			parts := []string{jsonInfo.Name}
			if jsonInfo.OmitEmpty {
				parts = append(parts, "omitempty")
			}
			if jsonInfo.String {
				parts = append(parts, "string")
			}
			fmt.Fprintf(w, "json\t%s\t\n", strings.Join(parts, ", "))
		}
	}

	if info.Db != nil {
		dbInfo := info.Db
		if dbInfo.Ignored {
			fmt.Fprintln(w, "db\t(ignored)\t")
		} else {
			parts := []string{dbInfo.Name}
			if dbInfo.IndexType != "" {
				parts = append(parts, string(dbInfo.IndexType))
			}
			fmt.Fprintf(w, "db\t%s\t\n", strings.Join(parts, ", "))
		}
	}

	if info.Validate != nil {
		vInfo := info.Validate
		var rules []string
		for _, rule := range vInfo.Rules {
			if rule.Value != "" {
				rules = append(rules, fmt.Sprintf("%s=%s", rule.Name, rule.Value))
			} else {
				rules = append(rules, rule.Name)
			}
		}
		fmt.Fprintf(w, "validate\t%s\t\n", strings.Join(rules, ", "))
	}

	w.Flush()
}
