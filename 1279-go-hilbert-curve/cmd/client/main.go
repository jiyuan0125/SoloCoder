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

	"github.com/example/hilbert-curve-service/pkg/api"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := "http://localhost:8506"
	if envURL := os.Getenv("HILBERT_SERVER_URL"); envURL != "" {
		serverURL = envURL
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "point-to-index":
		if len(args) != 3 {
			fmt.Println("Usage: point-to-index <order> <x> <y>")
			os.Exit(1)
		}
		pointToIndex(serverURL, args[0], args[1], args[2])
	case "index-to-point":
		if len(args) != 2 {
			fmt.Println("Usage: index-to-point <order> <index>")
			os.Exit(1)
		}
		indexToPoint(serverURL, args[0], args[1])
	case "range-query":
		if len(args) != 5 {
			fmt.Println("Usage: range-query <order> <min_x> <max_x> <min_y> <max_y>")
			os.Exit(1)
		}
		rangeQuery(serverURL, args[0], args[1], args[2], args[3], args[4])
	case "estimate-distance":
		if len(args) != 3 {
			fmt.Println("Usage: estimate-distance <order> <index1> <index2>")
			os.Exit(1)
		}
		estimateDistance(serverURL, args[0], args[1], args[2])
	case "batch-point-to-index":
		if len(args) < 3 {
			fmt.Println("Usage: batch-point-to-index <order> <ascending|descending> <x1,y1> [x2,y2 ...]")
			os.Exit(1)
		}
		batchPointToIndex(serverURL, args[0], args[1], args[2:])
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func pointToIndex(serverURL, orderStr, xStr, yStr string) {
	order, err := strconv.Atoi(orderStr)
	if err != nil {
		fmt.Printf("Invalid order: %s\n", orderStr)
		os.Exit(1)
	}

	x, err := strconv.ParseUint(xStr, 10, 64)
	if err != nil {
		fmt.Printf("Invalid x: %s\n", xStr)
		os.Exit(1)
	}

	y, err := strconv.ParseUint(yStr, 10, 64)
	if err != nil {
		fmt.Printf("Invalid y: %s\n", yStr)
		os.Exit(1)
	}

	req := api.PointToIndexRequest{
		Order: order,
		X:     x,
		Y:     y,
	}

	body, err := sendRequest(serverURL+"/point-to-index", req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	var resp api.PointToIndexResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Printf("Error parsing response: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Point (%d, %d) at order %d maps to index %d\n", x, y, order, resp.Index)
}

func indexToPoint(serverURL, orderStr, indexStr string) {
	order, err := strconv.Atoi(orderStr)
	if err != nil {
		fmt.Printf("Invalid order: %s\n", orderStr)
		os.Exit(1)
	}

	index, err := strconv.ParseUint(indexStr, 10, 64)
	if err != nil {
		fmt.Printf("Invalid index: %s\n", indexStr)
		os.Exit(1)
	}

	req := api.IndexToPointRequest{
		Order: order,
		Index: index,
	}

	body, err := sendRequest(serverURL+"/index-to-point", req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	var resp api.IndexToPointResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Printf("Error parsing response: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Index %d at order %d maps to point (%d, %d)\n", index, order, resp.X, resp.Y)
}

func rangeQuery(serverURL, orderStr, minXStr, maxXStr, minYStr, maxYStr string) {
	order, err := strconv.Atoi(orderStr)
	if err != nil {
		fmt.Printf("Invalid order: %s\n", orderStr)
		os.Exit(1)
	}

	minX, err := strconv.ParseUint(minXStr, 10, 64)
	if err != nil {
		fmt.Printf("Invalid min_x: %s\n", minXStr)
		os.Exit(1)
	}

	maxX, err := strconv.ParseUint(maxXStr, 10, 64)
	if err != nil {
		fmt.Printf("Invalid max_x: %s\n", maxXStr)
		os.Exit(1)
	}

	minY, err := strconv.ParseUint(minYStr, 10, 64)
	if err != nil {
		fmt.Printf("Invalid min_y: %s\n", minYStr)
		os.Exit(1)
	}

	maxY, err := strconv.ParseUint(maxYStr, 10, 64)
	if err != nil {
		fmt.Printf("Invalid max_y: %s\n", maxYStr)
		os.Exit(1)
	}

	req := api.RangeQueryRequest{
		Order: order,
		MinX:  minX,
		MaxX:  maxX,
		MinY:  minY,
		MaxY:  maxY,
	}

	body, err := sendRequest(serverURL+"/range-query", req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	var resp api.RangeQueryResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Printf("Error parsing response: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Range query result for order %d, rectangle [%d, %d]x[%d, %d]:\n", order, minX, maxX, minY, maxY)
	for i, r := range resp.Ranges {
		fmt.Printf("  Range %d: [%d, %d]\n", i+1, r.Min, r.Max)
	}
}

func estimateDistance(serverURL, orderStr, index1Str, index2Str string) {
	order, err := strconv.Atoi(orderStr)
	if err != nil {
		fmt.Printf("Invalid order: %s\n", orderStr)
		os.Exit(1)
	}

	index1, err := strconv.ParseUint(index1Str, 10, 64)
	if err != nil {
		fmt.Printf("Invalid index1: %s\n", index1Str)
		os.Exit(1)
	}

	index2, err := strconv.ParseUint(index2Str, 10, 64)
	if err != nil {
		fmt.Printf("Invalid index2: %s\n", index2Str)
		os.Exit(1)
	}

	req := api.EstimateDistanceRequest{
		Order:  order,
		Index1: index1,
		Index2: index2,
	}

	body, err := sendRequest(serverURL+"/estimate-distance", req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	var resp api.EstimateDistanceResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Printf("Error parsing response: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Estimated distance between index %d and %d at order %d: %d\n", index1, index2, order, resp.Distance)
}

func batchPointToIndex(serverURL, orderStr, sortDirection string, pointsStr []string) {
	order, err := strconv.Atoi(orderStr)
	if err != nil {
		fmt.Printf("Invalid order: %s\n", orderStr)
		os.Exit(1)
	}

	var ascending bool
	switch strings.ToLower(sortDirection) {
	case "ascending":
		ascending = true
	case "descending":
		ascending = false
	default:
		fmt.Printf("Invalid sort direction: %s (must be 'ascending' or 'descending')\n", sortDirection)
		os.Exit(1)
	}

	points := make([]api.Point, 0, len(pointsStr))
	for i, pointStr := range pointsStr {
		parts := strings.Split(pointStr, ",")
		if len(parts) != 2 {
			fmt.Printf("Invalid point format at position %d: %s (expected x,y)\n", i, pointStr)
			os.Exit(1)
		}

		x, err := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			fmt.Printf("Invalid x at position %d: %s\n", i, parts[0])
			os.Exit(1)
		}

		y, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			fmt.Printf("Invalid y at position %d: %s\n", i, parts[1])
			os.Exit(1)
		}

		points = append(points, api.Point{X: x, Y: y})
	}

	req := api.BatchPointToIndexRequest{
		Order:     order,
		Points:    points,
		Ascending: ascending,
	}

	body, err := sendRequest(serverURL+"/batch-point-to-index", req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	var resp api.BatchPointToIndexResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Printf("Error parsing response: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Batch conversion result (%s order):\n", sortDirection)
	for i, index := range resp.Indices {
		fmt.Printf("  Index %d: %d\n", i+1, index)
	}
}

func sendRequest(url string, req interface{}) ([]byte, error) {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %s", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("error sending request: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, fmt.Errorf("server error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("server returned status code %d", resp.StatusCode)
	}

	return body, nil
}

func printUsage() {
	fmt.Println("Hilbert Curve Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  hilbert-client <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  point-to-index <order> <x> <y>              Convert a 2D point to Hilbert index")
	fmt.Println("  index-to-point <order> <index>              Convert a Hilbert index to 2D point")
	fmt.Println("  range-query <order> <min_x> <max_x> <min_y> <max_y>  Query Hilbert index ranges for a rectangle")
	fmt.Println("  estimate-distance <order> <index1> <index2> Estimate distance between two Hilbert indices")
	fmt.Println("  batch-point-to-index <order> <ascending|descending> <x1,y1> [x2,y2 ...]  Batch convert points to indices")
	fmt.Println("  help                                        Show this help message")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  HILBERT_SERVER_URL                          Server URL (default: http://localhost:8506)")
}
