package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"decisiontree/pkg/common"
)

type clientOptions struct {
	serverURL     string
	trainFile     string
	predictFile   string
	outputFile    string
	maxDepth      int
	minSamples    int
	handleMissing string
	treeID        string
	action        string
	exportFormat  string
}

func readCSVFile(filepath string) (string, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", filepath, err)
	}
	return string(data), nil
}

func parseCSVForPredict(csvData string) ([][]string, error) {
	reader := csv.NewReader(strings.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %v", err)
	}

	if len(records) < 1 {
		return nil, fmt.Errorf("CSV is empty")
	}

	hasHeader := true
	if len(records) > 0 {
		firstRow := records[0]
		for _, val := range firstRow {
			if _, err := csv.NewReader(strings.NewReader(val)).ReadAll(); err == nil {
			}
		}
	}

	data := make([][]string, 0)
	startIdx := 0
	if hasHeader {
		startIdx = 1
	}

	for _, row := range records[startIdx:] {
		data = append(data, row)
	}

	return data, nil
}

func trainTree(opts *clientOptions) (string, error) {
	if opts.trainFile == "" {
		return "", fmt.Errorf("train file is required for training")
	}

	csvData, err := readCSVFile(opts.trainFile)
	if err != nil {
		return "", err
	}

	req := common.TrainRequest{
		CSVData:       csvData,
		MaxDepth:      opts.maxDepth,
		MinSamples:    opts.minSamples,
		HandleMissing: opts.handleMissing,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(opts.serverURL+"/train", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != "" {
			return "", fmt.Errorf("server error: %s", errResp.Error)
		}
		return "", fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var trainResp common.TrainResponse
	if err := json.NewDecoder(resp.Body).Decode(&trainResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if !trainResp.Success {
		return "", fmt.Errorf("training failed: %s", trainResp.Message)
	}

	return trainResp.TreeID, nil
}

func predictTree(opts *clientOptions) ([]string, error) {
	if opts.treeID == "" {
		return nil, fmt.Errorf("tree ID is required for prediction")
	}

	if opts.predictFile == "" {
		return nil, fmt.Errorf("predict file is required for prediction")
	}

	csvData, err := readCSVFile(opts.predictFile)
	if err != nil {
		return nil, err
	}

	data, err := parseCSVForPredict(csvData)
	if err != nil {
		return nil, err
	}

	req := common.PredictRequest{
		TreeID: opts.treeID,
		Data:   data,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(opts.serverURL+"/predict", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != "" {
			return nil, fmt.Errorf("server error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var predResp common.PredictResponse
	if err := json.NewDecoder(resp.Body).Decode(&predResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if !predResp.Success {
		return nil, fmt.Errorf("prediction failed: %s", predResp.Message)
	}

	return predResp.Labels, nil
}

func exportTree(opts *clientOptions) (string, string, error) {
	if opts.treeID == "" {
		return "", "", fmt.Errorf("tree ID is required for export")
	}

	req := common.TreeExportRequest{
		TreeID: opts.treeID,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(opts.serverURL+"/export", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != "" {
			return "", "", fmt.Errorf("server error: %s", errResp.Error)
		}
		return "", "", fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var exportResp common.TreeExportResponse
	if err := json.NewDecoder(resp.Body).Decode(&exportResp); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %v", err)
	}

	if !exportResp.Success {
		return "", "", fmt.Errorf("export failed: %s", exportResp.Message)
	}

	return exportResp.TextTree, exportResp.JSONTree, nil
}

func printUsage() {
	fmt.Println(`Decision Tree Client - Command Line Interface

Usage:
  client [flags]

Actions:
  train    Train a new decision tree
  predict  Make predictions using a trained tree
  export   Export the structure of a trained tree

Flags:
  --action string     Action to perform: train, predict, or export (required)
  --train string      Path to training CSV file (required for train action)
  --predict string    Path to prediction CSV file (required for predict action)
  --output string     Path to output file for predictions/export (optional)
  --max-depth int     Maximum tree depth (default: 10)
  --min-samples int   Minimum samples per leaf (default: 1)
  --missing string    Missing value handling: skip or mean (default: skip)
  --tree-id string    Tree ID for predict/export actions
  --format string     Export format: text, json, or both (default: both)
  --server string     Server URL (default: http://localhost:8080)

Examples:
  # Train a tree
  client --action train --train data.csv --max-depth 5

  # Predict using a trained tree
  client --action predict --tree-id tree_1 --predict test.csv --output predictions.txt

  # Export tree structure
  client --action export --tree-id tree_1 --format text --output tree.txt`)
}

func main() {
	opts := &clientOptions{}

	flag.StringVar(&opts.action, "action", "", "Action to perform: train, predict, or export")
	flag.StringVar(&opts.trainFile, "train", "", "Path to training CSV file")
	flag.StringVar(&opts.predictFile, "predict", "", "Path to prediction CSV file")
	flag.StringVar(&opts.outputFile, "output", "", "Path to output file")
	flag.IntVar(&opts.maxDepth, "max-depth", 10, "Maximum tree depth")
	flag.IntVar(&opts.minSamples, "min-samples", 1, "Minimum samples per leaf")
	flag.StringVar(&opts.handleMissing, "missing", "skip", "Missing value handling: skip or mean")
	flag.StringVar(&opts.treeID, "tree-id", "", "Tree ID for predict/export actions")
	flag.StringVar(&opts.exportFormat, "format", "both", "Export format: text, json, or both")
	flag.StringVar(&opts.serverURL, "server", "http://localhost:8080", "Server URL")

	flag.Parse()

	if opts.action == "" {
		printUsage()
		os.Exit(1)
	}

	opts.serverURL = strings.TrimRight(opts.serverURL, "/")

	var err error
	switch opts.action {
	case "train":
		fmt.Println("Training decision tree...")
		treeID, err := trainTree(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Tree trained successfully! Tree ID: %s\n", treeID)

	case "predict":
		fmt.Println("Making predictions...")
		labels, err := predictTree(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		output := "Predictions:\n"
		for i, label := range labels {
			output += fmt.Sprintf("  Row %d: %s\n", i+1, label)
		}

		if opts.outputFile != "" {
			if err := ioutil.WriteFile(opts.outputFile, []byte(output), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Predictions saved to %s\n", opts.outputFile)
		} else {
			fmt.Print(output)
		}

	case "export":
		fmt.Println("Exporting tree structure...")
		textTree, jsonTree, err := exportTree(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var output string
		switch opts.exportFormat {
		case "text":
			output = textTree
		case "json":
			output = jsonTree
		default:
			output = "=== Text Format ===\n" + textTree + "\n=== JSON Format ===\n" + jsonTree
		}

		if opts.outputFile != "" {
			if err := ioutil.WriteFile(opts.outputFile, []byte(output), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Tree exported to %s\n", opts.outputFile)
		} else {
			fmt.Print(output)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown action: %s\n", opts.action)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
