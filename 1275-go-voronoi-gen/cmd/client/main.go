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

	"voronoi/internal/types"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return &Client{baseURL: baseURL}
}

func (c *Client) doRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	return client.Do(req)
}

func (c *Client) Generate(seeds []types.Point, bounds types.Bounds) (*types.GenerateResponse, error) {
	req := types.GenerateRequest{
		Seeds:  seeds,
		Bounds: bounds,
	}

	resp, err := c.doRequest("POST", "/generate", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(body))
	}

	var result types.GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) NearestNeighbor(point types.Point) (*types.NearestNeighborResponse, error) {
	req := types.NearestNeighborRequest{Point: point}

	resp, err := c.doRequest("POST", "/nearest", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(body))
	}

	var result types.NearestNeighborResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) Clip(region types.Bounds) (*types.ClipResponse, error) {
	req := types.ClipRequest{Region: region}

	resp, err := c.doRequest("POST", "/clip", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(body))
	}

	var result types.ClipResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) Area() (*types.AreaResponse, error) {
	req := types.AreaRequest{}

	resp, err := c.doRequest("POST", "/area", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(body))
	}

	var result types.AreaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func readSeedsFromFile(filename string) ([]types.Point, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var seeds []types.Point
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format on line %d: expected x,y", lineNum)
		}

		x, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid x coordinate on line %d: %w", lineNum, err)
		}

		y, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid y coordinate on line %d: %w", lineNum, err)
		}

		seeds = append(seeds, types.Point{X: x, Y: y})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return seeds, nil
}

func printHelp() {
	fmt.Println("Voronoi Diagram Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client [flags] <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  generate <seeds_file> [bounds]  Generate Voronoi diagram from seeds file")
	fmt.Println("  nearest <x,y>                   Find nearest neighbor to a point")
	fmt.Println("  clip <minX,minY,maxX,maxY>      Clip edges to rectangular region")
	fmt.Println("  area                            Calculate cell areas")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -server string   Server URL (default \"http://localhost:8503\")")
	fmt.Println("  -output string   Output file (default: stdout)")
	fmt.Println("  -json            Output as JSON")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client generate seeds.txt")
	fmt.Println("  client generate seeds.txt 0,0,100,100")
	fmt.Println("  client nearest 50,50")
	fmt.Println("  client clip 20,20,80,80")
	fmt.Println("  client area")
}

func main() {
	serverURL := flag.String("server", "http://localhost:8503", "Server URL")
	outputFile := flag.String("output", "", "Output file")
	jsonOutput := flag.Bool("json", false, "Output as JSON")
	flag.Parse()

	if flag.NArg() < 1 {
		printHelp()
		os.Exit(1)
	}

	command := flag.Arg(0)
	client := NewClient(*serverURL)

	var output io.Writer = os.Stdout
	if *outputFile != "" {
		file, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		output = file
	}

	switch command {
	case "generate":
		if flag.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "Error: seeds file required for generate command")
			printHelp()
			os.Exit(1)
		}

		seedsFile := flag.Arg(1)
		seeds, err := readSeedsFromFile(seedsFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading seeds file: %v\n", err)
			os.Exit(1)
		}

		var bounds types.Bounds
		if flag.NArg() >= 3 {
			boundsStr := flag.Arg(2)
			parts := strings.Split(boundsStr, ",")
			if len(parts) == 4 {
				minX, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
				minY, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				maxX, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
				maxY, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
				bounds = types.Bounds{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
			}
		}

		resp, err := client.Generate(seeds, bounds)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating diagram: %v\n", err)
			os.Exit(1)
		}

		if *jsonOutput {
			json.NewEncoder(output).Encode(resp)
		} else {
			printGenerateResponse(output, resp)
		}

	case "nearest":
		if flag.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "Error: point required for nearest command (format: x,y)")
			printHelp()
			os.Exit(1)
		}

		pointStr := flag.Arg(1)
		parts := strings.Split(pointStr, ",")
		if len(parts) != 2 {
			fmt.Fprintln(os.Stderr, "Error: invalid point format, expected x,y")
			os.Exit(1)
		}

		x, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		y, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)

		resp, err := client.NearestNeighbor(types.Point{X: x, Y: y})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding nearest neighbor: %v\n", err)
			os.Exit(1)
		}

		if *jsonOutput {
			json.NewEncoder(output).Encode(resp)
		} else {
			fmt.Fprintf(output, "Nearest seed: (%.2f, %.2f)\n", resp.Seed.X, resp.Seed.Y)
			fmt.Fprintf(output, "Distance: %.4f\n", resp.Distance)
			fmt.Fprintf(output, "Cell index: %d\n", resp.CellIndex)
		}

	case "clip":
		if flag.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "Error: region required for clip command (format: minX,minY,maxX,maxY)")
			printHelp()
			os.Exit(1)
		}

		regionStr := flag.Arg(1)
		parts := strings.Split(regionStr, ",")
		if len(parts) != 4 {
			fmt.Fprintln(os.Stderr, "Error: invalid region format, expected minX,minY,maxX,maxY")
			os.Exit(1)
		}

		minX, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		minY, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		maxX, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
		maxY, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)

		resp, err := client.Clip(types.Bounds{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error clipping edges: %v\n", err)
			os.Exit(1)
		}

		if *jsonOutput {
			json.NewEncoder(output).Encode(resp)
		} else {
			fmt.Fprintf(output, "Clipped edges: %d\n", len(resp.Edges))
			for i, edge := range resp.Edges {
				fmt.Fprintf(output, "  Edge %d: (%.2f,%.2f) -> (%.2f,%.2f)\n",
					i, edge.Start.X, edge.Start.Y, edge.End.X, edge.End.Y)
			}
		}

	case "area":
		resp, err := client.Area()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error calculating areas: %v\n", err)
			os.Exit(1)
		}

		if *jsonOutput {
			json.NewEncoder(output).Encode(resp)
		} else {
			fmt.Fprintf(output, "Cell areas:\n")
			for i, area := range resp.Areas {
				fmt.Fprintf(output, "  Cell %d: %.4f\n", i, area)
			}
		}

	case "help", "--help", "-h":
		printHelp()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printGenerateResponse(output io.Writer, resp *types.GenerateResponse) {
	fmt.Fprintf(output, "Voronoi Diagram\n")
	fmt.Fprintf(output, "===============\n\n")
	fmt.Fprintf(output, "Number of seeds: %d\n", len(resp.Diagram.Seeds))
	fmt.Fprintf(output, "Number of edges: %d\n", len(resp.Diagram.Edges))
	fmt.Fprintf(output, "Bounds: [%.2f, %.2f] x [%.2f, %.2f]\n\n",
		resp.Diagram.Bounds.MinX, resp.Diagram.Bounds.MinY,
		resp.Diagram.Bounds.MaxX, resp.Diagram.Bounds.MaxY)

	if len(resp.Diagram.Cells) > 0 {
		fmt.Fprintf(output, "Cells:\n")
		for i, cell := range resp.Diagram.Cells {
			fmt.Fprintf(output, "\nCell %d:\n", i)
			fmt.Fprintf(output, "  Seed: (%.2f, %.2f)\n", cell.Seed.X, cell.Seed.Y)
			fmt.Fprintf(output, "  Vertices: %d\n", len(cell.Points))
			fmt.Fprintf(output, "  Edges: %d\n", len(cell.Edges))
			fmt.Fprintf(output, "  Area: %.4f\n", cell.Area)

			if len(cell.Points) > 0 {
				fmt.Fprintf(output, "  Polygon:\n")
				for j, p := range cell.Points {
					fmt.Fprintf(output, "    %d: (%.2f, %.2f)\n", j, p.X, p.Y)
				}
			}
		}
	}
}
