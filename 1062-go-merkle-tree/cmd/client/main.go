package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "build":
		runBuild()
	case "verify":
		runVerify()
	case "diff":
		runDiff()
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Merkle Tree Client - File Synchronization Tool

Usage:
  client build <file>                    Compute file hashes and build Merkle tree on server
  client verify <file> <index>           Verify a specific block's hash
  client diff <source-tree-id> <file>    Find differences between source tree and target file
  client help                            Show this help message

Examples:
  client build /path/to/file
  client verify /path/to/file 0
  client diff abc123def456 /path/to/other/file`)
}

func runBuild() {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	serverAddr := fs.String("server", "http://localhost:8080", "Server address")
	fs.Parse(os.Args[2:])

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: file path required")
		os.Exit(1)
	}

	filePath := fs.Arg(0)

	leafHashes, err := computeFileHashes(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	client := NewClient(*serverAddr)
	resp, err := client.BuildTree(leafHashes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building tree: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Success!\n")
	fmt.Printf("Tree ID:    %s\n", resp.TreeID)
	fmt.Printf("Root Hash:  %s\n", resp.RootHash)
	fmt.Printf("Blocks:     %d\n", len(leafHashes))
}

func runVerify() {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	serverAddr := fs.String("server", "http://localhost:8080", "Server address")
	treeID := fs.String("tree-id", "", "Tree ID to verify against (if not provided, builds a new tree)")
	fs.Parse(os.Args[2:])

	if fs.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "Error: file path and block index required")
		os.Exit(1)
	}

	filePath := fs.Arg(0)
	index := 0
	_, err := fmt.Sscanf(fs.Arg(1), "%d", &index)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid block index: %v\n", err)
		os.Exit(1)
	}

	leafHashes, err := computeFileHashes(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	if index < 0 || index >= len(leafHashes) {
		fmt.Fprintf(os.Stderr, "Error: block index %d out of range [0, %d]\n", index, len(leafHashes)-1)
		os.Exit(1)
	}

	client := NewClient(*serverAddr)

	var targetTreeID string
	if *treeID != "" {
		targetTreeID = *treeID
	} else {
		buildResp, err := client.BuildTree(leafHashes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error building tree: %v\n", err)
			os.Exit(1)
		}
		targetTreeID = buildResp.TreeID
	}

	verifyResp, err := client.VerifyProof(targetTreeID, index, leafHashes[index], leafHashes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying proof: %v\n", err)
		os.Exit(1)
	}

	if verifyResp.Valid {
		fmt.Printf("Verification PASSED: Block %d hash is correct\n", index)
	} else {
		fmt.Printf("Verification FAILED: Block %d hash does not match\n", index)
	}
}

func runDiff() {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	serverAddr := fs.String("server", "http://localhost:8080", "Server address")
	fs.Parse(os.Args[2:])

	if fs.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "Error: source tree ID and target file path required")
		os.Exit(1)
	}

	sourceTreeID := fs.Arg(0)
	targetFilePath := fs.Arg(1)

	leafHashes, err := computeFileHashes(targetFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	client := NewClient(*serverAddr)
	diffResp, err := client.FindDifferences(sourceTreeID, leafHashes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding differences: %v\n", err)
		os.Exit(1)
	}

	if len(diffResp.Differences) == 0 {
		fmt.Println("No differences found - files are identical")
	} else {
		fmt.Printf("Found %d differing block(s):\n", len(diffResp.Differences))
		for _, idx := range diffResp.Differences {
			fmt.Printf("  - Block %d\n", idx)
		}
	}
}
