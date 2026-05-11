package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"transit-system/common"
)

var baseURL string

func main() {
	flag.StringVar(&baseURL, "server", "http://localhost:8080", "Transit system server URL")
	flag.Parse()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "stations":
		handleStations()
	case "lines":
		handleLines()
	case "vehicles":
		handleVehicles()
	case "line-vehicles":
		handleLineVehicles()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Transit System Client")
	fmt.Println("Usage:")
	fmt.Println("  client stations              - List all stations")
	fmt.Println("  client lines                 - List all lines")
	fmt.Println("  client vehicles              - List all vehicles")
	fmt.Println("  client line-vehicles <code> [direction] - List vehicles on a line")
	fmt.Println("  client help                  - Show this help message")
	fmt.Println("\nOptions:")
	fmt.Println("  -server <url>                - Server URL (default: http://localhost:8080)")
}

func handleStations() {
	resp, err := http.Get(baseURL + "/api/stations")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", body)
		os.Exit(1)
	}

	var apiResp common.APIResponse
	json.Unmarshal(body, &apiResp)

	dataBytes, _ := json.Marshal(apiResp.Data)
	var stations []common.StationDTO
	json.Unmarshal(dataBytes, &stations)

	fmt.Println("\n=== Stations ===")
	for _, station := range stations {
		fmt.Printf("Code: %s, Name: %s, Type: %s\n", station.Code, station.Name, getStationTypeName(station.Type))
		fmt.Printf("  Coordinates: (%.4f, %.4f)\n", station.Latitude, station.Longitude)
	}
}

func handleLines() {
	resp, err := http.Get(baseURL + "/api/lines")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", body)
		os.Exit(1)
	}

	var apiResp common.APIResponse
	json.Unmarshal(body, &apiResp)

	dataBytes, _ := json.Marshal(apiResp.Data)
	var lines []common.LineDTO
	json.Unmarshal(dataBytes, &lines)

	fmt.Println("\n=== Lines ===")
	for _, line := range lines {
		fmt.Printf("\nCode: %s, Name: %s, Direction: %s\n", line.Code, line.Name, line.Direction)
		fmt.Printf("  First: %s, Last: %s\n", line.FirstTime, line.LastTime)
		fmt.Printf("  Operational: %v, HasError: %v\n", line.IsOperational, line.HasError)
		if line.HasError {
			fmt.Printf("  Error: %s\n", line.ErrorReason)
		}
		fmt.Printf("  Stations (%d):\n", len(line.Stations))
		for i, station := range line.Stations {
			fmt.Printf("    %d. %s (%s)\n", i+1, station.Name, station.Code)
		}
	}
}

func handleVehicles() {
	resp, err := http.Get(baseURL + "/api/vehicles")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", body)
		os.Exit(1)
	}

	var apiResp common.APIResponse
	json.Unmarshal(body, &apiResp)

	dataBytes, _ := json.Marshal(apiResp.Data)
	var vehicles []common.VehicleDTO
	json.Unmarshal(dataBytes, &vehicles)

	fmt.Println("\n=== Vehicles ===")
	for _, vehicle := range vehicles {
		fmt.Printf("\nID: %s\n", vehicle.ID)
		fmt.Printf("  Line: %s (%s)\n", vehicle.LineCode, vehicle.LineDirection)
		fmt.Printf("  Status: %s\n", vehicle.Status)
		fmt.Printf("  Current Station Index: %d\n", vehicle.CurrentStation)
		fmt.Printf("  Position: (%.4f, %.4f)\n", vehicle.Latitude, vehicle.Longitude)
	}
}

func handleLineVehicles() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: client line-vehicles <line-code> [direction]")
		os.Exit(1)
	}

	lineCode := os.Args[2]
	direction := "up"
	if len(os.Args) >= 4 {
		direction = os.Args[3]
	}

	url := fmt.Sprintf("%s/api/lines/%s/%s/vehicles", baseURL, lineCode, direction)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", body)
		os.Exit(1)
	}

	var apiResp common.APIResponse
	json.Unmarshal(body, &apiResp)

	dataBytes, _ := json.Marshal(apiResp.Data)
	var vehicles []common.VehicleDTO
	json.Unmarshal(dataBytes, &vehicles)

	fmt.Printf("\n=== Vehicles on Line %s (%s) ===\n", lineCode, direction)
	if len(vehicles) == 0 {
		fmt.Println("No vehicles found")
		return
	}

	for _, vehicle := range vehicles {
		fmt.Printf("\nID: %s\n", vehicle.ID)
		fmt.Printf("  Status: %s\n", vehicle.Status)
		fmt.Printf("  Current Station Index: %d\n", vehicle.CurrentStation)
		fmt.Printf("  Position: (%.4f, %.4f)\n", vehicle.Latitude, vehicle.Longitude)
	}
}

func getStationTypeName(typeStr string) string {
	switch typeStr {
	case "1":
		return "换乘站"
	case "2":
		return "首末站"
	default:
		return "普通站"
	}
}

func sendJSONRequest(method, url string, body interface{}) ([]byte, int, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return respBody, resp.StatusCode, nil
}
