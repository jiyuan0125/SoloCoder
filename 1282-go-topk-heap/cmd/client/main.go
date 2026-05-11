package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"topk-service/pkg/api"
)

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "http://" + serverURL
	}
	return &Client{serverURL: serverURL}
}

func (c *Client) Submit(values []float64) error {
	req := api.SubmitRequest{Values: values}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	
	resp, err := http.Post(c.serverURL+"/submit", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("submit failed: %s", string(respBody))
	}
	
	return nil
}

func (c *Client) Query() (*api.QueryResponse, error) {
	resp, err := http.Get(c.serverURL + "/query")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("query failed: %s", string(respBody))
	}
	
	var result api.QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return &result, nil
}

func (c *Client) Clear() error {
	resp, err := http.Post(c.serverURL+"/clear", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("clear failed: %s", string(respBody))
	}
	
	return nil
}

func (c *Client) AdjustK(k int) error {
	req := api.AdjustKRequest{K: k}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	
	resp, err := http.Post(c.serverURL+"/adjust-k", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("adjust-k failed: %s", string(respBody))
	}
	
	return nil
}

func readValuesFromFile(filePath string) ([]float64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var values []float64
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		val, err := strconv.ParseFloat(line, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", line)
		}
		values = append(values, val)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	return values, nil
}

func formatResult(resp *api.QueryResponse) string {
	if len(resp.Data) == 0 {
		return "No data found."
	}
	
	var builder strings.Builder
	builder.WriteString("Top K Elements (by frequency, lowest first):\n")
	builder.WriteString("---------------------------------------------\n")
	builder.WriteString(fmt.Sprintf("%-10s %-20s %-10s\n", "Rank", "Element", "Frequency"))
	builder.WriteString("---------------------------------------------\n")
	
	for i, item := range resp.Data {
		var elemStr string
		if item.Element == float64(int64(item.Element)) {
			elemStr = fmt.Sprintf("%d", int64(item.Element))
		} else {
			elemStr = fmt.Sprintf("%.6f", item.Element)
		}
		builder.WriteString(fmt.Sprintf("%-10d %-20s %-10d\n", i+1, elemStr, item.Freq))
	}
	
	builder.WriteString("---------------------------------------------\n")
	return builder.String()
}

func main() {
	serverURL := flag.String("server", "http://localhost:8509", "Server URL")
	command := flag.String("command", "", "Command: submit|query|clear|adjust-k")
	filePath := flag.String("file", "", "File to submit (for submit command)")
	k := flag.Int("k", 10, "K value (for adjust-k command)")
	batchSize := flag.Int("batch-size", 1000, "Batch size for submit")
	flag.Parse()
	
	if *command == "" {
		fmt.Println("Please specify a command: submit|query|clear|adjust-k")
		os.Exit(1)
	}
	
	client := NewClient(*serverURL)
	
	switch *command {
	case "submit":
		if *filePath == "" {
			fmt.Println("Please specify --file for submit command")
			os.Exit(1)
		}
		
		values, err := readValuesFromFile(*filePath)
		if err != nil {
			fmt.Printf("Failed to read file: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("Read %d values from file, submitting in batches of %d...\n", len(values), *batchSize)
		
		for i := 0; i < len(values); i += *batchSize {
			end := i + *batchSize
			if end > len(values) {
				end = len(values)
			}
			
			batch := values[i:end]
			if err := client.Submit(batch); err != nil {
				fmt.Printf("Failed to submit batch %d-%d: %v\n", i+1, end, err)
				os.Exit(1)
			}
			
			fmt.Printf("Submitted batch %d-%d (%d values)\n", i+1, end, len(batch))
		}
		
		fmt.Println("All data submitted successfully!")
		
	case "query":
		resp, err := client.Query()
		if err != nil {
			fmt.Printf("Query failed: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println(formatResult(resp))
		
	case "clear":
		if err := client.Clear(); err != nil {
			fmt.Printf("Clear failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Data cleared successfully!")
		
	case "adjust-k":
		if err := client.AdjustK(*k); err != nil {
			fmt.Printf("Adjust K failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("K value adjusted to %d successfully!\n", *k)
		
	default:
		fmt.Printf("Unknown command: %s\n", *command)
		fmt.Println("Available commands: submit|query|clear|adjust-k")
		os.Exit(1)
	}
}
