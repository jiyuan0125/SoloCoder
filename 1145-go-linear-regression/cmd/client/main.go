package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/linearregression/pkg/api"
)

const serverURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  lreg fit <json_data>")
		fmt.Println("  lreg predict <x_value>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "fit":
		if len(os.Args) < 3 {
			fmt.Println("Usage: lreg fit <json_data>")
			os.Exit(1)
		}
		handleFit(os.Args[2])
	case "predict":
		if len(os.Args) < 3 {
			fmt.Println("Usage: lreg predict <x_value>")
			os.Exit(1)
		}
		handlePredict(os.Args[2])
	default:
		fmt.Printf("Unknown command: %s", command)
		os.Exit(1)
	}
}

func handleFit(data string) {
	var req api.FitRequest

	var simple [][]float64
	err := json.Unmarshal([]byte(data), &simple)
	if err == nil {
		req.Simple = simple
	} else {
		temp := map[string]interface{}{}
		err := json.Unmarshal([]byte(data), &temp)
		if err != nil {
			fmt.Printf("Invalid JSON: %v\n", err)
			os.Exit(1)
		}

		if features, ok := temp["features"]; ok {
			featuresJSON, _ := json.Marshal(features)
			json.Unmarshal(featuresJSON, &req.Features)
		}
		if target, ok := temp["target"]; ok {
			targetJSON, _ := json.Marshal(target)
			json.Unmarshal(targetJSON, &req.Target)
		}
	}

	if req.Simple == nil && (req.Features == nil || req.Target == nil) {
		fmt.Println("Could not parse input data")
		os.Exit(1)
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/fit", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result api.FitResponse
	json.Unmarshal(respBody, &result)

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	if result.Coefficients != nil {
		fmt.Printf("Intercept: %s\n", formatFloat64(result.Intercept))
		fmt.Printf("Coefficients: %s\n", formatFloat64Slice(result.Coefficients))
	} else {
		fmt.Printf("Slope: %s\n", formatFloat64(result.Slope))
		fmt.Printf("Intercept: %s\n", formatFloat64(result.Intercept))
	}
	fmt.Printf("R²: %s\n", formatFloat64(result.R2))
}

func handlePredict(data string) {
	var req api.PredictRequest

	x, err := strconv.ParseFloat(strings.TrimSpace(data), 64)
	if err == nil {
		req.SimpleX = &x
	} else {
		features := []float64{}
		err := json.Unmarshal([]byte(data), &features)
		if err != nil {
			fmt.Printf("Invalid input: %v\n", err)
			os.Exit(1)
		}
		req.Features = features
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/predict", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result api.PredictResponse
	json.Unmarshal(respBody, &result)

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("Predicted Y: %s\n", formatFloat64(result.Value))
}

func formatFloat64(f api.Float64) string {
	val := float64(f)
	if math.IsNaN(val) {
		return "NaN"
	}
	return fmt.Sprintf("%.6f", val)
}

func formatFloat64Slice(arr []api.Float64) string {
	result := "["
	for i, f := range arr {
		if i > 0 {
			result += " "
		}
		result += formatFloat64(f)
	}
	result += "]"
	return result
}
