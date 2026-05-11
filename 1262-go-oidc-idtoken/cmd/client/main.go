package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go-oidc-idtoken/pkg/api"
)

const (
	defaultServerURL = "http://localhost:8080"
)

type Client struct {
	serverURL string
}

func NewClient() *Client {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = defaultServerURL
	}
	return &Client{serverURL: serverURL}
}

func (c *Client) post(endpoint string, body interface{}, resp interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := c.serverURL + endpoint
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	respHTTP, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer respHTTP.Body.Close()

	if respHTTP.StatusCode >= 400 {
		body, _ := io.ReadAll(respHTTP.Body)
		return fmt.Errorf("server error %d: %s", respHTTP.StatusCode, string(body))
	}

	if resp != nil {
		return json.NewDecoder(respHTTP.Body).Decode(resp)
	}
	return nil
}

func (c *Client) verify(token string) error {
	req := api.VerifyRequest{Token: token}
	resp := &api.VerifyResponse{}

	if err := c.post("/oidc/verify", req, resp); err != nil {
		return err
	}

	fmt.Printf("Valid: %v\n", resp.Valid)
	if len(resp.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range resp.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	fmt.Println("\nClaims:")
	claimsJSON, _ := json.MarshalIndent(resp.Claims, "", "  ")
	fmt.Println(string(claimsJSON))

	if !resp.Valid {
		os.Exit(1)
	}
	return nil
}

func (c *Client) parse(token string) error {
	req := api.ParseRequest{Token: token}
	resp := &api.ParseResponse{}

	if err := c.post("/oidc/parse", req, resp); err != nil {
		return err
	}

	fmt.Println("Header:")
	headerJSON, _ := json.MarshalIndent(resp.Header, "", "  ")
	fmt.Println(string(headerJSON))

	fmt.Println("\nClaims:")
	claimsJSON, _ := json.MarshalIndent(resp.Claims, "", "  ")
	fmt.Println(string(claimsJSON))

	return nil
}

func (c *Client) setConfig(key, value string) error {
	req := api.SetConfigRequest{Key: key, Value: value}
	resp := &api.SetConfigResponse{}

	if err := c.post("/oidc/set-config", req, resp); err != nil {
		return err
	}

	fmt.Printf("Config set: %s = %s\n", key, value)
	return nil
}

func readTokenFromFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  token verify <jwt-file>")
		fmt.Println("  token parse <jwt-file>")
		fmt.Println("  token config set <key> <value>")
		os.Exit(1)
	}

	client := NewClient()

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "verify":
		if len(args) != 1 {
			fmt.Println("Usage: token verify <jwt-file>")
			os.Exit(1)
		}
		token, err := readTokenFromFile(filepath.Clean(args[0]))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading token file: %v\n", err)
			os.Exit(1)
		}
		if err := client.verify(token); err != nil {
			fmt.Fprintf(os.Stderr, "Verification error: %v\n", err)
			os.Exit(1)
		}

	case "parse":
		if len(args) != 1 {
			fmt.Println("Usage: token parse <jwt-file>")
			os.Exit(1)
		}
		token, err := readTokenFromFile(filepath.Clean(args[0]))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading token file: %v\n", err)
			os.Exit(1)
		}
		if err := client.parse(token); err != nil {
			fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
			os.Exit(1)
		}

	case "config":
		if len(args) < 2 {
			fmt.Println("Usage: token config set <key> <value>")
			os.Exit(1)
		}
		subCmd := args[0]
		if subCmd != "set" {
			fmt.Println("Usage: token config set <key> <value>")
			os.Exit(1)
		}
		if len(args) < 3 {
			fmt.Println("Usage: token config set <key> <value>")
			os.Exit(1)
		}
		key := args[1]
		value := args[2]
		if err := client.setConfig(key, value); err != nil {
			fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage:")
		fmt.Println("  token verify <jwt-file>")
		fmt.Println("  token parse <jwt-file>")
		fmt.Println("  token config set <key> <value>")
		os.Exit(1)
	}
}
