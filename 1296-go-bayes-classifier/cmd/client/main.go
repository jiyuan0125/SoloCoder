package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/bayes-classifier/common"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	return &Client{baseURL: baseURL}
}

func (c *Client) Train(samples []common.TrainingSample) (*common.TrainResponse, error) {
	reqBody := common.TrainRequest{Samples: samples}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.baseURL+"/train", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result common.TrainResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

func (c *Client) Classify(text string) (*common.ClassifyResponse, error) {
	reqBody := common.ClassifyRequest{Text: text}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.baseURL+"/classify", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result common.ClassifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

func (c *Client) ExportModel() (*common.ExportModelResponse, error) {
	resp, err := http.Post(c.baseURL+"/export", "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result common.ExportModelResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

func (c *Client) ImportModel(modelFile string) (*common.ImportModelResponse, error) {
	reqBody := common.ImportModelRequest{ModelFile: modelFile}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.baseURL+"/import", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result common.ImportModelResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

func readTrainingData(filePath string) ([]common.TrainingSample, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var samples []common.TrainingSample
	if err := json.Unmarshal(data, &samples); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return samples, nil
}

func printHelp() {
	fmt.Println("Bayes Classifier CLI Client")
	fmt.Println("Usage: client [command] [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  train --file <path>         Train with data from JSON file")
	fmt.Println("  classify --text <text>      Classify text")
	fmt.Println("  classify --file <path>      Classify text from file")
	fmt.Println("  export                       Export model to server")
	fmt.Println("  import --model <path>       Import model from file on server")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --server <url>              Server URL (default: http://localhost:8105)")
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	serverURL := flag.String("server", "http://localhost:8105", "Server URL")
	trainFile := flag.String("file", "", "Path to training data JSON file")
	classifyText := flag.String("text", "", "Text to classify")
	modelFile := flag.String("model", "", "Path to model file for import")

	command := os.Args[1]

	flag.CommandLine.Parse(os.Args[2:])

	client := NewClient(*serverURL)

	switch command {
	case "train":
		if *trainFile == "" {
			fmt.Println("Error: --file is required for train command")
			os.Exit(1)
		}

		samples, err := readTrainingData(*trainFile)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		resp, err := client.Train(samples)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if resp.Success {
			fmt.Printf("Success! Trained with %d samples\n", resp.TotalDocs)
			if resp.Message != "" {
				fmt.Printf("Message: %s\n", resp.Message)
			}
		} else {
			fmt.Printf("Error: %s\n", resp.Message)
		}

	case "classify":
		text := *classifyText

		if text == "" && *trainFile != "" {
			data, err := os.ReadFile(*trainFile)
			if err != nil {
				fmt.Printf("Error reading file: %v\n", err)
				os.Exit(1)
			}
			text = string(data)
		}

		if text == "" {
			fmt.Println("Error: --text or --file is required for classify command")
			os.Exit(1)
		}

		resp, err := client.Classify(text)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if !resp.Success {
			fmt.Printf("Error: %s\n", resp.Message)
			os.Exit(1)
		}

		fmt.Printf("Best label: %s\n", resp.BestLabel)
		fmt.Println("\nProbability ranking:")
		for i, result := range resp.Results {
			fmt.Printf("  %d. %s: %.4f\n", i+1, result.Label, result.Prob)
		}

	case "export":
		resp, err := client.ExportModel()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if resp.Success {
			fmt.Printf("Success! Model exported to: %s\n", resp.ModelFile)
			if resp.Message != "" {
				fmt.Printf("Message: %s\n", resp.Message)
			}
		} else {
			fmt.Printf("Error: %s\n", resp.Message)
		}

	case "import":
		if *modelFile == "" {
			fmt.Println("Error: --model is required for import command")
			os.Exit(1)
		}

		resp, err := client.ImportModel(*modelFile)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if resp.Success {
			fmt.Println("Success! Model imported")
			if resp.Message != "" {
				fmt.Printf("Message: %s\n", resp.Message)
			}
		} else {
			fmt.Printf("Error: %s\n", resp.Message)
		}

	case "help", "-h", "--help":
		printHelp()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}
