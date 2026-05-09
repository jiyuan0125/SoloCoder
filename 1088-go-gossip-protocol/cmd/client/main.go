package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/example/gossip-simulator/pkg/api"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

func (c *Client) postJSON(endpoint string, body interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	
	url := c.baseURL + endpoint
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return ioutil.ReadAll(resp.Body)
}

func (c *Client) getJSON(endpoint string) ([]byte, error) {
	url := c.baseURL + endpoint
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return ioutil.ReadAll(resp.Body)
}

func (c *Client) AddNode(nodeID string) error {
	req := api.AddNodeRequest{NodeID: nodeID}
	resp, err := c.postJSON("/nodes/add", req)
	if err != nil {
		return err
	}
	
	var result api.SuccessResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	
	if !result.Success {
		var errResp api.ErrorResponse
		json.Unmarshal(resp, &errResp)
		return fmt.Errorf(errResp.Error)
	}
	
	fmt.Printf("✓ %s\n", result.Message)
	return nil
}

func (c *Client) RemoveNode(nodeID string) error {
	req := api.RemoveNodeRequest{NodeID: nodeID}
	resp, err := c.postJSON("/nodes/remove", req)
	if err != nil {
		return err
	}
	
	var result api.SuccessResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	
	if !result.Success {
		var errResp api.ErrorResponse
		json.Unmarshal(resp, &errResp)
		return fmt.Errorf(errResp.Error)
	}
	
	fmt.Printf("✓ %s\n", result.Message)
	return nil
}

func (c *Client) InjectData(nodeID, key, value string) error {
	req := api.InjectDataRequest{
		NodeID: nodeID,
		Key:    key,
		Value:  value,
	}
	resp, err := c.postJSON("/data/inject", req)
	if err != nil {
		return err
	}
	
	var result api.SuccessResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	
	if !result.Success {
		var errResp api.ErrorResponse
		json.Unmarshal(resp, &errResp)
		return fmt.Errorf(errResp.Error)
	}
	
	fmt.Printf("✓ %s\n", result.Message)
	return nil
}

func (c *Client) Step(steps int) error {
	req := api.StepRequest{Steps: steps}
	resp, err := c.postJSON("/step", req)
	if err != nil {
		return err
	}
	
	var result api.StepResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	
	fmt.Printf("✓ Executed %d step(s)\n", result.StepsDone)
	fmt.Printf("  Current consistency: %.2f%%\n", result.Consistency)
	return nil
}

func (c *Client) GetState() error {
	resp, err := c.getJSON("/state")
	if err != nil {
		return err
	}
	
	var result api.StateResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	
	fmt.Printf("Round: %d\n\n", result.Round)
	fmt.Println("Nodes:")
	for nodeID, state := range result.Nodes {
		fmt.Printf("\n  %s [%s]\n", nodeID, state.Status)
		if len(state.Storage) > 0 {
			fmt.Println("    Storage:")
			for k, v := range state.Storage {
				fmt.Printf("      %s = %s (v%d)\n", k, v.Value, v.Version)
			}
		} else {
			fmt.Println("    Storage: (empty)")
		}
	}
	return nil
}

func (c *Client) GetConsistency() error {
	resp, err := c.getJSON("/consistency")
	if err != nil {
		return err
	}
	
	var result api.ConsistencyResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	
	fmt.Printf("Current consistency: %.2f%%\n", result.Consistency)
	return nil
}

func (c *Client) SimulateFailure(nodeID string) error {
	req := api.SimulateFailureRequest{NodeID: nodeID}
	resp, err := c.postJSON("/nodes/fail", req)
	if err != nil {
		return err
	}
	
	var result api.SuccessResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	
	if !result.Success {
		var errResp api.ErrorResponse
		json.Unmarshal(resp, &errResp)
		return fmt.Errorf(errResp.Error)
	}
	
	fmt.Printf("✓ %s\n", result.Message)
	return nil
}

func (c *Client) SimulateRecovery(nodeID string) error {
	req := api.SimulateRecoveryRequest{NodeID: nodeID}
	resp, err := c.postJSON("/nodes/recover", req)
	if err != nil {
		return err
	}
	
	var result api.SuccessResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	
	if !result.Success {
		var errResp api.ErrorResponse
		json.Unmarshal(resp, &errResp)
		return fmt.Errorf(errResp.Error)
	}
	
	fmt.Printf("✓ %s\n", result.Message)
	return nil
}

func printUsage() {
	fmt.Println("Gossip Protocol Simulator Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gossip-client <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  add-node <node-id>           Add a new node")
	fmt.Println("  remove-node <node-id>        Remove a node")
	fmt.Println("  inject <node-id> <key> <value>  Inject data into a node")
	fmt.Println("  step [count]                 Advance simulation (default 1 step)")
	fmt.Println("  state                        Show current state")
	fmt.Println("  consistency                  Show current consistency percentage")
	fmt.Println("  fail <node-id>               Simulate node failure")
	fmt.Println("  recover <node-id>            Simulate node recovery")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  GOSSIP_SERVER_URL            Server URL (default: http://localhost:8080)")
}

func main() {
	serverURL := os.Getenv("GOSSIP_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}
	
	client := NewClient(serverURL)
	
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	
	cmd := os.Args[1]
	args := os.Args[2:]
	
	var err error
	
	switch cmd {
	case "add-node":
		if len(args) < 1 {
			fmt.Println("Error: node-id required")
			printUsage()
			os.Exit(1)
		}
		err = client.AddNode(args[0])
		
	case "remove-node":
		if len(args) < 1 {
			fmt.Println("Error: node-id required")
			printUsage()
			os.Exit(1)
		}
		err = client.RemoveNode(args[0])
		
	case "inject":
		if len(args) < 3 {
			fmt.Println("Error: node-id, key, and value required")
			printUsage()
			os.Exit(1)
		}
		err = client.InjectData(args[0], args[1], args[2])
		
	case "step":
		steps := 1
		if len(args) > 0 {
			s, parseErr := strconv.Atoi(args[0])
			if parseErr != nil {
				fmt.Printf("Error: invalid step count: %v\n", parseErr)
				os.Exit(1)
			}
			steps = s
		}
		err = client.Step(steps)
		
	case "state":
		err = client.GetState()
		
	case "consistency":
		err = client.GetConsistency()
		
	case "fail":
		if len(args) < 1 {
			fmt.Println("Error: node-id required")
			printUsage()
			os.Exit(1)
		}
		err = client.SimulateFailure(args[0])
		
	case "recover":
		if len(args) < 1 {
			fmt.Println("Error: node-id required")
			printUsage()
			os.Exit(1)
		}
		err = client.SimulateRecovery(args[0])
		
	default:
		fmt.Printf("Error: unknown command '%s'\n", cmd)
		printUsage()
		os.Exit(1)
	}
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
