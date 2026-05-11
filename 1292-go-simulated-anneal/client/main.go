package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tsp-simulated-anneal/common"
)

var (
	serverURL string
	inputFile string
	taskID    string
	download  bool
	outputFile string
	interactive bool
)

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:8101", "Server URL")
	flag.StringVar(&inputFile, "file", "", "Input CSV file with cities")
	flag.StringVar(&taskID, "task", "", "Task ID to query progress or result")
	flag.BoolVar(&download, "download", false, "Download result CSV")
	flag.StringVar(&outputFile, "output", "", "Output file for downloaded result")
	flag.BoolVar(&interactive, "interactive", false, "Interactive mode to add cities")
	flag.Parse()

	if taskID != "" {
		if download {
			downloadResult(taskID)
		} else {
			queryTask(taskID)
		}
		return
	}

	if interactive {
		runInteractiveMode()
		return
	}

	if inputFile != "" {
		runFileMode(inputFile)
		return
	}

	fmt.Println("Usage:")
	fmt.Println("  client -interactive                     : Interactive mode to add cities")
	fmt.Println("  client -file cities.csv                 : Submit cities from CSV file")
	fmt.Println("  client -task <taskId>                   : Query task progress/result")
	fmt.Println("  client -task <taskId> -download -output result.csv : Download result CSV")
	fmt.Println("")
	fmt.Println("Server URL can be specified with -server flag")
	os.Exit(1)
}

func runInteractiveMode() {
	reader := bufio.NewReader(os.Stdin)
	var cities []common.City

	fmt.Println("=== TSP Solver - Interactive Mode ===")
	fmt.Println("Enter cities one by one. Type 'done' when finished.")
	fmt.Println("")

	for {
		fmt.Printf("City %d:\n", len(cities)+1)
		fmt.Print("  Name: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)

		if strings.ToLower(name) == "done" {
			break
		}

		if name == "" {
			fmt.Println("  Error: Name cannot be empty")
			continue
		}

		fmt.Print("  Coordinate type (1=Cartesian, 2=Geographic): ")
		coordTypeStr, _ := reader.ReadString('\n')
		coordTypeStr = strings.TrimSpace(coordTypeStr)

		var coordType common.CoordinateType
		var city common.City
		city.Name = name

		if coordTypeStr == "1" {
			coordType = common.CoordinateTypeCartesian
			fmt.Print("  X: ")
			xStr, _ := reader.ReadString('\n')
			x, err := strconv.ParseFloat(strings.TrimSpace(xStr), 64)
			if err != nil {
				fmt.Println("  Error: Invalid X value")
				continue
			}
			fmt.Print("  Y: ")
			yStr, _ := reader.ReadString('\n')
			y, err := strconv.ParseFloat(strings.TrimSpace(yStr), 64)
			if err != nil {
				fmt.Println("  Error: Invalid Y value")
				continue
			}
			city.X = x
			city.Y = y
		} else if coordTypeStr == "2" {
			coordType = common.CoordinateTypeGeographic
			fmt.Print("  Latitude: ")
			latStr, _ := reader.ReadString('\n')
			lat, err := strconv.ParseFloat(strings.TrimSpace(latStr), 64)
			if err != nil {
				fmt.Println("  Error: Invalid Latitude value")
				continue
			}
			fmt.Print("  Longitude: ")
			lonStr, _ := reader.ReadString('\n')
			lon, err := strconv.ParseFloat(strings.TrimSpace(lonStr), 64)
			if err != nil {
				fmt.Println("  Error: Invalid Longitude value")
				continue
			}
			city.Latitude = lat
			city.Longitude = lon
		} else {
			fmt.Println("  Error: Invalid coordinate type")
			continue
		}
		city.CoordinateType = coordType
		cities = append(cities, city)
		fmt.Println("")
	}

	if len(cities) == 0 {
		fmt.Println("Error: No cities entered")
		os.Exit(1)
	}

	submitAndMonitor(cities)
}

func runFileMode(filename string) {
	cities, err := readCitiesFromFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	submitAndMonitor(cities)
}

func readCitiesFromFile(filename string) ([]common.City, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("file must have at least header and one data row")
	}

	headers := records[0]
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	_, hasName := headerMap["name"]
	_, hasLat := headerMap["latitude"]
	_, hasLon := headerMap["longitude"]
	_, hasX := headerMap["x"]
	_, hasY := headerMap["y"]

	if !hasName {
		return nil, fmt.Errorf("CSV must have 'name' column")
	}

	var cities []common.City
	for i := 1; i < len(records); i++ {
		row := records[i]
		city := common.City{
			Name: strings.TrimSpace(row[headerMap["name"]]),
		}

		if hasLat && hasLon {
			city.CoordinateType = common.CoordinateTypeGeographic
			if latIdx, ok := headerMap["latitude"]; ok {
				lat, _ := strconv.ParseFloat(strings.TrimSpace(row[latIdx]), 64)
				city.Latitude = lat
			}
			if lonIdx, ok := headerMap["longitude"]; ok {
				lon, _ := strconv.ParseFloat(strings.TrimSpace(row[lonIdx]), 64)
				city.Longitude = lon
			}
		} else if hasX && hasY {
			city.CoordinateType = common.CoordinateTypeCartesian
			if xIdx, ok := headerMap["x"]; ok {
				x, _ := strconv.ParseFloat(strings.TrimSpace(row[xIdx]), 64)
				city.X = x
			}
			if yIdx, ok := headerMap["y"]; ok {
				y, _ := strconv.ParseFloat(strings.TrimSpace(row[yIdx]), 64)
				city.Y = y
			}
		} else {
			return nil, fmt.Errorf("CSV must have either (latitude,longitude) or (x,y) columns")
		}

		cities = append(cities, city)
	}

	return cities, nil
}

func submitAndMonitor(cities []common.City) {
	fmt.Printf("Submitting %d cities to server...\n", len(cities))

	req := common.SolveRequest{
		Cities: cities,
	}

	reqBody, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/solve", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("Error submitting request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Server error: %s\n", string(body))
		os.Exit(1)
	}

	var solveResp common.SolveResponse
	json.NewDecoder(resp.Body).Decode(&solveResp)
	taskID = solveResp.TaskID

	fmt.Printf("Task ID: %s\n", taskID)
	fmt.Println("Monitoring progress...")
	fmt.Println("")

	monitorTask(taskID)
}

func queryTask(taskID string) {
	monitorTask(taskID)
}

func monitorTask(taskID string) {
	for {
		resp, err := http.Get(serverURL + "/progress?taskId=" + taskID)
		if err != nil {
			fmt.Printf("Error querying progress: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		var progress common.ProgressResponse
		json.NewDecoder(resp.Body).Decode(&progress)
		resp.Body.Close()

		if progress.Status == "completed" {
			fmt.Println("\n=== Task Completed ===")
			printResult(progress)
			return
		} else if progress.Status == "failed" {
			fmt.Printf("\nTask failed\n")
			return
		}

		fmt.Printf("\rStatus: %s | Temp: %.2f | Iter: %d | Best: %.4f | Improv: %.2f%%",
			progress.Status,
			progress.CurrentTemperature,
			progress.CurrentIteration,
			progress.BestDistance,
			progress.ImprovementPct)

		time.Sleep(500 * time.Millisecond)
	}
}

func printResult(result common.ProgressResponse) {
	fmt.Printf("\nTotal Distance: %.4f\n", result.TotalDistance)
	fmt.Printf("Initial Distance: %.4f\n", result.InitialDistance)
	fmt.Printf("Improvement: %.2f%%\n", result.ImprovementPct)
	fmt.Printf("Duration: %s\n", result.DurationStr)
	fmt.Printf("Total Iterations: %d\n", result.TotalIterations)
	fmt.Printf("Random Seed: %d\n", result.RandomSeed)
	fmt.Println("")
	fmt.Println("Path:")
	for i, idx := range result.Path {
		fmt.Printf("  %d. %s\n", i+1, result.CityNames[i])
		_ = idx
	}
	if len(result.CityNames) > 0 {
		fmt.Printf("  %d. %s (Return to start)\n", len(result.CityNames)+1, result.CityNames[0])
	}

	if len(result.Convergence) > 0 {
		fmt.Println("")
		fmt.Println("Convergence (first 5 and last 5 points):")
		showCount := 5
		if len(result.Convergence) <= showCount*2 {
			for i, cp := range result.Convergence {
				fmt.Printf("  [%d] Temp: %.4f, Value: %.4f\n", i+1, cp.Temperature, cp.ObjectiveValue)
			}
		} else {
			for i := 0; i < showCount; i++ {
				cp := result.Convergence[i]
				fmt.Printf("  [%d] Temp: %.4f, Value: %.4f\n", i+1, cp.Temperature, cp.ObjectiveValue)
			}
			fmt.Printf("  ... (%d more points) ...\n", len(result.Convergence)-showCount*2)
			for i := len(result.Convergence) - showCount; i < len(result.Convergence); i++ {
				cp := result.Convergence[i]
				fmt.Printf("  [%d] Temp: %.4f, Value: %.4f\n", i+1, cp.Temperature, cp.ObjectiveValue)
			}
		}
	}

	fmt.Println("")
	fmt.Printf("To download result CSV:\n")
	fmt.Printf("  client -server %s -task %s -download -output result.csv\n", serverURL, taskID)
}

func downloadResult(taskID string) {
	resp, err := http.Get(serverURL + "/download?taskId=" + taskID)
	if err != nil {
		fmt.Printf("Error downloading: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Server error: %s\n", string(body))
		os.Exit(1)
	}

	var writer io.Writer
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Printf("Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		writer = file
		fmt.Printf("Downloading to %s...\n", outputFile)
	} else {
		writer = os.Stdout
	}

	io.Copy(writer, resp.Body)
	if outputFile != "" {
		fmt.Println("Download completed.")
	}
}
