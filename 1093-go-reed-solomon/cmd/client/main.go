package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/solocoder/reed-solomon/pkg/api"
)

const serverURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "encode":
		encodeCmd()
	case "decode":
		decodeCmd()
	case "status":
		statusCmd()
	case "simulate":
		simulateCmd()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: client <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  encode    - Encode a file")
	fmt.Println("  decode    - Decode and recover data")
	fmt.Println("  status    - Check server status")
	fmt.Println("  simulate  - Simulate shard loss and test recovery")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client encode -f input.txt -n 4 -m 2")
	fmt.Println("  client decode -b <block_id> -d ./shards")
	fmt.Println("  client status")
	fmt.Println("  client simulate -b <block_id> -k 2")
}

func encodeCmd() {
	fs := flag.NewFlagSet("encode", flag.ExitOnError)
	filePath := fs.String("f", "", "Input file path")
	n := fs.Int("n", 4, "Number of data shards")
	m := fs.Int("m", 2, "Number of parity shards")
	fs.Parse(os.Args[2:])

	if *filePath == "" {
		fmt.Println("Error: -f is required")
		fs.Usage()
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(*filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	req := api.EncodeRequest{
		Data: string(data),
		N:    *n,
		M:    *m,
	}

	var resp api.EncodeResponse
	err = postJSON(serverURL+"/api/encode", req, &resp)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Encode successful!\n")
	fmt.Printf("Block ID: %s\n", resp.BlockID)
	fmt.Printf("N=%d, M=%d, Total shards=%d\n", resp.N, resp.M, len(resp.Shards))
	fmt.Printf("Original length: %d bytes\n", resp.OriginalLen)
	fmt.Println("\nShards:")
	for _, s := range resp.Shards {
		fmt.Printf("  Index %d: ID=%s\n", s.Index, s.ID)
	}
}

func decodeCmd() {
	fs := flag.NewFlagSet("decode", flag.ExitOnError)
	blockID := fs.String("b", "", "Block ID to decode")
	shardDir := fs.String("d", "", "Directory to save recovered data")
	outputFile := fs.String("o", "recovered", "Output file name")
	fs.Parse(os.Args[2:])

	if *blockID == "" {
		fmt.Println("Error: -b is required")
		fs.Usage()
		os.Exit(1)
	}

	var statusResp api.StatusResponse
	err := getJSON(serverURL+"/api/status", &statusResp)
	if err != nil {
		fmt.Printf("Error getting status: %v\n", err)
		os.Exit(1)
	}

	var targetBlock *api.BlockStatus
	for i := range statusResp.Blocks {
		if statusResp.Blocks[i].BlockID == *blockID {
			targetBlock = &statusResp.Blocks[i]
			break
		}
	}

	if targetBlock == nil {
		fmt.Printf("Error: Block %s not found\n", *blockID)
		os.Exit(1)
	}

	var availableShards []string
	for _, s := range targetBlock.Shards {
		if s.IsAvailable {
			availableShards = append(availableShards, s.ShardID)
		}
	}

	fmt.Printf("Block %s has %d available shards (need %d)\n", 
		*blockID, len(availableShards), targetBlock.N)

	req := api.DecodeRequest{
		BlockID:  *blockID,
		ShardIDs: availableShards,
	}

	var decodeResp api.DecodeResponse
	err = postJSON(serverURL+"/api/decode", req, &decodeResp)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !decodeResp.Success {
		fmt.Printf("Decode failed: %s\n", decodeResp.Error)
		os.Exit(1)
	}

	outputPath := *outputFile
	if *shardDir != "" {
		if err := os.MkdirAll(*shardDir, 0755); err != nil {
			fmt.Printf("Error creating directory: %v\n", err)
			os.Exit(1)
		}
		outputPath = filepath.Join(*shardDir, *outputFile)
	}

	if err := ioutil.WriteFile(outputPath, []byte(decodeResp.Data), 0644); err != nil {
		fmt.Printf("Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Decode successful!\n")
	fmt.Printf("Recovered data saved to: %s\n", outputPath)
	fmt.Printf("Data size: %d bytes\n", len(decodeResp.Data))
}

func statusCmd() {
	var resp api.StatusResponse
	err := getJSON(serverURL+"/api/status", &resp)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(resp.Blocks) == 0 {
		fmt.Println("No blocks stored on server.")
		return
	}

	fmt.Printf("Total blocks: %d\n\n", len(resp.Blocks))

	for _, block := range resp.Blocks {
		fmt.Printf("Block ID: %s\n", block.BlockID)
		fmt.Printf("  N=%d, M=%d\n", block.N, block.M)
		fmt.Printf("  Shards: %d available / %d total\n", 
			block.AvailableShards, block.TotalShards)
		fmt.Println("  Shard status:")
		for _, s := range block.Shards {
			status := "AVAILABLE"
			if !s.IsAvailable {
				status = "LOST"
			}
			fmt.Printf("    Index %d: ID=%s [%s]\n", s.Index, s.ShardID, status)
		}
		fmt.Println()
	}
}

func simulateCmd() {
	fs := flag.NewFlagSet("simulate", flag.ExitOnError)
	blockID := fs.String("b", "", "Block ID to simulate")
	k := fs.Int("k", 1, "Number of shards to lose")
	iterations := fs.Int("i", 100, "Number of simulation iterations")
	fs.Parse(os.Args[2:])

	if *blockID == "" {
		fmt.Println("Error: -b is required")
		fs.Usage()
		os.Exit(1)
	}

	var statusResp api.StatusResponse
	err := getJSON(serverURL+"/api/status", &statusResp)
	if err != nil {
		fmt.Printf("Error getting status: %v\n", err)
		os.Exit(1)
	}

	var targetBlock *api.BlockStatus
	for i := range statusResp.Blocks {
		if statusResp.Blocks[i].BlockID == *blockID {
			targetBlock = &statusResp.Blocks[i]
			break
		}
	}

	if targetBlock == nil {
		fmt.Printf("Error: Block %s not found\n", *blockID)
		os.Exit(1)
	}

	if *k >= targetBlock.TotalShards {
		fmt.Printf("Error: Cannot lose all %d shards\n", targetBlock.TotalShards)
		os.Exit(1)
	}

	fmt.Printf("Simulating loss of %d shards (need %d to recover, total %d)\n", 
		*k, targetBlock.N, targetBlock.TotalShards)
	fmt.Printf("Running %d iterations...\n\n", *iterations)

	rand.Seed(time.Now().UnixNano())
	successCount := 0
	failCount := 0

	allShardIDs := make([]string, 0, len(targetBlock.Shards))
	for _, s := range targetBlock.Shards {
		allShardIDs = append(allShardIDs, s.ShardID)
	}

	for i := 0; i < *iterations; i++ {
		shuffled := make([]string, len(allShardIDs))
		copy(shuffled, allShardIDs)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		keptShards := shuffled[:len(shuffled)-*k]

		req := api.DecodeRequest{
			BlockID:  *blockID,
			ShardIDs: keptShards,
		}

		var decodeResp api.DecodeResponse
		err = postJSON(serverURL+"/api/decode", req, &decodeResp)
		if err != nil {
			fmt.Printf("Iteration %d: Error - %v\n", i+1, err)
			failCount++
			continue
		}

		if decodeResp.Success {
			successCount++
		} else {
			failCount++
		}
	}

	successRate := float64(successCount) / float64(*iterations) * 100

	fmt.Println("=== Simulation Results ===")
	fmt.Printf("Total iterations: %d\n", *iterations)
	fmt.Printf("Successful recoveries: %d\n", successCount)
	fmt.Printf("Failed recoveries: %d\n", failCount)
	fmt.Printf("Success rate: %.2f%%\n", successRate)

	if *k > targetBlock.M {
		fmt.Printf("\nNote: Losing %d shards exceeds M=%d, recovery is not guaranteed.\n", 
			*k, targetBlock.M)
	} else {
		fmt.Printf("\nNote: Losing %d shards is within M=%d, recovery should always succeed.\n", 
			*k, targetBlock.M)
	}
}

func postJSON(url string, req interface{}, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpResp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(httpResp.Body)
		return fmt.Errorf("HTTP %d: %s", httpResp.StatusCode, string(respBody))
	}

	return json.NewDecoder(httpResp.Body).Decode(resp)
}

func getJSON(url string, resp interface{}) error {
	httpResp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(httpResp.Body)
		return fmt.Errorf("HTTP %d: %s", httpResp.StatusCode, string(respBody))
	}

	return json.NewDecoder(httpResp.Body).Decode(resp)
}
