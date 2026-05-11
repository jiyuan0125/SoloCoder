package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"

	"jsonschema/common"
	"jsonschema/schema"
)

const (
	defaultServerURL = "http://localhost:8504"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "register":
		handleRegister()
	case "validate":
		handleValidate()
	default:
		fmt.Printf("unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("JSON Schema Validator Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client register --name <name> --schema <file.json>")
	fmt.Println("  client validate --local --schema <schema.json> --data <data.json>")
	fmt.Println("  client validate --schema-name <name> --data <data.json>")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --local          Validate locally without server")
	fmt.Println("  --name           Schema name for registration")
	fmt.Println("  --schema         Path to schema JSON file")
	fmt.Println("  --schema-name    Name of registered schema on server")
	fmt.Println("  --data           Path to data JSON file to validate")
	fmt.Println("  --server         Server URL (default: http://localhost:8504)")
}

func handleRegister() {
	fs := flag.NewFlagSet("register", flag.ExitOnError)
	name := fs.String("name", "", "schema name")
	schemaFile := fs.String("schema", "", "path to schema file")
	serverURL := fs.String("server", defaultServerURL, "server URL")
	fs.Parse(os.Args[2:])

	if *name == "" || *schemaFile == "" {
		fmt.Println("error: --name and --schema are required")
		os.Exit(1)
	}

	schemaData, err := os.ReadFile(*schemaFile)
	if err != nil {
		fmt.Printf("error: failed to read schema file: %v\n", err)
		os.Exit(1)
	}

	req := common.RegisterRequest{
		Name:   *name,
		Schema: json.RawMessage(schemaData),
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("error: failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(*serverURL+"/register", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("error: server returned status %d\n", resp.StatusCode)
		fmt.Println(string(respBody))
		os.Exit(1)
	}

	var registerResp common.RegisterResponse
	if err := json.Unmarshal(respBody, &registerResp); err != nil {
		fmt.Printf("error: failed to parse response: %v\n", err)
		fmt.Println(string(respBody))
		os.Exit(1)
	}

	if registerResp.Success {
		fmt.Printf("Schema registered successfully: %s\n", *name)
	} else {
		fmt.Printf("Registration failed: %s\n", registerResp.Message)
		os.Exit(1)
	}
}

func handleValidate() {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	local := fs.Bool("local", false, "validate locally")
	schemaName := fs.String("schema-name", "", "name of schema on server")
	schemaFile := fs.String("schema", "", "path to schema file (for local validation)")
	dataFile := fs.String("data", "", "path to data file")
	serverURL := fs.String("server", defaultServerURL, "server URL")
	fs.Parse(os.Args[2:])

	if *dataFile == "" {
		fmt.Println("error: --data is required")
		os.Exit(1)
	}

	dataData, err := os.ReadFile(*dataFile)
	if err != nil {
		fmt.Printf("error: failed to read data file: %v\n", err)
		os.Exit(1)
	}

	if *local {
		if *schemaFile == "" {
			fmt.Println("error: --schema is required for local validation")
			os.Exit(1)
		}
		validateLocal(*schemaFile, dataData)
	} else {
		if *schemaName == "" {
			fmt.Println("error: --schema-name is required for remote validation")
			os.Exit(1)
		}
		validateRemote(*serverURL, *schemaName, dataData)
	}
}

func validateLocal(schemaFile string, dataData []byte) {
	schemaData, err := os.ReadFile(schemaFile)
	if err != nil {
		fmt.Printf("error: failed to read schema file: %v\n", err)
		os.Exit(1)
	}

	validator, err := schema.NewValidator(schemaData)
	if err != nil {
		fmt.Printf("error: invalid schema: %v\n", err)
		os.Exit(1)
	}

	errors := validator.Validate(dataData)
	printErrors(errors)
}

func validateRemote(serverURL, schemaName string, dataData []byte) {
	req := common.ValidateRequest{
		SchemaName: schemaName,
		Data:       json.RawMessage(dataData),
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("error: failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/validate", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("error: server returned status %d\n", resp.StatusCode)
		fmt.Println(string(respBody))
		os.Exit(1)
	}

	var validateResp common.ValidateResponse
	if err := json.Unmarshal(respBody, &validateResp); err != nil {
		fmt.Printf("error: failed to parse response: %v\n", err)
		fmt.Println(string(respBody))
		os.Exit(1)
	}

	if validateResp.Valid {
		fmt.Println("Validation passed!")
		os.Exit(0)
	}

	errors := make([]schema.ValidationError, 0, len(validateResp.Errors))
	for _, e := range validateResp.Errors {
		errors = append(errors, schema.ValidationError{
			Path:       e.Path,
			Constraint: e.Constraint,
			Actual:     e.Actual,
			Expected:   e.Expected,
		})
	}

	printErrors(errors)
}

func printErrors(errors []schema.ValidationError) {
	if len(errors) == 0 {
		fmt.Println("Validation passed!")
		os.Exit(0)
	}

	sort.Slice(errors, func(i, j int) bool {
		return errors[i].Path < errors[j].Path
	})

	fmt.Printf("Validation failed with %d error(s):\n\n", len(errors))
	for _, e := range errors {
		fmt.Println(e.String())
	}
	os.Exit(1)
}
