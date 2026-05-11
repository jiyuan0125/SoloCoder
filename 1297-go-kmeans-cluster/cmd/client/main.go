package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"kmeans-cluster/common"
)

const DefaultServerURL = "http://localhost:8200"

func main() {
	var (
		filePath    string
		k           int
		autoK       bool
		maxK        int
		serverURL   string
		showLabels  bool
		wcss        bool
	)

	flag.StringVar(&filePath, "file", "", "Path to data file (CSV or JSON)")
	flag.IntVar(&k, "k", 0, "Number of clusters")
	flag.BoolVar(&autoK, "auto", false, "Auto-select K using Elbow Method")
	flag.IntVar(&maxK, "max-k", 10, "Maximum K for Elbow Method")
	flag.StringVar(&serverURL, "server", DefaultServerURL, "K-Means server URL")
	flag.BoolVar(&showLabels, "labels", false, "Show cluster labels for each point")
	flag.BoolVar(&wcss, "wcss", false, "Compute WCSS for K=1 to max-k")
	flag.Parse()

	if filePath == "" {
		fmt.Println("Error: --file is required")
		flag.Usage()
		os.Exit(1)
	}

	if wcss {
		if err := computeWCSS(filePath, maxK, serverURL); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if k == 0 && !autoK {
		fmt.Println("Error: either --k or --auto must be specified")
		flag.Usage()
		os.Exit(1)
	}

	if err := runClustering(filePath, k, autoK, maxK, serverURL, showLabels); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func readPointsFromFile(filePath string) ([]common.Point, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	if ext == ".csv" {
		return readCSV(file)
	}

	if ext == ".json" {
		return readJSON(file)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var points []common.Point
	if err := json.Unmarshal(data, &points); err == nil && len(points) > 0 {
		return points, nil
	}

	file.Seek(0, 0)
	return readCSV(file)
}

func readCSV(r io.Reader) ([]common.Point, error) {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %v", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no data in CSV")
	}

	points := make([]common.Point, 0, len(records))
	for i, record := range records {
		point := make(common.Point, 0, len(record))
		for j, val := range record {
			val = strings.TrimSpace(val)
			if val == "" {
				continue
			}
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number '%s' at row %d, column %d: %v", val, i+1, j+1, err)
			}
			point = append(point, f)
		}
		if len(point) == 0 {
			continue
		}
		if len(points) > 0 && len(point) != len(points[0]) {
			return nil, fmt.Errorf("dimension mismatch at row %d: expected %d, got %d", i+1, len(points[0]), len(point))
		}
		points = append(points, point)
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("no valid data points found")
	}

	return points, nil
}

func readJSON(r io.Reader) ([]common.Point, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var req struct {
		Points []common.Point `json:"points"`
	}
	if err := json.Unmarshal(data, &req); err == nil && len(req.Points) > 0 {
		return req.Points, nil
	}

	var points []common.Point
	if err := json.Unmarshal(data, &points); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("no data points in JSON")
	}

	return points, nil
}

func runClustering(filePath string, k int, autoK bool, maxK int, serverURL string, showLabels bool) error {
	points, err := readPointsFromFile(filePath)
	if err != nil {
		return err
	}

	fmt.Printf("Loaded %d data points, dimension: %d\n", len(points), len(points[0]))

	req := common.ClusterRequest{
		Points: points,
	}

	if autoK {
		req.AutoK = true
		req.MaxK = &maxK
		fmt.Printf("Using Elbow Method to find optimal K (max K=%d)...\n", maxK)
	} else {
		req.K = &k
		fmt.Printf("Running K-Means with K=%d...\n", k)
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(serverURL+"/cluster", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result common.ClusterResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("clustering failed: %s", result.Message)
	}

	if result.Message != "" {
		fmt.Println(result.Message)
	}

	fmt.Printf("\n=== Clustering Results ===\n")
	fmt.Printf("K: %d\n", result.K)
	fmt.Printf("Iterations: %d\n", result.Iterations)
	fmt.Printf("Converged: %v\n", result.Converged)

	fmt.Printf("\n=== Cluster Statistics ===\n")
	for _, cluster := range result.Clusters {
		fmt.Printf("\nCluster %d:\n", cluster.Index)
		fmt.Printf("  Point count: %d\n", cluster.PointCount)
		fmt.Printf("  Center: %v\n", formatPoint(cluster.Center))
		fmt.Printf("  Average distance to center: %.6f\n", cluster.AvgDistance)
	}

	if showLabels {
		fmt.Printf("\n=== Point Labels ===\n")
		for i, label := range result.Labels {
			fmt.Printf("Point %d: Cluster %d\n", i, label)
		}
	}

	return nil
}

func computeWCSS(filePath string, maxK int, serverURL string) error {
	points, err := readPointsFromFile(filePath)
	if err != nil {
		return err
	}

	fmt.Printf("Loaded %d data points, dimension: %d\n", len(points), len(points[0]))
	fmt.Printf("Computing WCSS for K=1 to K=%d...\n", maxK)

	req := common.WCSSRequest{
		Points: points,
		MaxK:   maxK,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(serverURL+"/wcss", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result common.WCSSResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("WCSS computation failed: %s", result.Message)
	}

	fmt.Printf("\n=== WCSS Results ===\n")
	fmt.Println("K | WCSS")
	fmt.Println("---|---------")
	for k := 1; k <= maxK; k++ {
		if wcss, ok := result.WCSS[k]; ok {
			fmt.Printf("%d | %.6f\n", k, wcss)
		}
	}

	fmt.Println("\nUse these values to plot the elbow curve and determine the optimal K.")
	return nil
}

func formatPoint(p common.Point) string {
	parts := make([]string, len(p))
	for i, v := range p {
		parts[i] = fmt.Sprintf("%.6f", v)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
