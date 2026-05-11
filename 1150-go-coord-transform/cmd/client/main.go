package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"coordconv/pkg/model"
)

func main() {
	var from, to, server string
	flag.StringVar(&from, "from", "", "Source coordinate system (wgs84, gcj02, bd09)")
	flag.StringVar(&to, "to", "", "Target coordinate system (wgs84, gcj02, bd09)")
	flag.StringVar(&server, "server", "http://localhost:8080", "Server URL")
	flag.Parse()

	if from == "" || to == "" {
		fmt.Println("Usage: coordconv --from <source> --to <target> <coordinates...>")
		fmt.Println("Example: coordconv --from wgs84 --to gcj02 116.397428 39.90923")
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Error: no coordinates provided")
		os.Exit(1)
	}

	points, err := parseCoordinates(args)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	req := model.ConvertRequest{
		From:   strings.ToLower(from),
		To:     strings.ToLower(to),
		Points: points,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/api/convert", server)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: failed to read response: %v\n", err)
		os.Exit(1)
	}

	var result model.ConvertResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Printf("Error: failed to parse response: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	for i, p := range result.Points {
		if len(result.Points) > 1 {
			fmt.Printf("%d: ", i+1)
		}
		fmt.Printf("%.6f,%.6f\n", p.Lng, p.Lat)
	}
}

func parseCoordinates(args []string) ([]model.PointRequest, error) {
	var points []model.PointRequest

	if len(args) == 1 && strings.Contains(args[0], ",") {
		parts := strings.Split(args[0], ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid coordinate format")
		}
		lng, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid longitude")
		}
		lat, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid latitude")
		}
		points = append(points, model.PointRequest{Lng: lng, Lat: lat})
		return points, nil
	}

	if len(args) >= 2 && !strings.Contains(args[0], ",") {
		for i := 0; i < len(args); i += 2 {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing latitude for coordinate pair")
			}
			lng, err := strconv.ParseFloat(args[i], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid longitude: %s", args[i])
			}
			lat, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid latitude: %s", args[i+1])
			}
			points = append(points, model.PointRequest{Lng: lng, Lat: lat})
		}
		return points, nil
	}

	return nil, fmt.Errorf("invalid coordinate format")
}
