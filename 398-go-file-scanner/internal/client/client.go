package client

import (
	"encoding/json"
	"fmt"
	"net"

	"filescanner/internal/protocol"
)

type Client struct {
	serverAddr string
	conn       net.Conn
}

func NewClient(serverAddr string) *Client {
	return &Client{
		serverAddr: serverAddr,
	}
}

func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	c.conn = conn
	return nil
}

func (c *Client) Disconnect() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

func (c *Client) SendScanRequest(req *protocol.ScanRequest) (*protocol.ScanResponse, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to server")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scan request: %w", err)
	}

	msg := &protocol.Message{
		Type:    protocol.TypeScanRequest,
		Payload: string(payload),
	}

	if err := protocol.EncodeMessage(c.conn, msg); err != nil {
		return nil, fmt.Errorf("failed to send scan request: %w", err)
	}

	responseMsg, err := protocol.DecodeMessage(c.conn)
	if err != nil {
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	switch responseMsg.Type {
	case protocol.TypeScanResponse:
		var resp protocol.ScanResponse
		if err := json.Unmarshal([]byte(responseMsg.Payload), &resp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal scan response: %w", err)
		}
		return &resp, nil
	case protocol.TypeError:
		return nil, fmt.Errorf("server error: %s", responseMsg.Payload)
	default:
		return nil, fmt.Errorf("unexpected response type: %s", responseMsg.Type)
	}
}

func PrintResults(resp *protocol.ScanResponse, outputJSON bool) {
	if outputJSON {
		printJSONResults(resp)
	} else {
		printTextResults(resp)
	}
}

func printJSONResults(resp *protocol.ScanResponse) {
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Printf("Error: Failed to marshal JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func printTextResults(resp *protocol.ScanResponse) {
	fmt.Println("=")
	fmt.Println("Scan Results")
	fmt.Println("=")
	fmt.Println()

	if len(resp.Results) == 0 {
		fmt.Println("No sensitive information found.")
	} else {
		for _, result := range resp.Results {
			fmt.Printf("File: %s\n", result.FilePath)
			fmt.Println("---")

			for _, match := range result.Matches {
				fmt.Printf("  Line %d [%s] (%s):\n", match.LineNumber, match.Severity, match.RuleName)
				fmt.Printf("    %s\n", match.MaskedContent)
				fmt.Println()
			}

			fmt.Println()
		}
	}

	fmt.Println("=")
	fmt.Println("Scan Summary")
	fmt.Println("=")
	fmt.Printf("Total files scanned: %d\n", resp.Summary.TotalFilesScanned)
	fmt.Printf("Total files with matches: %d\n", resp.Summary.TotalFilesMatched)
	fmt.Printf("Total matches found: %d\n", resp.Summary.TotalMatchesFound)

	if resp.Error != "" {
		fmt.Println()
		fmt.Printf("Warnings/Errors: %s\n", resp.Error)
	}
}
