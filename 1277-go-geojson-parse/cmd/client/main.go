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

	"github.com/spatial-index/pkg/common"
)

const defaultServerURL = "http://localhost:8504"

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	return &Client{baseURL: strings.TrimSuffix(baseURL, "/")}
}

func (c *Client) Insert(geojsonData []byte) (*common.InsertResponse, error) {
	var geojsonObj interface{}
	if err := json.Unmarshal(geojsonData, &geojsonObj); err != nil {
		return nil, fmt.Errorf("invalid GeoJSON: %w", err)
	}

	reqBody := map[string]interface{}{
		"geojson": geojsonObj,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.baseURL+"/insert", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result common.InsertResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("server response: %s", string(respBody))
	}

	return &result, nil
}

func (c *Client) RangeQuery(minLon, minLat, maxLon, maxLat float64) (*common.QueryResponse, error) {
	req := common.RangeQueryRequest{
		MinLon: minLon,
		MinLat: minLat,
		MaxLon: maxLon,
		MaxLat: maxLat,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.baseURL+"/query/range", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result common.QueryResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("server response: %s", string(respBody))
	}

	return &result, nil
}

func (c *Client) PointQuery(lon, lat float64) (*common.QueryResponse, error) {
	req := common.PointQueryRequest{
		Lon: lon,
		Lat: lat,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.baseURL+"/query/point", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result common.QueryResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("server response: %s", string(respBody))
	}

	return &result, nil
}

func (c *Client) NearestNeighbor(lon, lat float64, k int) (*common.QueryResponse, error) {
	req := common.NearestNeighborRequest{
		Lon: lon,
		Lat: lat,
		K:   k,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.baseURL+"/query/nearest", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result common.QueryResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("server response: %s", string(respBody))
	}

	return &result, nil
}

func (c *Client) Stats() (*common.StatsResponse, error) {
	resp, err := http.Get(c.baseURL + "/stats")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result common.StatsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("server response: %s", string(respBody))
	}

	return &result, nil
}

func (c *Client) Clear() (*common.ClearResponse, error) {
	resp, err := http.Post(c.baseURL+"/clear", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result common.ClearResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("server response: %s", string(respBody))
	}

	return &result, nil
}

func printUsage() {
	fmt.Println("Spatial Index Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client [flags] <command> [args]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -server string   Server URL (default: http://localhost:8504 or $SERVER_URL)")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  insert <file>              Insert GeoJSON from file")
	fmt.Println("  range <minLon> <minLat> <maxLon> <maxLat>  Range query")
	fmt.Println("  point <lon> <lat>          Point query (find polygons containing point)")
	fmt.Println("  nearest <lon> <lat> [k]    Nearest neighbor query (default k=1)")
	fmt.Println("  stats                      Get index statistics")
	fmt.Println("  clear                      Clear all data")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client insert data.geojson")
	fmt.Println("  client range -74 40 -73 41")
	fmt.Println("  client point -73.9857 40.7484")
	fmt.Println("  client nearest -73.9857 40.7484 5")
	fmt.Println("  client stats")
}

func main() {
	serverURL := flag.String("server", "", "Server URL")
	flag.Parse()

	if *serverURL == "" {
		if envURL := os.Getenv("SERVER_URL"); envURL != "" {
			*serverURL = envURL
		} else {
			*serverURL = defaultServerURL
		}
	}

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(*serverURL)

	cmd := args[0]
	switch cmd {
	case "insert":
		if len(args) < 2 {
			fmt.Println("Error: insert command requires a file path")
			printUsage()
			os.Exit(1)
		}
		filePath := args[1]
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
		resp, err := client.Insert(data)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Printf("Server error: %s\n", resp.Error)
			os.Exit(1)
		}
		fmt.Printf("Inserted %d features\n", resp.Count)

	case "range":
		if len(args) < 5 {
			fmt.Println("Error: range command requires minLon, minLat, maxLon, maxLat")
			printUsage()
			os.Exit(1)
		}
		minLon := parseFloat(args[1], "minLon")
		minLat := parseFloat(args[2], "minLat")
		maxLon := parseFloat(args[3], "maxLon")
		maxLat := parseFloat(args[4], "maxLat")
		resp, err := client.RangeQuery(minLon, minLat, maxLon, maxLat)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printQueryResponse(resp)

	case "point":
		if len(args) < 3 {
			fmt.Println("Error: point command requires lon, lat")
			printUsage()
			os.Exit(1)
		}
		lon := parseFloat(args[1], "lon")
		lat := parseFloat(args[2], "lat")
		resp, err := client.PointQuery(lon, lat)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printQueryResponse(resp)

	case "nearest":
		if len(args) < 3 {
			fmt.Println("Error: nearest command requires lon, lat")
			printUsage()
			os.Exit(1)
		}
		lon := parseFloat(args[1], "lon")
		lat := parseFloat(args[2], "lat")
		k := 1
		if len(args) >= 4 {
			k = parseInt(args[3], "k")
		}
		resp, err := client.NearestNeighbor(lon, lat, k)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printQueryResponse(resp)

	case "stats":
		resp, err := client.Stats()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Index contains %d features\n", resp.Count)

	case "clear":
		resp, err := client.Clear()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if resp.Success {
			fmt.Println("Index cleared")
		}

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func parseFloat(s, name string) float64 {
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		fmt.Printf("Error: invalid %s value: %s\n", name, s)
		os.Exit(1)
	}
	return f
}

func parseInt(s, name string) int {
	var i int
	if _, err := fmt.Sscanf(s, "%d", &i); err != nil {
		fmt.Printf("Error: invalid %s value: %s\n", name, s)
		os.Exit(1)
	}
	return i
}

func printQueryResponse(resp *common.QueryResponse) {
	if !resp.Success {
		fmt.Printf("Server error: %s\n", resp.Error)
		os.Exit(1)
	}
	fmt.Printf("Found %d results\n", resp.Count)
	if resp.Features != nil {
		output, err := json.MarshalIndent(resp.Features, "", "  ")
		if err != nil {
			fmt.Printf("Error formatting response: %v\n", err)
			return
		}
		fmt.Println(string(output))
	}
}
