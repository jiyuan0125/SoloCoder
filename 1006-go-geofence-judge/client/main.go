package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"geofence/common"
)

var baseURL string

func init() {
	flag.StringVar(&baseURL, "server", "http://localhost:8080", "geofence server URL")
	flag.Parse()
}

func httpPost(path string, req interface{}, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpResp, err := http.Post(baseURL+path, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	if httpResp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("HTTP %d", httpResp.StatusCode)
	}

	if resp != nil {
		return json.Unmarshal(respBody, resp)
	}
	return nil
}

func listFences() {
	var resp common.FenceListResponse
	if err := httpGet("/fences", &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("\n=== Fences List ===")
	if len(resp.Fences) == 0 {
		fmt.Println("No fences found.")
		return
	}
	for i, f := range resp.Fences {
		fmt.Printf("%d. ID: %s\n   Name: %s\n   Description: %s\n   Area: %.2f m²\n   Vertices: %d\n",
			i+1, f.ID, f.Name, f.Description, f.Area, f.PointCount)
	}
}

func httpGet(path string, resp interface{}) error {
	httpResp, err := http.Get(baseURL + path)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	if httpResp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("HTTP %d", httpResp.StatusCode)
	}

	return json.Unmarshal(respBody, resp)
}

func createFenceInteractive() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter fence name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Enter fence description: ")
	desc, _ := reader.ReadString('\n')
	desc = strings.TrimSpace(desc)

	fmt.Println("Enter vertices (lat,lng), one per line. Enter blank line when done:")
	var points []common.Point
	for {
		fmt.Print("> ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			fmt.Println("Invalid format, use: lat,lng")
			continue
		}
		lat, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		lng, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err1 != nil || err2 != nil {
			fmt.Println("Invalid numbers")
			continue
		}
		points = append(points, common.Point{Lat: lat, Lng: lng})
	}

	if len(points) < 3 {
		fmt.Println("Error: need at least 3 vertices")
		return
	}

	req := common.FenceCreateRequest{
		Name:        name,
		Description: desc,
		Points:      points,
	}

	var resp common.FenceCreateResponse
	if err := httpPost("/fences", &req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nFence created:\n  ID: %s\n  Name: %s\n  Area: %.2f m²\n", resp.ID, resp.Name, resp.Area)
	if resp.Warning != "" {
		fmt.Printf("  Warning: %s\n", resp.Warning)
	}
}

func judgePointInteractive() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter fence ID (optional, blank for all): ")
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	fmt.Print("Enter point (lat,lng): ")
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	parts := strings.Split(line, ",")
	if len(parts) != 2 {
		fmt.Println("Invalid format")
		return
	}
	lat, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lng, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err1 != nil || err2 != nil {
		fmt.Println("Invalid numbers")
		return
	}

	req := common.JudgeRequest{
		Point:   common.Point{Lat: lat, Lng: lng},
		FenceID: id,
	}

	var resp common.JudgeResponse
	if err := httpPost("/judge", &req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("\n=== Result ===")
	fmt.Printf("Point: %.6f, %.6f\n", resp.Point.Lat, resp.Point.Lng)
	fmt.Printf("Min distance to fence boundary: %.2f m\n", resp.MinDistance)
	if len(resp.Results) > 0 && resp.Results[0].In {
		fmt.Printf("Inside fences:\n")
		for i, name := range resp.Results[0].FenceNames {
			fmt.Printf("  - %s (%s)\n", name, resp.Results[0].FenceIDs[i])
		}
	} else {
		fmt.Println("Not inside any fence")
	}
}

func batchJudgeFromFile() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter input file path: ")
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	var points []common.Point
	scanner := bufio.NewScanner(bytes.NewReader(data))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			fmt.Printf("Warning: line %d invalid format, skipping\n", lineNum)
			continue
		}
		lat, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		lng, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err1 != nil || err2 != nil {
			fmt.Printf("Warning: line %d invalid numbers, skipping\n", lineNum)
			continue
		}
		points = append(points, common.Point{Lat: lat, Lng: lng})
	}

	if len(points) == 0 {
		fmt.Println("No valid points found")
		return
	}

	fmt.Printf("Loaded %d points, judging...\n", len(points))

	req := common.BatchJudgeRequest{Points: points}
	var resp common.BatchJudgeResponse
	if err := httpPost("/batch-judge", &req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("\n=== Batch Results ===")
	inCount := 0
	for i, r := range resp.Results {
		status := "OUT"
		if r.In {
			status = "IN "
			inCount++
		}
		fmt.Printf("%3d. [%s] %.6f, %.6f | dist: %8.2f m", i+1, status, r.Point.Lat, r.Point.Lng, r.MinDistance)
		if r.In && len(r.FenceNames) > 0 {
			fmt.Printf(" | fences: %s", strings.Join(r.FenceNames, ", "))
		}
		fmt.Println()
	}

	total := len(resp.Results)
	if total > 0 {
		fmt.Printf("\nSummary: %d/%d (%.1f%%) inside fences\n", inCount, total, float64(inCount)/float64(total)*100)
	}
}

func printMenu() {
	fmt.Println("\n=== Geofence Client ===")
	fmt.Println("1. List all fences")
	fmt.Println("2. Create new fence")
	fmt.Println("3. Judge single point")
	fmt.Println("4. Batch judge from file")
	fmt.Println("5. Exit")
	fmt.Print("\nSelect option: ")
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Connected to server: %s\n", baseURL)

	for {
		printMenu()
		line, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(line)

		switch choice {
		case "1":
			listFences()
		case "2":
			createFenceInteractive()
		case "3":
			judgePointInteractive()
		case "4":
			batchJudgeFromFile()
		case "5":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid option")
		}
	}
}
