package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	
	"github.com/example/benchcompare/pkg/api"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

func main() {
	baselineFile := flag.String("baseline", "", "Baseline benchmark result file")
	currentFile := flag.String("current", "", "Current benchmark result file")
	serverURL := flag.String("server", "http://localhost:8080", "Benchmark compare server URL")
	noColor := flag.Bool("no-color", false, "Disable colored output")
	flag.Parse()
	
	var baselineContent, currentContent []byte
	var err error
	
	if *baselineFile == "-" {
		baselineContent, err = io.ReadAll(os.Stdin)
	} else {
		baselineContent, err = os.ReadFile(*baselineFile)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading baseline: %v\n", err)
		os.Exit(1)
	}
	
	if *currentFile == "-" {
		currentContent, err = io.ReadAll(os.Stdin)
	} else {
		currentContent, err = os.ReadFile(*currentFile)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading current: %v\n", err)
		os.Exit(1)
	}
	
	req := api.CompareRequest{
		Baseline: string(baselineContent),
		Current:  string(currentContent),
	}
	
	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling request: %v\n", err)
		os.Exit(1)
	}
	
	resp, err := http.Post(*serverURL+"/compare", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calling server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Server error: %s\n", string(body))
		os.Exit(1)
	}
	
	var compareResp api.CompareResponse
	if err := json.NewDecoder(resp.Body).Decode(&compareResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		os.Exit(1)
	}
	
	printReport(compareResp, *noColor)
}

func printReport(resp api.CompareResponse, noColor bool) {
	items := resp.Items
	sort.Slice(items, func(i, j int) bool {
		return getChangeMagnitude(items[i]) > getChangeMagnitude(items[j])
	})
	
	fmt.Printf("%-50s %-15s %-20s %-20s\n", 
		"Benchmark", "Status", "Time Change", "Memory Change")
	fmt.Println("---------------------------------------------------------------------------------------------------")
	
	for _, item := range items {
		statusStr := string(item.Status)
		timeChange := "N/A"
		memChange := "N/A"
		color := colorReset
		
		if item.Status == api.StatusRegression {
			if !noColor {
				color = colorRed
			}
		} else if item.Status == api.StatusImprovement {
			if !noColor {
				color = colorGreen
			}
		} else if item.Status == api.StatusChanged {
			if !noColor {
				color = colorYellow
			}
		} else if item.Status == api.StatusAdded {
			if !noColor {
				color = colorBlue
			}
		} else if item.Status == api.StatusRemoved {
			if !noColor {
				color = colorBlue
			}
		}
		
		if item.Baseline != nil && item.Current != nil {
			sign := ""
			if item.NsChangePercent >= 0 {
				sign = "+"
			}
			timeChange = fmt.Sprintf("%s%.2f%%", sign, item.NsChangePercent)
			
			if item.Baseline.BytesPerOp > 0 && item.Current.BytesPerOp > 0 {
				sign = ""
				if item.MemChangePercent >= 0 {
					sign = "+"
				}
				memChange = fmt.Sprintf("%s%.2f%%", sign, item.MemChangePercent)
			}
		}
		
		if item.Baseline != nil && item.Baseline.Uncertain || 
		   item.Current != nil && item.Current.Uncertain {
			statusStr += " (uncertain)"
		}
		
		fmt.Printf("%s%-50s %-15s %-20s %-20s%s\n",
			color, item.Name, statusStr, timeChange, memChange, colorReset)
	}
}

func getChangeMagnitude(item api.ComparisonItem) float64 {
	if item.Status == api.StatusAdded || item.Status == api.StatusRemoved {
		return 1000
	}
	return math.Max(math.Abs(item.NsChangePercent), math.Abs(item.MemChangePercent))
}
