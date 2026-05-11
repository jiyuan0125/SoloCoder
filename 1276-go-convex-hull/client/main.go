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

	"convex-hull/api"
)

func main() {
	serverAddr := flag.String("server", "http://localhost:8504", "Server address")
	command := flag.String("cmd", "", "Command: compute|perimeter-area|point-in-hull|intersection|filter-boundary|weighted|dynamic-add|dynamic-get")
	flag.Parse()

	if *command == "" {
		fmt.Println("Usage: client -cmd <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  compute           - Compute convex hull from points")
		fmt.Println("  perimeter-area    - Compute perimeter and area")
		fmt.Println("  point-in-hull     - Check if point is in convex hull")
		fmt.Println("  intersection      - Compute intersection area of two hulls")
		fmt.Println("  filter-boundary   - Filter boundary points")
		fmt.Println("  weighted          - Compute weighted convex hull")
		fmt.Println("  dynamic-add       - Add points to dynamic convex hull")
		fmt.Println("  dynamic-get       - Get current dynamic convex hull")
		fmt.Println("\nPoint input format: x1,y1 x2,y2 ...")
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)

	switch *command {
	case "compute":
		runCompute(*serverAddr, scanner)
	case "perimeter-area":
		runPerimeterArea(*serverAddr, scanner)
	case "point-in-hull":
		runPointInHull(*serverAddr, scanner)
	case "intersection":
		runIntersection(*serverAddr, scanner)
	case "filter-boundary":
		runFilterBoundary(*serverAddr, scanner)
	case "weighted":
		runWeighted(*serverAddr, scanner)
	case "dynamic-add":
		runDynamicAdd(*serverAddr, scanner)
	case "dynamic-get":
		runDynamicGet(*serverAddr, scanner)
	default:
		fmt.Printf("Unknown command: %s\n", *command)
		os.Exit(1)
	}
}

func runCompute(server string, scanner *bufio.Scanner) {
	fmt.Print("Enter points (x1,y1 x2,y2 ...): ")
	points := readPoints(scanner)
	req := api.ComputeRequest{Points: points}
	var resp api.ComputeResponse
	if err := post(server+"/compute", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	printResponse(resp)
}

func runPerimeterArea(server string, scanner *bufio.Scanner) {
	fmt.Print("Enter points (x1,y1 x2,y2 ...): ")
	points := readPoints(scanner)
	req := api.PerimeterAreaRequest{Points: points}
	var resp api.PerimeterAreaResponse
	if err := post(server+"/perimeter-area", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Success: %v\n", resp.Success)
	fmt.Printf("Type: %s\n", resp.Hull.Type)
	fmt.Printf("Perimeter: %.4f\n", resp.Perimeter)
	fmt.Printf("Area: %.4f\n", resp.Area)
	fmt.Printf("Hull Points: %v\n", resp.Hull.Points)
}

func runPointInHull(server string, scanner *bufio.Scanner) {
	fmt.Print("Enter point to check (x,y): ")
	scanner.Scan()
	ptStr := scanner.Text()
	p := parsePoint(ptStr)

	fmt.Print("Enter hull points (x1,y1 x2,y2 ...): ")
	hullPoints := readPoints(scanner)

	req := api.PointInHullRequest{Point: p, HullPoints: hullPoints}
	var resp api.PointInHullResponse
	if err := post(server+"/point-in-hull", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Success: %v\n", resp.Success)
	fmt.Printf("Inside: %v\n", resp.Inside)
}

func runIntersection(server string, scanner *bufio.Scanner) {
	fmt.Print("Enter hull1 points (x1,y1 x2,y2 ...): ")
	hull1 := readPoints(scanner)

	fmt.Print("Enter hull2 points (x1,y1 x2,y2 ...): ")
	hull2 := readPoints(scanner)

	req := api.IntersectionRequest{Hull1: hull1, Hull2: hull2}
	var resp api.IntersectionResponse
	if err := post(server+"/intersection", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Success: %v\n", resp.Success)
	fmt.Printf("Intersection Type: %s\n", resp.Intersection.Type)
	fmt.Printf("Intersection Points: %v\n", resp.Intersection.Points)
	fmt.Printf("Intersection Area: %.4f\n", resp.IntersectionArea)
}

func runFilterBoundary(server string, scanner *bufio.Scanner) {
	fmt.Print("Enter points (x1,y1 x2,y2 ...): ")
	points := readPoints(scanner)
	req := api.FilterBoundaryRequest{Points: points}
	var resp api.FilterBoundaryResponse
	if err := post(server+"/filter-boundary", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Success: %v\n", resp.Success)
	fmt.Printf("Boundary Points (%d): %v\n", len(resp.Boundary), resp.Boundary)
}

func runWeighted(server string, scanner *bufio.Scanner) {
	fmt.Print("Enter weighted points (x1,y1,w1 x2,y2,w2 ...): ")
	scanner.Scan()
	line := scanner.Text()
	parts := strings.Fields(line)
	var points []api.WeightedPoint
	for _, p := range parts {
		wp := parseWeightedPoint(p)
		if wp != nil {
			points = append(points, *wp)
		}
	}
	req := api.WeightedComputeRequest{Points: points}
	var resp api.ComputeResponse
	if err := post(server+"/weighted", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	printResponse(resp)
}

func runDynamicAdd(server string, scanner *bufio.Scanner) {
	fmt.Print("Enter session ID (empty for new session): ")
	scanner.Scan()
	sessionID := scanner.Text()

	fmt.Print("Enter points to add (x1,y1 x2,y2 ...): ")
	points := readPoints(scanner)

	req := api.DynamicAddRequest{SessionID: sessionID, Points: points}
	var resp api.DynamicAddResponse
	if err := post(server+"/dynamic/add", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Success: %v\n", resp.Success)
	fmt.Printf("Session ID: %s\n", resp.SessionID)
	fmt.Printf("Hull Type: %s\n", resp.Hull.Type)
	fmt.Printf("Hull Points: %v\n", resp.Hull.Points)
}

func runDynamicGet(server string, scanner *bufio.Scanner) {
	fmt.Print("Enter session ID: ")
	scanner.Scan()
	sessionID := scanner.Text()

	req := api.DynamicGetRequest{SessionID: sessionID}
	var resp api.DynamicGetResponse
	if err := post(server+"/dynamic/get", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Success: %v\n", resp.Success)
	fmt.Printf("Hull Type: %s\n", resp.Hull.Type)
	fmt.Printf("Hull Points: %v\n", resp.Hull.Points)
}

func readPoints(scanner *bufio.Scanner) []api.Point {
	scanner.Scan()
	line := scanner.Text()
	parts := strings.Fields(line)
	var points []api.Point
	for _, p := range parts {
		pt := parsePoint(p)
		if pt.X != 0 || pt.Y != 0 || p == "0,0" {
			points = append(points, pt)
		}
	}
	return points
}

func parsePoint(s string) api.Point {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return api.Point{}
	}
	x, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	y, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	return api.Point{X: x, Y: y}
}

func parseWeightedPoint(s string) *api.WeightedPoint {
	parts := strings.Split(s, ",")
	if len(parts) != 3 {
		return nil
	}
	x, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	y, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	w, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	return &api.WeightedPoint{
		Point:  api.Point{X: x, Y: y},
		Weight: w,
	}
}

func printResponse(resp api.ComputeResponse) {
	fmt.Printf("Success: %v\n", resp.Success)
	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		return
	}
	fmt.Printf("Type: %s\n", resp.Hull.Type)
	fmt.Printf("Points (%d):\n", len(resp.Hull.Points))
	for i, p := range resp.Hull.Points {
		fmt.Printf("  %d: (%.4f, %.4f)\n", i, p.X, p.Y)
	}
}

func post(url string, req, resp interface{}) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	httpResp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	body, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("server error: %s - %s", httpResp.Status, string(body))
	}
	return json.Unmarshal(body, resp)
}
