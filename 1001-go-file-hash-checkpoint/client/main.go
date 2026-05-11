package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"hash-checkpoint/common"
	"hash-checkpoint/core"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := strings.ToLower(os.Args[1])
	args := os.Args[2:]

	switch cmd {
	case "check":
		runCheck(args)
	case "progress":
		runProgress(args)
	case "verify":
		runVerify(args)
	default:
		fmt.Printf("unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("hash-checkpoint client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client check <file> [flags]    - Start or resume file hash check")
	fmt.Println("  client progress <file> [flags] - Check current progress")
	fmt.Println("  client verify <file> [flags]   - Verify file integrity")
	fmt.Println()
	fmt.Println("Common flags:")
	fmt.Println("  -server <addr>   Server address (default: localhost:8100)")
	fmt.Println("  -algo <md5|sha256>  Hash algorithm (default: sha256)")
	fmt.Println("  -chunk <size>    Chunk size in bytes (default: 4194304)")
}

func parseCommonFlags(args []string) (string, string, int64) {
	fs := flag.NewFlagSet("common", flag.ExitOnError)
	server := fs.String("server", "localhost:8100", "server address")
	algo := fs.String("algo", "sha256", "hash algorithm (md5|sha256)")
	chunkSize := fs.Int64("chunk", common.DefaultChunkSize, "chunk size in bytes")
	fs.Parse(args)
	return *server, *algo, *chunkSize
}

func getFileInfo(filePath string) (int64, time.Time, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, time.Time{}, err
	}
	return info.Size(), info.ModTime(), nil
}

func getAlgorithm(name string) common.HashAlgorithm {
	switch strings.ToLower(name) {
	case "md5":
		return common.AlgorithmMD5
	default:
		return common.AlgorithmSHA256
	}
}

func runCheck(args []string) {
	if len(args) < 1 {
		fmt.Println("error: missing file path")
		os.Exit(1)
	}
	filePath := args[0]
	args = args[1:]

	serverAddr, algoName, chunkSize := parseCommonFlags(args)
	algorithm := getAlgorithm(algoName)

	fileSize, modTime, err := getFileInfo(filePath)
	if err != nil {
		fmt.Printf("error: cannot access file: %v\n", err)
		os.Exit(1)
	}

	client := NewAPIClient(serverAddr)

	fmt.Printf("Starting check for: %s\n", filePath)
	fmt.Printf("  File size: %d bytes\n", fileSize)
	fmt.Printf("  Algorithm: %s\n", algorithm)
	fmt.Printf("  Chunk size: %d bytes\n", chunkSize)

	startResp, err := client.StartCheck(filePath, modTime, fileSize, algorithm, chunkSize)
	if err != nil {
		fmt.Printf("error: start check failed: %v\n", err)
		os.Exit(1)
	}

	startIndex := startResp.NextChunkIndex
	fmt.Printf("\n%s\n", startResp.Message)
	fmt.Printf("Starting from chunk %d / %d\n", startIndex, startResp.Progress.TotalChunks)

	reader, err := core.NewFileChunkReader(filePath, chunkSize)
	if err != nil {
		fmt.Printf("error: cannot open file: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	totalChunks := startResp.Progress.TotalChunks
	currentIndex := startIndex

	for currentIndex < totalChunks {
		chunkHash, chunkLen, err := reader.CalculateChunkHash(currentIndex, algorithm)
		if err != nil {
			fmt.Printf("\nerror: failed to calculate chunk %d: %v\n", currentIndex, err)
			os.Exit(1)
		}

		submitResp, err := client.SubmitChunk(filePath, modTime, currentIndex, chunkHash, chunkLen)
		if err != nil {
			fmt.Printf("\nerror: failed to submit chunk %d: %v\n", currentIndex, err)
			os.Exit(1)
		}

		progress := submitResp.Progress
		fmt.Printf("\rProgress: %.1f%% (%d/%d chunks)",
			progress.Percentage, progress.CompletedChunks, progress.TotalChunks)

		if submitResp.Progress.Status == common.StatusChunkMismatch {
			fmt.Printf("\nChunk %d hash mismatch, restarting from this chunk\n", currentIndex)
		}

		currentIndex = submitResp.NextChunkIndex
	}

	fmt.Println()

	completeResp, err := client.CompleteCheck(filePath, modTime)
	if err != nil {
		fmt.Printf("error: failed to complete check: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nCheck completed!\n")
	fmt.Printf("Server final hash (%s): %s\n", algorithm, completeResp.FinalHash)

	localFinalHash, err := core.CalculateFinalHashFromChunks(filePath, chunkSize, algorithm)
	if err != nil {
		fmt.Printf("error: failed to calculate local hash: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Client final hash (%s): %s\n", algorithm, localFinalHash)

	if completeResp.FinalHash == localFinalHash {
		fmt.Println("\n✓ Integrity check PASSED")
	} else {
		fmt.Println("\n✗ Integrity check FAILED")
		os.Exit(1)
	}
}

func runProgress(args []string) {
	if len(args) < 1 {
		fmt.Println("error: missing file path")
		os.Exit(1)
	}
	filePath := args[0]
	args = args[1:]

	serverAddr, _, _ := parseCommonFlags(args)

	fileSize, modTime, err := getFileInfo(filePath)
	if err != nil {
		fmt.Printf("error: cannot access file: %v\n", err)
		os.Exit(1)
	}

	client := NewAPIClient(serverAddr)
	resp, err := client.GetProgress(filePath, modTime)
	if err != nil {
		fmt.Printf("error: failed to get progress: %v\n", err)
		os.Exit(1)
	}

	p := resp.Progress
	fmt.Printf("Progress for: %s\n", filePath)
	fmt.Printf("  File size: %d bytes\n", fileSize)
	fmt.Printf("  Algorithm: %s\n", p.Algorithm)
	fmt.Printf("  Status: %s\n", p.Status)
	fmt.Printf("  Progress: %.1f%% (%d/%d chunks)\n",
		p.Percentage, p.CompletedChunks, p.TotalChunks)
	if !p.StartTime.IsZero() {
		fmt.Printf("  Started at: %s\n", p.StartTime.Format(time.RFC3339))
	}
	if !p.CompleteTime.IsZero() {
		fmt.Printf("  Completed at: %s\n", p.CompleteTime.Format(time.RFC3339))
	}
	if p.FinalHash != "" {
		fmt.Printf("  Final hash: %s\n", p.FinalHash)
	}
}

func runVerify(args []string) {
	if len(args) < 1 {
		fmt.Println("error: missing file path")
		os.Exit(1)
	}
	filePath := args[0]
	args = args[1:]

	_, algoName, chunkSize := parseCommonFlags(args)
	algorithm := getAlgorithm(algoName)

	fmt.Printf("Calculating direct whole-file hash for: %s\n", filePath)

	reader, err := core.NewFileChunkReader(filePath, chunkSize)
	if err != nil {
		fmt.Printf("error: cannot open file: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	directHash, err := reader.CalculateWholeFileHash(algorithm)
	if err != nil {
		fmt.Printf("error: failed to calculate hash: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Direct whole-file hash (%s): %s\n", algorithm, directHash)

	chunkedHash, err := core.CalculateFinalHashFromChunks(filePath, chunkSize, algorithm)
	if err != nil {
		fmt.Printf("error: failed to calculate chunked hash: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Chunked-combined hash (%s): %s\n", algorithm, chunkedHash)
}
