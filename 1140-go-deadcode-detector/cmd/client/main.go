package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"deadcode-detector/internal/common"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[37m"
)

var useColors = true

func main() {
	serverURL := flag.String("server", "http://localhost:8440", "Server URL")
	dir := flag.String("dir", ".", "Directory to scan")
	excludes := flag.String("exclude", "", "Comma-separated exclude patterns")
	filter := flag.String("filter", "", "Filter by type: function, type, variable, constant, field")
	references := flag.String("references", "", "Show references for identifier (requires report_id:name)")
	help := flag.Bool("help", false, "Show help")
	noColor := flag.Bool("no-color", false, "Disable colored output")

	flag.Parse()

	if *help {
		printHelp()
		return
	}

	if *noColor {
		useColors = false
	}

	if *references != "" {
		showReferences(*serverURL, *references)
		return
	}

	excludePatterns := []string{}
	if *excludes != "" {
		excludePatterns = strings.Split(*excludes, ",")
		for i, p := range excludePatterns {
			excludePatterns[i] = strings.TrimSpace(p)
		}
	}

	filterType := common.DeclarationType(*filter)
	if *filter != "" {
		switch *filter {
		case "function", "type", "variable", "constant", "field":
			filterType = common.DeclarationType(*filter)
		default:
			fmt.Printf("Invalid filter type: %s\n", *filter)
			fmt.Println("Valid types: function, type, variable, constant, field")
			os.Exit(1)
		}
	}

	runScan(*serverURL, *dir, excludePatterns, filterType)
}

func printHelp() {
	fmt.Println("Dead Code Detector Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dcd-client [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -server       Server URL (default: http://localhost:8440)")
	fmt.Println("  -dir          Directory to scan (default: .)")
	fmt.Println("  -exclude      Comma-separated exclude patterns")
	fmt.Println("  -filter       Filter by type: function, type, variable, constant, field")
	fmt.Println("  -references   Show references for identifier (format: report_id:name)")
	fmt.Println("  -no-color     Disable colored output")
	fmt.Println("  -help         Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  dcd-client -dir ./src")
	fmt.Println("  dcd-client -dir ./src -exclude deprecated_,test_")
	fmt.Println("  dcd-client -dir ./src -filter function")
	fmt.Println("  dcd-client -references report_1:myFunction")
}

func runScan(serverURL, dir string, excludes []string, filterType common.DeclarationType) {
	fmt.Printf("Scanning directory: %s\n", dir)

	files, err := collectGoFiles(dir)
	if err != nil {
		fmt.Printf("Error collecting files: %s\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No Go files found")
		return
	}

	fmt.Printf("Found %d Go file(s)\n", len(files))
	fmt.Println("Uploading to server...")

	req := common.AnalyzeRequest{
		Files:       files,
		PackageName: "",
	}

	resp, err := sendAnalyzeRequest(serverURL, req)
	if err != nil {
		fmt.Printf("Error analyzing: %s\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Server error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Printf("Report ID: %s\n", resp.ReportID)

	report := resp.Report

	if len(excludes) > 0 {
		report = applyExclusions(report, excludes)
	}

	if filterType != "" {
		report = applyFilter(report, filterType)
	}

	printReport(report, resp.ReportID)
}

func showReferences(serverURL, refParam string) {
	parts := strings.SplitN(refParam, ":", 2)
	if len(parts) != 2 {
		fmt.Println("Invalid references format. Use: report_id:name")
		os.Exit(1)
	}

	reportID := parts[0]
	name := parts[1]

	url := fmt.Sprintf("%s/references?report_id=%s&name=%s", serverURL, reportID, name)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error getting references: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %s\n", err)
		os.Exit(1)
	}

	var result common.ReferencePathResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error parsing response: %s\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("Server error: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("References for %s:\n", result.Name)
	fmt.Printf("Declaration: %s:%d\n", result.Declaration.File, result.Declaration.Line)

	if len(result.References) == 0 {
		fmt.Println("No references found")
		return
	}

	fmt.Printf("\nFound %d reference(s):\n", len(result.References))
	for i, ref := range result.References {
		fmt.Printf("%d. %s:%d\n", i+1, ref.File, ref.Line)
		fmt.Printf("   %s\n", ref.LineText)
	}
}

func collectGoFiles(dir string) ([]common.UploadFile, error) {
	var files []common.UploadFile

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			files = append(files, common.UploadFile{
				Path:    path,
				Content: string(content),
			})
		}

		return nil
	})

	return files, err
}

func sendAnalyzeRequest(serverURL string, req common.AnalyzeRequest) (*common.AnalyzeResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s/analyze", serverURL)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var result common.AnalyzeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &result, nil
}

func applyExclusions(report common.AnalysisReport, patterns []string) common.AnalysisReport {
	filtered := common.AnalysisReport{
		PackageName: report.PackageName,
		Entries:     make([]common.DeadCodeEntry, 0),
	}

	for _, entry := range report.Entries {
		excluded := false
		for _, pattern := range patterns {
			if strings.Contains(entry.Declaration.Name, pattern) {
				excluded = true
				break
			}
			if strings.Contains(entry.Declaration.Location.File, pattern) {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered.Entries = append(filtered.Entries, entry)
		}
	}

	return filtered
}

func applyFilter(report common.AnalysisReport, filterType common.DeclarationType) common.AnalysisReport {
	filtered := common.AnalysisReport{
		PackageName: report.PackageName,
		Entries:     make([]common.DeadCodeEntry, 0),
	}

	for _, entry := range report.Entries {
		if entry.Declaration.Type == filterType {
			filtered.Entries = append(filtered.Entries, entry)
		}
	}

	return filtered
}

func printReport(report common.AnalysisReport, reportID string) {
	fmt.Println()
	fmt.Println(colorize(ColorCyan, "=" + strings.Repeat("=", 58)))
	fmt.Println(colorize(ColorCyan, "  Dead Code Detection Report"))
	fmt.Println(colorize(ColorCyan, "=" + strings.Repeat("=", 58)))
	fmt.Println()
	fmt.Printf("Package: %s\n", report.PackageName)
	fmt.Printf("Report ID: %s\n", reportID)
	fmt.Println()

	if len(report.Entries) == 0 {
		fmt.Println(colorize(ColorGreen, "✓ No dead code found!"))
		return
	}

	printSummary(report.Summary)
	fmt.Println()

	entriesByFile := groupByFile(report.Entries)
	files := sortedKeys(entriesByFile)

	for _, file := range files {
		entries := entriesByFile[file]
		fmt.Println(colorize(ColorBlue, "📄 "+file))
		fmt.Println(colorize(ColorBlue, strings.Repeat("-", len(file)+4)))

		for _, entry := range entries {
			printEntry(entry)
		}
		fmt.Println()
	}
}

func printSummary(summary common.Summary) {
	fmt.Println(colorize(ColorYellow, "📊 Summary:"))
	fmt.Printf("  Functions: %d\n", summary.UnusedFunctions)
	fmt.Printf("  Types:     %d\n", summary.UnusedTypes)
	fmt.Printf("  Variables: %d\n", summary.UnusedVariables)
	fmt.Printf("  Constants: %d\n", summary.UnusedConstants)
	fmt.Printf("  Fields:    %d\n", summary.UnusedFields)

	total := summary.UnusedFunctions + summary.UnusedTypes + summary.UnusedVariables + summary.UnusedConstants + summary.UnusedFields
	fmt.Printf(colorize(ColorYellow, "  Total:     %d\n"), total)
}

func printEntry(entry common.DeadCodeEntry) {
	decl := entry.Declaration

	typeColor := getTypeColor(decl.Type)
	typeStr := colorize(typeColor, fmt.Sprintf("[%s]", decl.Type.String()))

	confColor := ColorGreen
	if entry.Confidence == common.ConfidenceCertain {
		confColor = ColorRed
	}
	confStr := colorize(confColor, entry.Confidence.String())

	fmt.Printf("  %s %s %s\n", typeStr, decl.Name, confStr)
	fmt.Printf("    Location: %s:%d\n", decl.Location.File, decl.Location.Line)
	fmt.Printf("    Reason: %s\n", entry.Reason)

	if decl.Snippet != "" {
		snippet := decl.Snippet
		if len(snippet) > 80 {
			snippet = snippet[:77] + "..."
		}
		fmt.Printf(colorize(ColorGray, "    Snippet: %s\n"), snippet)
	}
}

func getTypeColor(t common.DeclarationType) string {
	switch t {
	case common.TypeFunction:
		return ColorPurple
	case common.TypeType:
		return ColorBlue
	case common.TypeVariable:
		return ColorYellow
	case common.TypeConstant:
		return ColorCyan
	case common.TypeField:
		return ColorGray
	default:
		return ColorReset
	}
}

func groupByFile(entries []common.DeadCodeEntry) map[string][]common.DeadCodeEntry {
	result := make(map[string][]common.DeadCodeEntry)
	for _, entry := range entries {
		file := entry.Declaration.Location.File
		result[file] = append(result[file], entry)
	}
	return result
}

func sortedKeys(m map[string][]common.DeadCodeEntry) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func colorize(color, text string) string {
	if !useColors {
		return text
	}
	return color + text + ColorReset
}
