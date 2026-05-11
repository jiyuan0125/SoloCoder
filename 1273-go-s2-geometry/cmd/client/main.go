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

	"s2geometry/api"
)

var serverURL string

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:8500", "S2 server URL")
	flag.Parse()

	if flag.NArg() < 1 {
		printUsage()
		os.Exit(1)
	}

	command := flag.Arg(0)
	args := flag.Args()[1:]

	var err error
	switch command {
	case "latlng-to-cellid":
		err = cmdLatLngToCellID(args)
	case "cellid-to-latlng":
		err = cmdCellIDToLatLng(args)
	case "contains":
		err = cmdContains(args)
	case "neighbors":
		err = cmdNeighbors(args)
	case "covering":
		err = cmdCovering(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`S2 Geometry Client

Usage:
  s2client [--server=<url>] <command> [arguments]

Commands:
  latlng-to-cellid <lat> <lng> [level]     Convert latitude/longitude to Cell ID
  cellid-to-latlng <cellid>               Convert Cell ID to latitude/longitude bounds
  contains <container> <contained>       Check if one Cell contains another
  neighbors <cellid> [level]              Get neighboring Cell IDs
  covering <lat_lo> <lat_hi> <lng_lo> <lng_hi> [min_level] [max_level] [max_cells]
                                        Get Cell IDs covering a rectangle

Examples:
  s2client latlng-to-cellid 37.7749 -122.4194 10
  s2client cellid-to-latlng 1234567890123456789
  s2client contains 1234567890123456789 9876543210987654321
  s2client neighbors 1234567890123456789 10
  s2client covering 37 38 -123 -122 5 10 100`)
}

func cmdLatLngToCellID(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: latlng-to-cellid <lat> <lng> [level]")
	}

	lat, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid latitude: %v", err)
	}
	lng, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid longitude: %v", err)
	}
	level := 30
	if len(args) >= 3 {
		level, err = strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid level: %v", err)
		}
	}

	req := api.LatLngToCellIDRequest{
		Lat:   lat,
		Lng:   lng,
		Level: level,
	}

	resp, err := postJSON[api.LatLngToCellIDResponse]("/latlng-to-cellid", req)
	if err != nil {
		return err
	}

	fmt.Printf("Cell ID: %s\n", resp.CellID)
	fmt.Printf("Face: %d\n", resp.Face)
	fmt.Printf("Level: %d\n", resp.Level)
	return nil
}

func cmdCellIDToLatLng(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cellid-to-latlng <cellid>")
	}

	req := api.CellIDToLatLngRequest{
		CellID: args[0],
	}

	resp, err := postJSON[api.CellIDToLatLngResponse]("/cellid-to-latlng", req)
	if err != nil {
		return err
	}

	fmt.Printf("Center Lat: %.6f\n", resp.Lat)
	fmt.Printf("Center Lng: %.6f\n", resp.Lng)
	fmt.Printf("Lat Range: %.6f - %.6f\n", resp.LatLo, resp.LatHi)
	fmt.Printf("Lng Range: %.6f - %.6f\n", resp.LngLo, resp.LngHi)
	fmt.Printf("Face: %d\n", resp.Face)
	fmt.Printf("Level: %d\n", resp.Level)
	return nil
}

func cmdContains(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: contains <container> <contained>")
	}

	req := api.ContainsRequest{
		Container: args[0],
		Contained: args[1],
	}

	resp, err := postJSON[api.ContainsResponse]("/contains", req)
	if err != nil {
		return err
	}

	if resp.Contains {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}
	return nil
}

func cmdNeighbors(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: neighbors <cellid> [level]")
	}

	level := 30
	var err error
	if len(args) >= 2 {
		level, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid level: %v", err)
		}
	}

	req := api.NeighborsRequest{
		CellID: args[0],
		Level:  level,
	}

	resp, err := postJSON[api.NeighborsResponse]("/neighbors", req)
	if err != nil {
		return err
	}

	fmt.Println("Neighbors:")
	for _, n := range resp.Neighbors {
		fmt.Printf("  %s\n", n)
	}
	return nil
}

func cmdCovering(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: covering <lat_lo> <lat_hi> <lng_lo> <lng_hi> [min_level] [max_level] [max_cells]")
	}

	latLo, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid lat_lo: %v", err)
	}
	latHi, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid lat_hi: %v", err)
	}
	lngLo, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Errorf("invalid lng_lo: %v", err)
	}
	lngHi, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		return fmt.Errorf("invalid lng_hi: %v", err)
	}

	minLevel := 0
	maxLevel := 30
	maxCells := 100

	if len(args) >= 5 {
		minLevel, err = strconv.Atoi(args[4])
		if err != nil {
			return fmt.Errorf("invalid min_level: %v", err)
		}
	}
	if len(args) >= 6 {
		maxLevel, err = strconv.Atoi(args[5])
		if err != nil {
			return fmt.Errorf("invalid max_level: %v", err)
		}
	}
	if len(args) >= 7 {
		maxCells, err = strconv.Atoi(args[6])
		if err != nil {
			return fmt.Errorf("invalid max_cells: %v", err)
		}
	}

	req := api.CoveringRequest{
		LatLo:    latLo,
		LatHi:    latHi,
		LngLo:    lngLo,
		LngHi:    lngHi,
		MinLevel: minLevel,
		MaxLevel: maxLevel,
		MaxCells: maxCells,
	}

	resp, err := postJSON[api.CoveringResponse]("/covering", req)
	if err != nil {
		return err
	}

	fmt.Printf("Covering Cells (%d):\n", len(resp.CellIDs))
	for _, id := range resp.CellIDs {
		fmt.Printf("  %s\n", id)
	}
	return nil
}

func postJSON[T any](path string, req interface{}) (*T, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := strings.TrimSuffix(serverURL, "/") + path
	httpResp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != "" {
			return nil, fmt.Errorf("server error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("HTTP %d: %s", httpResp.StatusCode, string(respBody))
	}

	var result T
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
