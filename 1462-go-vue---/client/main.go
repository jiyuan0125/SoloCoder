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
	"time"

	"energymanagement/common"
)

const defaultServerURL = "http://localhost:8080"

var serverURL string

func main() {
	flag.StringVar(&serverURL, "server", defaultServerURL, "server URL")
	flag.Parse()

	if flag.NArg() < 1 {
		printUsage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	switch cmd {
	case "list-points":
		listPoints()
	case "create-point":
		createPoint()
	case "report":
		reportData()
	case "point-analysis":
		pointAnalysis()
	case "area-analysis":
		areaAnalysis()
	case "area-compare":
		areaCompare()
	case "overview":
		overview()
	case "suggestions":
		suggestions()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage: em-client [options] <command> [args]

Options:
  -server string    Server URL (default: http://localhost:8080)

Commands:
  list-points                    List all monitoring points
  
  create-point <name> <area> <energy-type> <unit> [area-size]
    Create a new monitoring point
    energy-type: electricity, water, gas, steam
  
  report <point-id> <reading> [timestamp]
    Report energy data for a point
  
  point-analysis <point-id> <granularity> <start> <end>
    Get consumption analysis for a point
    granularity: hour, day, month
    start/end: RFC3339 format
  
  area-analysis <area> <granularity> <start> <end>
    Get consumption analysis for an area
  
  area-compare <start> <end>
    Compare consumption across areas
  
  overview [time]
    Get overview metrics
    time: RFC3339 format (default: now)
  
  suggestions
    List all energy saving suggestions
`)
}

func listPoints() {
	resp, err := http.Get(serverURL + "/api/points")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if !result.Success {
		fmt.Printf("Server error: %s\n", result.Error)
		return
	}

	pointsData, _ := json.MarshalIndent(result.Data, "", "  ")
	fmt.Println(string(pointsData))
}

func createPoint() {
	args := flag.Args()[1:]
	if len(args) < 4 {
		fmt.Println("Usage: create-point <name> <area> <energy-type> <unit> [area-size]")
		os.Exit(1)
	}

	req := common.CreatePointRequest{
		Name:       args[0],
		Area:       args[1],
		EnergyType: common.EnergyType(args[2]),
		Unit:       args[3],
	}

	if len(args) >= 5 {
		var areaSize float64
		fmt.Sscanf(args[4], "%f", &areaSize)
		req.AreaSize = areaSize
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/points", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if !result.Success {
		fmt.Printf("Server error: %s\n", result.Error)
		return
	}

	data, _ := json.MarshalIndent(result.Data, "", "  ")
	fmt.Println(string(data))
}

func reportData() {
	args := flag.Args()[1:]
	if len(args) < 2 {
		fmt.Println("Usage: report <point-id> <reading> [timestamp]")
		os.Exit(1)
	}

	var reading float64
	fmt.Sscanf(args[1], "%f", &reading)

	timestamp := time.Now()
	if len(args) >= 3 {
		if t, err := time.Parse(time.RFC3339, args[2]); err == nil {
			timestamp = t
		}
	}

	req := common.ReportDataRequest{
		PointID:   args[0],
		Timestamp: timestamp,
		Reading:   reading,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/data/report", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if !result.Success {
		fmt.Printf("Server error: %s\n", result.Error)
		return
	}

	data, _ := json.MarshalIndent(result.Data, "", "  ")
	fmt.Println(string(data))
}

func pointAnalysis() {
	args := flag.Args()[1:]
	if len(args) < 4 {
		fmt.Println("Usage: point-analysis <point-id> <granularity> <start> <end>")
		os.Exit(1)
	}

	start, _ := time.Parse(time.RFC3339, args[2])
	end, _ := time.Parse(time.RFC3339, args[3])

	req := common.PointTimeQuery{
		PointID:     args[0],
		Start:       start,
		End:         end,
		Granularity: common.TimeGranularity(args[1]),
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/analysis/point", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if !result.Success {
		fmt.Printf("Server error: %s\n", result.Error)
		return
	}

	data, _ := json.MarshalIndent(result.Data, "", "  ")
	fmt.Println(string(data))
}

func areaAnalysis() {
	args := flag.Args()[1:]
	if len(args) < 4 {
		fmt.Println("Usage: area-analysis <area> <granularity> <start> <end>")
		os.Exit(1)
	}

	start, _ := time.Parse(time.RFC3339, args[2])
	end, _ := time.Parse(time.RFC3339, args[3])

	req := common.AreaTimeQuery{
		Area:        args[0],
		Start:       start,
		End:         end,
		Granularity: common.TimeGranularity(args[1]),
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/analysis/area", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if !result.Success {
		fmt.Printf("Server error: %s\n", result.Error)
		return
	}

	data, _ := json.MarshalIndent(result.Data, "", "  ")
	fmt.Println(string(data))
}

func areaCompare() {
	args := flag.Args()[1:]
	if len(args) < 2 {
		fmt.Println("Usage: area-compare <start> <end>")
		os.Exit(1)
	}

	start, _ := time.Parse(time.RFC3339, args[0])
	end, _ := time.Parse(time.RFC3339, args[1])

	req := common.TimeRangeQuery{
		Start: start,
		End:   end,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/analysis/area-comparison", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if !result.Success {
		fmt.Printf("Server error: %s\n", result.Error)
		return
	}

	data, _ := json.MarshalIndent(result.Data, "", "  ")
	fmt.Println(string(data))
}

func overview() {
	args := flag.Args()[1:]
	refTime := time.Now()
	if len(args) >= 1 {
		if t, err := time.Parse(time.RFC3339, args[0]); err == nil {
			refTime = t
		}
	}

	url := serverURL + "/api/overview?time=" + strings.Replace(refTime.Format(time.RFC3339), "+", "%2B", -1)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if !result.Success {
		fmt.Printf("Server error: %s\n", result.Error)
		return
	}

	data, _ := json.MarshalIndent(result.Data, "", "  ")
	fmt.Println(string(data))
}

func suggestions() {
	resp, err := http.Get(serverURL + "/api/suggestions")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if !result.Success {
		fmt.Printf("Server error: %s\n", result.Error)
		return
	}

	data, _ := json.MarshalIndent(result.Data, "", "  ")
	fmt.Println(string(data))
}

func _checkError(err error) {
}

func _readAll(resp *http.Response) {
	io.ReadAll(resp.Body)
}
