package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/solo-coder/douglas-peucker/internal/api"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8501", "Server URL")
	mode := flag.String("mode", "fixed", "Threshold mode: fixed or percentage")
	threshold := flag.Float64("threshold", 10.0, "Distance threshold (meters) or percentage")
	isClosed := flag.Bool("closed", false, "Is closed loop")
	inputFile := flag.String("input", "", "Input JSON file")
	outputFile := flag.String("output", "", "Output JSON file")
	minDist := flag.Float64("min-distance", 0, "Minimum distance between points")
	maxDist := flag.Float64("max-distance", 0, "Maximum distance between points")
	flag.Parse()

	if *inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: input file is required")
		flag.Usage()
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	var points []api.Point
	if err := json.Unmarshal(data, &points); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing input JSON: %v\n", err)
		os.Exit(1)
	}

	req := api.SimplifyRequest{
		Points:   points,
		Mode:     api.ThresholdMode(*mode),
		Threshold: *threshold,
		IsClosed: *isClosed,
	}

	if *minDist > 0 || *maxDist > 0 {
		req.UniformOptions = &api.UniformOptions{
			MinDistance: *minDist,
			MaxDistance: *maxDist,
		}
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling request: %v\n", err)
		os.Exit(1)
	}

	url := *serverURL + "/simplify"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Server error: %s\n", body)
		os.Exit(1)
	}

	var result api.SimplifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Original points: %d\n", result.OriginalCount)
	fmt.Printf("Simplified points: %d\n", result.SimplifiedCount)
	fmt.Printf("Reduction rate: %.2f%%\n", result.ReductionRate)
	fmt.Printf("Area deviation: %.4f%%\n", result.AreaDeviation)
	fmt.Printf("Total length: %.2f meters\n", result.TotalLength)

	outputData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling output: %v\n", err)
		os.Exit(1)
	}

	if *outputFile != "" {
		if err := ioutil.WriteFile(*outputFile, outputData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Result written to: %s\n", *outputFile)
	} else {
		fmt.Println("\nSimplified points:")
		for i, p := range result.Points {
			fmt.Printf("  %d: %.6f, %.6f\n", i, p.Lat, p.Lon)
		}
	}
}
