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

	"graphql-parser/internal/common"
	"graphql-parser/internal/graphql/executor"
)

const defaultServerURL = "http://localhost:8080"

type client struct {
	serverURL string
	httpClient *http.Client
}

func newClient() *client {
	url := os.Getenv("GRAPHQL_SERVER_URL")
	if url == "" {
		url = defaultServerURL
	}
	return &client{
		serverURL:  url,
		httpClient: &http.Client{},
	}
}

func (c *client) postJSON(endpoint string, reqBody interface{}, respBody interface{}) error {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	url := strings.TrimSuffix(c.serverURL, "/") + endpoint
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server error: %s - %s", resp.Status, string(body))
	}

	if respBody != nil {
		if err := json.Unmarshal(body, respBody); err != nil {
			return fmt.Errorf("failed to unmarshal response: %v\nResponse: %s", err, string(body))
		}
	}

	return nil
}

func (c *client) getJSON(endpoint string, respBody interface{}) error {
	url := strings.TrimSuffix(c.serverURL, "/") + endpoint
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server error: %s - %s", resp.Status, string(body))
	}

	if respBody != nil {
		if err := json.Unmarshal(body, respBody); err != nil {
			return fmt.Errorf("failed to unmarshal response: %v\nResponse: %s", err, string(body))
		}
	}

	return nil
}

func cmdQuery(args []string) {
	fs := flag.NewFlagSet("query", flag.ExitOnError)
	variables := fs.String("variables", "{}", "Variables as JSON string")
	serverFlag := fs.String("server", "", "GraphQL server URL (overrides GRAPHQL_SERVER_URL)")
	fs.Parse(args)

	queryArgs := fs.Args()
	if len(queryArgs) == 0 {
		fmt.Fprintln(os.Stderr, "Error: query required")
		fmt.Fprintln(os.Stderr, "Usage: gqlc query <graphql-query>")
		os.Exit(1)
	}
	query := strings.Join(queryArgs, " ")

	var vars map[string]interface{}
	if *variables != "" && *variables != "{}" {
		if err := json.Unmarshal([]byte(*variables), &vars); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid variables JSON: %v\n", err)
			os.Exit(1)
		}
	}

	client := newClient()
	if *serverFlag != "" {
		client.serverURL = *serverFlag
	}

	req := common.GraphQLRequest{
		Query:     query,
		Variables: vars,
	}

	var resp common.GraphQLResponse
	if err := client.postJSON("/api/graphql", &req, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printResponse(&resp)
}

func cmdFile(args []string) {
	fs := flag.NewFlagSet("file", flag.ExitOnError)
	variables := fs.String("variables", "{}", "Variables as JSON string")
	serverFlag := fs.String("server", "", "GraphQL server URL")
	fs.Parse(args)

	fileArgs := fs.Args()
	if len(fileArgs) == 0 {
		fmt.Fprintln(os.Stderr, "Error: file required")
		fmt.Fprintln(os.Stderr, "Usage: gqlc file <query-file.gql>")
		os.Exit(1)
	}

	queryBytes, err := ioutil.ReadFile(fileArgs[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read file: %v\n", err)
		os.Exit(1)
	}
	query := string(queryBytes)

	var vars map[string]interface{}
	if *variables != "" && *variables != "{}" {
		if err := json.Unmarshal([]byte(*variables), &vars); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid variables JSON: %v\n", err)
			os.Exit(1)
		}
	}

	client := newClient()
	if *serverFlag != "" {
		client.serverURL = *serverFlag
	}

	req := common.GraphQLRequest{
		Query:     query,
		Variables: vars,
	}

	var resp common.GraphQLResponse
	if err := client.postJSON("/api/graphql", &req, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printResponse(&resp)
}

func cmdLint(args []string) {
	fs := flag.NewFlagSet("lint", flag.ExitOnError)
	fs.Parse(args)

	lintArgs := fs.Args()
	if len(lintArgs) == 0 {
		fmt.Fprintln(os.Stderr, "Error: query or file required")
		fmt.Fprintln(os.Stderr, "Usage: gqlc lint <query-or-file>")
		os.Exit(1)
	}

	input := strings.Join(lintArgs, " ")
	var query string

	if strings.HasSuffix(strings.ToLower(lintArgs[0]), ".gql") || strings.HasSuffix(strings.ToLower(lintArgs[0]), ".graphql") {
		queryBytes, err := ioutil.ReadFile(lintArgs[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to read file: %v\n", err)
			os.Exit(1)
		}
		query = string(queryBytes)
	} else {
		query = input
	}

	_, parseErr := executor.Validate(query)
	formatted, formatErr := executor.Format(query)
	if parseErr != nil {
		fmt.Println("✗ Query is invalid")
		fmt.Printf("  Error: %v\n", parseErr)
		os.Exit(1)
	}

	fmt.Println("✓ Query is valid")
	if formatErr == nil {
		fmt.Println("\nFormatted query:")
		fmt.Println(formatted)
	}
}

func cmdFormat(args []string) {
	fs := flag.NewFlagSet("format", flag.ExitOnError)
	output := fs.String("output", "", "Output file (defaults to stdout)")
	fs.Parse(args)

	formatArgs := fs.Args()
	if len(formatArgs) == 0 {
		fmt.Fprintln(os.Stderr, "Error: query or file required")
		fmt.Fprintln(os.Stderr, "Usage: gqlc format <query-or-file>")
		os.Exit(1)
	}

	input := strings.Join(formatArgs, " ")
	var query string
	var sourceFile string

	if strings.HasSuffix(strings.ToLower(formatArgs[0]), ".gql") || strings.HasSuffix(strings.ToLower(formatArgs[0]), ".graphql") {
		sourceFile = formatArgs[0]
		queryBytes, err := ioutil.ReadFile(sourceFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to read file: %v\n", err)
			os.Exit(1)
		}
		query = string(queryBytes)
	} else {
		query = input
	}

	formatted, err := executor.Format(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *output != "" {
		if err := ioutil.WriteFile(*output, []byte(formatted), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to write output: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Formatted query written to %s\n", *output)
	} else if sourceFile != "" {
		if err := ioutil.WriteFile(sourceFile, []byte(formatted), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to write file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Formatted %s\n", sourceFile)
	} else {
		fmt.Println(formatted)
	}
}

func printResponse(resp *common.GraphQLResponse) {
	if len(resp.Errors) > 0 {
		fmt.Println("Errors:")
		for _, err := range resp.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	if len(resp.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, w := range resp.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	if resp.Data != nil {
		dataJSON, err := json.MarshalIndent(resp.Data, "", "  ")
		if err != nil {
			fmt.Printf("Data: %v\n", resp.Data)
		} else {
			fmt.Println("\nData:")
			fmt.Println(string(dataJSON))
		}
	}
}

func cmdSchema(args []string) {
	fs := flag.NewFlagSet("schema", flag.ExitOnError)
	serverFlag := fs.String("server", "", "GraphQL server URL")
	fs.Parse(args)

	client := newClient()
	if *serverFlag != "" {
		client.serverURL = *serverFlag
	}

	var resp common.SchemaResponse
	if err := client.getJSON("/api/schema", &resp); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for _, t := range resp.Types {
		fmt.Printf("type %s {\n", t.Name)
		for _, f := range t.Fields {
			args := ""
			if len(f.Args) > 0 {
				argParts := []string{}
				for k, v := range f.Args {
					argParts = append(argParts, fmt.Sprintf("%s: %s", k, v))
				}
				args = fmt.Sprintf("(%s)", strings.Join(argParts, ", "))
			}
			fmt.Printf("  %s%s: %s\n", f.Name, args, f.Type)
		}
		fmt.Println("}")
	}
}

func cmdVariables(args []string) {
	fs := flag.NewFlagSet("variables", flag.ExitOnError)
	serverFlag := fs.String("server", "", "GraphQL server URL")
	setFlag := fs.String("set", "", "Set variables as JSON")
	fs.Parse(args)

	client := newClient()
	if *serverFlag != "" {
		client.serverURL = *serverFlag
	}

	if *setFlag != "" {
		var vars map[string]interface{}
		if err := json.Unmarshal([]byte(*setFlag), &vars); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid variables JSON: %v\n", err)
			os.Exit(1)
		}

		req := common.VariablesSetRequest{Variables: vars}
		var resp common.VariablesResponse
		if err := client.postJSON("/api/variables", &req, &resp); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Variables set:")
		for k, v := range resp.Variables {
			fmt.Printf("  %s = %v\n", k, v)
		}
		return
	}

	var resp common.VariablesResponse
	if err := client.getJSON("/api/variables", &resp); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(resp.Variables) == 0 {
		fmt.Println("No preset variables")
		return
	}

	fmt.Println("Preset variables:")
	for k, v := range resp.Variables {
		fmt.Printf("  %s = %v\n", k, v)
	}
}

func printUsage() {
	fmt.Println("GraphQL Client CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gqlc <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  query    Execute a GraphQL query")
	fmt.Println("  file     Execute query from .gql file")
	fmt.Println("  lint     Validate query syntax only")
	fmt.Println("  format   Format/pretty-print query")
	fmt.Println("  schema   Show server schema")
	fmt.Println("  variables  Manage preset variables")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  GRAPHQL_SERVER_URL  Server URL (default: http://localhost:8080)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gqlc query '{ user(id: \"1\") { name } }'")
	fmt.Println("  gqlc file query.gql")
	fmt.Println("  gqlc lint '{ user { name } }'")
	fmt.Println("  gqlc format '{user{name}}'")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "query":
		cmdQuery(args)
	case "file":
		cmdFile(args)
	case "lint":
		cmdLint(args)
	case "format":
		cmdFormat(args)
	case "schema":
		cmdSchema(args)
	case "variables":
		cmdVariables(args)
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
