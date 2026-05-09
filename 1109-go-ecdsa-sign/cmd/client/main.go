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

	"ecdsa-sign/pkg/api"
)

const (
	defaultServer = "http://localhost:8080"
)

type client struct {
	serverURL string
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := defaultServer
	if envURL := os.Getenv("ECDSA_SERVER_URL"); envURL != "" {
		serverURL = envURL
	}

	var flagSet *flag.FlagSet
	var message, signature, publicKey string

	switch os.Args[1] {
	case "sign":
		flagSet = flag.NewFlagSet("sign", flag.ExitOnError)
		flagSet.StringVar(&serverURL, "server", serverURL, "server URL")
		flagSet.StringVar(&message, "message", "", "message to sign")
		if err := flagSet.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}
		if message == "" {
			if flagSet.NArg() > 0 {
				message = strings.Join(flagSet.Args(), " ")
			}
		}
		if message == "" {
			fmt.Fprintln(os.Stderr, "Error: message is required")
			os.Exit(1)
		}
		c := &client{serverURL: serverURL}
		if err := c.sign(message); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "verify":
		flagSet = flag.NewFlagSet("verify", flag.ExitOnError)
		flagSet.StringVar(&serverURL, "server", serverURL, "server URL")
		flagSet.StringVar(&message, "message", "", "message to verify")
		flagSet.StringVar(&signature, "signature", "", "signature (base64 encoded)")
		flagSet.StringVar(&publicKey, "pubkey", "", "public key (base64 encoded)")
		if err := flagSet.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}
		if message == "" {
			fmt.Fprintln(os.Stderr, "Error: message is required")
			os.Exit(1)
		}
		if signature == "" {
			fmt.Fprintln(os.Stderr, "Error: signature is required")
			os.Exit(1)
		}
		if publicKey == "" {
			fmt.Fprintln(os.Stderr, "Error: public key is required")
			os.Exit(1)
		}
		c := &client{serverURL: serverURL}
		if err := c.verify(message, signature, publicKey); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "pubkey":
		flagSet = flag.NewFlagSet("pubkey", flag.ExitOnError)
		flagSet.StringVar(&serverURL, "server", serverURL, "server URL")
		if err := flagSet.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}
		c := &client{serverURL: serverURL}
		if err := c.getPublicKey(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("ECDSA Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client sign --message \"hello world\"")
	fmt.Println("  client verify --message \"hello world\" --signature <sig> --pubkey <key>")
	fmt.Println("  client pubkey")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  sign     Sign a message")
	fmt.Println("  verify   Verify a signature")
	fmt.Println("  pubkey   Get the server's public key")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  ECDSA_SERVER_URL  Server URL (default: http://localhost:8080)")
}

func (c *client) sign(message string) error {
	reqBody := &api.SignRequest{Message: message}
	respBody, err := c.post("/sign", reqBody)
	if err != nil {
		return err
	}

	var resp api.SignResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}

	fmt.Printf("Signature:\n%s\n", resp.Signature)
	return nil
}

func (c *client) verify(message, signature, publicKey string) error {
	reqBody := &api.VerifyRequest{
		Message:   message,
		Signature: signature,
		PublicKey: publicKey,
	}
	respBody, err := c.post("/verify", reqBody)
	if err != nil {
		return err
	}

	var resp api.VerifyResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}

	if resp.Valid {
		fmt.Println("Signature is VALID")
	} else {
		fmt.Println("Signature is INVALID")
	}
	return nil
}

func (c *client) getPublicKey() error {
	respBody, err := c.get("/key/public")
	if err != nil {
		return err
	}

	var resp api.PublicKeyResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}

	fmt.Printf("Curve: %s\n", resp.Curve)
	fmt.Printf("Public Key:\n%s\n", resp.PublicKey)
	return nil
}

func (c *client) get(path string) ([]byte, error) {
	url := c.serverURL + path
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: %s", string(body))
	}

	return body, nil
}

func (c *client) post(path string, data interface{}) ([]byte, error) {
	bodyBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.serverURL + path
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: %s", string(respBody))
	}

	return respBody, nil
}
