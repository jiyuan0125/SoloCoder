package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/example/crdt/pkg/api"
)

func printUsage() {
	fmt.Println("CRDT Client - Command Line Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  crdt-client [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -server <url>    Server URL (default: http://localhost:8080)")
	fmt.Println()
	fmt.Println("Interactive Commands:")
	fmt.Println("  create <type>          Create a new CRDT instance (type: gcounter, pncounter, gset)")
	fmt.Println("  list                   List all instances")
	fmt.Println("  get <id>               Get instance details")
	fmt.Println("  inc <id> <node-id>     Increment a G-Counter or PN-Counter")
	fmt.Println("  dec <id> <node-id>     Decrement a PN-Counter")
	fmt.Println("  add <id> <element>     Add element to G-Set")
	fmt.Println("  merge <target> <source> Merge source into target")
	fmt.Println("  demo                   Run a demonstration")
	fmt.Println("  help                   Show this help message")
	fmt.Println("  exit                   Exit the program")
}

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	flag.Parse()

	client := NewClient(*serverURL)

	fmt.Println("CRDT Client")
	fmt.Println("Server:", *serverURL)
	fmt.Println()

	if err := client.HealthCheck(); err != nil {
		fmt.Printf("Warning: Cannot connect to server: %v\n", err)
	} else {
		fmt.Println("Connected to server successfully!")
	}

	fmt.Println()
	printUsage()
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "exit", "quit":
			fmt.Println("Goodbye!")
			return

		case "help":
			printUsage()

		case "create":
			if len(parts) < 2 {
				fmt.Println("Usage: create <type>")
				continue
			}
			handleCreate(client, parts[1])

		case "list":
			handleList(client)

		case "get":
			if len(parts) < 2 {
				fmt.Println("Usage: get <id>")
				continue
			}
			handleGet(client, parts[1])

		case "inc":
			if len(parts) < 3 {
				fmt.Println("Usage: inc <id> <node-id> [delta]")
				continue
			}
			handleInc(client, parts)

		case "dec":
			if len(parts) < 3 {
				fmt.Println("Usage: dec <id> <node-id> [delta]")
				continue
			}
			handleDec(client, parts)

		case "add":
			if len(parts) < 3 {
				fmt.Println("Usage: add <id> <element>")
				continue
			}
			handleAdd(client, parts[1], strings.Join(parts[2:], " "))

		case "merge":
			if len(parts) < 3 {
				fmt.Println("Usage: merge <target-id> <source-id>")
				continue
			}
			handleMerge(client, parts[1], parts[2])

		case "demo":
			runDemo(client)

		default:
			fmt.Printf("Unknown command: %s\n", cmd)
			fmt.Println("Type 'help' for available commands.")
		}
	}
}

func handleCreate(client *Client, typeStr string) {
	var crdtType api.CRDTType
	switch strings.ToLower(typeStr) {
	case "gcounter":
		crdtType = api.TypeGCounter
	case "pncounter":
		crdtType = api.TypePNCounter
	case "gset":
		crdtType = api.TypeGSet
	default:
		fmt.Printf("Unknown type: %s\n", typeStr)
		return
	}

	resp, err := client.CreateInstance(crdtType)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Created instance:")
	fmt.Printf("  ID:   %s\n", resp.ID)
	fmt.Printf("  Type: %s\n", resp.Type)
}

func handleList(client *Client) {
	resp, err := client.ListInstances()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(resp) == 0 {
		fmt.Println("No instances found.")
		return
	}

	fmt.Println("Instances:")
	for _, inst := range resp {
		fmt.Printf("  ID: %s, Type: %s", inst.ID, inst.Type)
		if inst.Type == api.TypeGSet {
			fmt.Printf(", Elements: %v", inst.Elements)
		} else {
			fmt.Printf(", Value: %d", inst.Value)
		}
		fmt.Println()
	}
}

func handleGet(client *Client, id string) {
	resp, err := client.GetInstance(id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Instance %s:\n", resp.ID)
	fmt.Printf("  Type: %s\n", resp.Type)
	if resp.Type == api.TypeGSet {
		fmt.Printf("  Elements: %v\n", resp.Elements)
	} else {
		fmt.Printf("  Value: %d\n", resp.Value)
	}
	fmt.Printf("  State: %v\n", resp.State)
}

func handleInc(client *Client, parts []string) {
	id := parts[1]
	nodeID := parts[2]

	req := api.OperationRequest{
		Operation: "increment",
		NodeID:    nodeID,
	}

	if len(parts) >= 4 {
		var delta int
		fmt.Sscanf(parts[3], "%d", &delta)
		req.Delta = delta
	}

	resp, err := client.ExecuteOperation(id, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Increment successful. Value: %d\n", resp.Value)
}

func handleDec(client *Client, parts []string) {
	id := parts[1]
	nodeID := parts[2]

	req := api.OperationRequest{
		Operation: "decrement",
		NodeID:    nodeID,
	}

	if len(parts) >= 4 {
		var delta int
		fmt.Sscanf(parts[3], "%d", &delta)
		req.Delta = delta
	}

	resp, err := client.ExecuteOperation(id, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Decrement successful. Value: %d\n", resp.Value)
}

func handleAdd(client *Client, id, element string) {
	req := api.OperationRequest{
		Operation: "add",
		Element:   element,
	}

	resp, err := client.ExecuteOperation(id, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Add successful. Elements: %v\n", resp.Elements)
}

func handleMerge(client *Client, targetID, sourceID string) {
	resp, err := client.MergeInstances(targetID, sourceID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Merge successful:")
	fmt.Printf("  Target: %s\n", resp.TargetID)
	fmt.Printf("  Source: %s\n", resp.SourceID)
	if resp.Elements != nil {
		fmt.Printf("  Elements: %v\n", resp.Elements)
	} else {
		fmt.Printf("  Value: %d\n", resp.Value)
	}
}

func runDemo(client *Client) {
	fmt.Println("=== CRDT Demo ===")
	fmt.Println()

	fmt.Println("1. G-Counter Demo:")
	fmt.Println("   Creating two G-Counter instances for nodes A and B...")

	counterA, err := client.CreateInstance(api.TypeGCounter)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	counterB, err := client.CreateInstance(api.TypeGCounter)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("   Counter A ID: %s\n", counterA.ID)
	fmt.Printf("   Counter B ID: %s\n", counterB.ID)

	fmt.Println("\n   Incrementing Counter A 3 times on node 'alice'...")
	for i := 0; i < 3; i++ {
		client.ExecuteOperation(counterA.ID, api.OperationRequest{
			Operation: "increment",
			NodeID:    "alice",
		})
	}

	fmt.Println("   Incrementing Counter B 5 times on node 'bob'...")
	for i := 0; i < 5; i++ {
		client.ExecuteOperation(counterB.ID, api.OperationRequest{
			Operation: "increment",
			NodeID:    "bob",
		})
	}

	instA, _ := client.GetInstance(counterA.ID)
	instB, _ := client.GetInstance(counterB.ID)

	fmt.Printf("   Counter A value: %d\n", instA.Value)
	fmt.Printf("   Counter B value: %d\n", instB.Value)

	fmt.Println("\n   Merging B into A...")
	mergeResp, _ := client.MergeInstances(counterA.ID, counterB.ID)
	fmt.Printf("   Counter A after merge: %d\n", mergeResp.Value)

	fmt.Println("\n2. PN-Counter Demo:")
	fmt.Println("   Creating a PN-Counter...")

	pnCounter, err := client.CreateInstance(api.TypePNCounter)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("   Increment 10 times, then decrement 3 times on node 'charlie'...")
	for i := 0; i < 10; i++ {
		client.ExecuteOperation(pnCounter.ID, api.OperationRequest{
			Operation: "increment",
			NodeID:    "charlie",
		})
	}
	for i := 0; i < 3; i++ {
		client.ExecuteOperation(pnCounter.ID, api.OperationRequest{
			Operation: "decrement",
			NodeID:    "charlie",
		})
	}

	instPN, _ := client.GetInstance(pnCounter.ID)
	fmt.Printf("   PN-Counter value: %d (expected: 7)\n", instPN.Value)

	fmt.Println("\n3. G-Set Demo:")
	fmt.Println("   Creating two G-Set instances...")

	setA, _ := client.CreateInstance(api.TypeGSet)
	setB, _ := client.CreateInstance(api.TypeGSet)

	fmt.Println("   Adding 'apple', 'banana' to Set A...")
	client.ExecuteOperation(setA.ID, api.OperationRequest{Operation: "add", Element: "apple"})
	client.ExecuteOperation(setA.ID, api.OperationRequest{Operation: "add", Element: "banana"})

	fmt.Println("   Adding 'banana', 'cherry' to Set B...")
	client.ExecuteOperation(setB.ID, api.OperationRequest{Operation: "add", Element: "banana"})
	client.ExecuteOperation(setB.ID, api.OperationRequest{Operation: "add", Element: "cherry"})

	instSetA, _ := client.GetInstance(setA.ID)
	instSetB, _ := client.GetInstance(setB.ID)

	fmt.Printf("   Set A elements: %v\n", instSetA.Elements)
	fmt.Printf("   Set B elements: %v\n", instSetB.Elements)

	fmt.Println("\n   Merging B into A...")
	mergeSetResp, _ := client.MergeInstances(setA.ID, setB.ID)
	fmt.Printf("   Set A after merge: %v (expected union of both)\n", mergeSetResp.Elements)

	fmt.Println("\n=== Demo Complete ===")
}
