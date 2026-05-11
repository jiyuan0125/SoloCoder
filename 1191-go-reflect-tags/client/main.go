package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect-tags/api"
	"reflect-tags/tagparser"
	"strings"
)

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: strings.TrimSuffix(serverURL, "/"),
	}
}

func (c *Client) Register(def tagparser.StructDef) error {
	req := api.RegisterRequest{Struct: def}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request failed: %w", err)
	}

	resp, err := http.Post(c.serverURL+"/register", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return fmt.Errorf("server error: %s", errResp.Error)
	}

	var regResp api.RegisterResponse
	if err := json.Unmarshal(respBody, &regResp); err != nil {
		return fmt.Errorf("unmarshal response failed: %w", err)
	}

	fmt.Println(regResp.Message)
	return nil
}

func (c *Client) Query(structName, fieldPath string) error {
	req := api.QueryRequest{
		StructName: structName,
		FieldPath:  fieldPath,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request failed: %w", err)
	}

	resp, err := http.Post(c.serverURL+"/query", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return fmt.Errorf("server error: %s", errResp.Error)
	}

	var queryResp api.QueryResponse
	if err := json.Unmarshal(respBody, &queryResp); err != nil {
		return fmt.Errorf("unmarshal response failed: %w", err)
	}

	if fieldPath != "" && queryResp.Field != nil {
		printFieldTable(queryResp.Field)
	} else if len(queryResp.All) > 0 {
		printAllFieldsTable(queryResp.All)
	} else {
		fmt.Println("No fields found.")
	}

	return nil
}

func printFieldTable(field *tagparser.FieldTags) {
	fmt.Printf("%-30s %-20s\n", "Field Path:", field.Path)
	fmt.Println("-----------------------------------------------------------------")

	if field.JSON != nil {
		fmt.Printf("%-30s %-20s\n", "JSON Name:", field.JSON.Name)
		fmt.Printf("%-30s %-20v\n", "JSON OmitEmpty:", field.JSON.OmitEmpty)
		fmt.Printf("%-30s %-20v\n", "JSON String:", field.JSON.String)
	}

	if field.DB != nil {
		fmt.Printf("%-30s %-20s\n", "DB Column:", field.DB.Column)
		fmt.Printf("%-30s %-20s\n", "DB Index Type:", field.DB.IndexType)
	}

	if field.Validate != nil {
		fmt.Printf("%-30s %-20v\n", "Validate Required:", field.Validate.Required)
		if field.Validate.Min != nil {
			fmt.Printf("%-30s %-20g\n", "Validate Min:", *field.Validate.Min)
		}
		if field.Validate.Max != nil {
			fmt.Printf("%-30s %-20g\n", "Validate Max:", *field.Validate.Max)
		}
		fmt.Printf("%-30s %-20v\n", "Validate Email:", field.Validate.Email)
		if field.Validate.Regex != "" {
			fmt.Printf("%-30s %-20s\n", "Validate Regex:", field.Validate.Regex)
		}
	}
}

func printAllFieldsTable(fields []tagparser.FieldTags) {
	fmt.Printf("%-20s %-20s %-20s %-20s\n", "Field Path", "JSON", "DB", "Validate")
	fmt.Println(strings.Repeat("-", 80))

	for _, field := range fields {
		jsonDesc := "-"
		if field.JSON != nil {
			jsonDesc = field.JSON.Name
			opts := []string{}
			if field.JSON.OmitEmpty {
				opts = append(opts, "omitempty")
			}
			if field.JSON.String {
				opts = append(opts, "string")
			}
			if len(opts) > 0 {
				jsonDesc += " (" + strings.Join(opts, ",") + ")"
			}
		}

		dbDesc := "-"
		if field.DB != nil {
			dbDesc = field.DB.Column
			if field.DB.IndexType != "" {
				dbDesc += " [" + field.DB.IndexType + "]"
			}
		}

		valDesc := "-"
		if field.Validate != nil {
			rules := []string{}
			if field.Validate.Required {
				rules = append(rules, "required")
			}
			if field.Validate.Min != nil {
				rules = append(rules, fmt.Sprintf("min=%.0f", *field.Validate.Min))
			}
			if field.Validate.Max != nil {
				rules = append(rules, fmt.Sprintf("max=%.0f", *field.Validate.Max))
			}
			if field.Validate.Email {
				rules = append(rules, "email")
			}
			if field.Validate.Regex != "" {
				rules = append(rules, "regex")
			}
			if len(rules) > 0 {
				valDesc = strings.Join(rules, ",")
			}
		}

		fmt.Printf("%-20s %-20s %-20s %-20s\n", truncate(field.Path, 18), truncate(jsonDesc, 18), truncate(dbDesc, 18), truncate(valDesc, 18))
	}
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-1] + "…"
	}
	return s
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  client register [--server <url>]")
	fmt.Println("  client query [--server <url>] --struct <name> [--field <path>]")
	fmt.Println("")
	fmt.Println("Subcommands:")
	fmt.Println("  register  Register a struct (read JSON from stdin)")
	fmt.Println("  query     Query struct field tags")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --server  Server URL (default: http://localhost:8080)")
	fmt.Println("  --struct  Struct name")
	fmt.Println("  --field   Field path (optional, query all if not specified)")
	os.Exit(1)
}

func getServerURL() string {
	if url := os.Getenv("SERVER_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	subcommand := os.Args[1]
	args := os.Args[2:]

	switch subcommand {
	case "register":
		registerFlagSet := flag.NewFlagSet("register", flag.ExitOnError)
		serverURL := registerFlagSet.String("server", getServerURL(), "server URL")
		registerFlagSet.Parse(args)

		var def tagparser.StructDef
		if err := json.NewDecoder(os.Stdin).Decode(&def); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading JSON from stdin: %v\n", err)
			os.Exit(1)
		}

		client := NewClient(*serverURL)
		if err := client.Register(def); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "query":
		queryFlagSet := flag.NewFlagSet("query", flag.ExitOnError)
		serverURL := queryFlagSet.String("server", getServerURL(), "server URL")
		structName := queryFlagSet.String("struct", "", "struct name")
		fieldPath := queryFlagSet.String("field", "", "field path")
		queryFlagSet.Parse(args)

		if *structName == "" {
			usage()
		}

		client := NewClient(*serverURL)
		if err := client.Query(*structName, *fieldPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcommand)
		usage()
	}
}
