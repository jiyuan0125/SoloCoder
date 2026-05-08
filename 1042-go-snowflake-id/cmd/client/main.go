package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"snowflake-id/pkg/common"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Generate(count int) (*common.GenerateResponse, error) {
	u := fmt.Sprintf("%s/generate", c.baseURL)
	
	if count <= 1 {
		resp, err := c.client.Get(u)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		
		var result common.GenerateResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}
		
		return &result, nil
	}
	
	params := url.Values{}
	params.Set("count", strconv.Itoa(count))
	u = fmt.Sprintf("%s?%s", u, params.Encode())
	
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var result common.GenerateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	
	return &result, nil
}

func (c *Client) Parse(id int64, epoch int64) (*common.ParseResponse, error) {
	u := fmt.Sprintf("%s/parse", c.baseURL)
	
	params := url.Values{}
	params.Set("id", strconv.FormatInt(id, 10))
	if epoch > 0 {
		params.Set("epoch", strconv.FormatInt(epoch, 10))
	}
	u = fmt.Sprintf("%s?%s", u, params.Encode())
	
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var result common.ParseResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	
	return &result, nil
}

func (c *Client) Status() (*common.StatusResponse, error) {
	u := fmt.Sprintf("%s/status", c.baseURL)
	
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var result common.StatusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	
	return &result, nil
}

func prettyJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func formatRealTimestamp(ts int64) string {
	t := time.UnixMilli(ts)
	return t.Format("2006-01-02 15:04:05.000")
}

func main() {
	var (
		serverURL  string
		command    string
		count      int
		id         int64
		epoch      int64
		jsonOutput bool
	)

	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Snowflake ID server URL")
	flag.StringVar(&command, "command", "generate", "Command: generate, batch, parse, status")
	flag.IntVar(&count, "count", 1, "Number of IDs to generate (for batch command)")
	flag.Int64Var(&id, "id", 0, "ID to parse (for parse command)")
	flag.Int64Var(&epoch, "epoch", 0, "Custom epoch for ID parsing")
	flag.BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	flag.Parse()

	client := NewClient(serverURL)

	switch command {
	case "generate":
		resp, err := client.Generate(1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating ID: %v\n", err)
			os.Exit(1)
		}
		
		if !resp.Success {
			fmt.Fprintf(os.Stderr, "Server error: %s\n", resp.Error)
			os.Exit(1)
		}
		
		if jsonOutput {
			output, _ := prettyJSON(resp)
			fmt.Println(output)
		} else {
			fmt.Println(resp.ID)
		}

	case "batch":
		if count <= 0 {
			fmt.Fprintf(os.Stderr, "Count must be greater than 0\n")
			os.Exit(1)
		}
		
		resp, err := client.Generate(count)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating IDs: %v\n", err)
			os.Exit(1)
		}
		
		if !resp.Success {
			fmt.Fprintf(os.Stderr, "Server error: %s\n", resp.Error)
			if len(resp.IDs) > 0 {
				fmt.Fprintf(os.Stderr, "Generated %d partial IDs:\n", len(resp.IDs))
			}
			os.Exit(1)
		}
		
		if jsonOutput {
			output, _ := prettyJSON(resp)
			fmt.Println(output)
		} else {
			for _, id := range resp.IDs {
				fmt.Println(id)
			}
		}

	case "parse":
		if id <= 0 {
			fmt.Fprintf(os.Stderr, "ID must be provided for parse command (use -id flag)\n")
			os.Exit(1)
		}
		
		resp, err := client.Parse(id, epoch)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing ID: %v\n", err)
			os.Exit(1)
		}
		
		if !resp.Success {
			fmt.Fprintf(os.Stderr, "Server error: %s\n", resp.Error)
			os.Exit(1)
		}
		
		if jsonOutput {
			output, _ := prettyJSON(resp)
			fmt.Println(output)
		} else {
			fmt.Printf("ID: %d\n", resp.ID)
			fmt.Printf("Timestamp (relative): %d\n", resp.Timestamp)
			fmt.Printf("Timestamp (absolute): %d\n", resp.RealTimestamp)
			fmt.Printf("Timestamp (formatted): %s\n", formatRealTimestamp(resp.RealTimestamp))
			fmt.Printf("Machine ID: %d\n", resp.MachineID)
			fmt.Printf("Sequence: %d\n", resp.Sequence)
		}

	case "status":
		resp, err := client.Status()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting status: %v\n", err)
			os.Exit(1)
		}
		
		if !resp.Success {
			fmt.Fprintf(os.Stderr, "Server error\n")
			os.Exit(1)
		}
		
		if jsonOutput {
			output, _ := prettyJSON(resp)
			fmt.Println(output)
		} else {
			fmt.Printf("Server Status:\n")
			fmt.Printf("  Last Timestamp: %d\n", resp.LastTimestamp)
			fmt.Printf("  Current Timestamp: %d\n", resp.CurrentTimestamp)
			fmt.Printf("  Machine ID: %d\n", resp.MachineID)
			fmt.Printf("  Last Sequence: %d\n", resp.LastSequence)
			fmt.Printf("  Epoch: %d\n", resp.Epoch)
			fmt.Printf("  Max Sequence: %d\n", resp.MaxSequence)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Fprintf(os.Stderr, "Valid commands: generate, batch, parse, status\n")
		os.Exit(1)
	}
}

func (c *Client) post(endpoint string, req interface{}, resp interface{}) error {
	u := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	
	var body io.Reader
	if req != nil {
		reqBody, err := json.Marshal(req)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(reqBody)
	}
	
	httpResp, err := c.client.Post(u, "application/json", body)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(respBody, resp)
}
