package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go-escape-analyze/pkg/api"
)

const defaultThreshold = 20

func main() {
	var serverURL string
	var verbose bool
	var threshold int
	var detailed bool
	var filterFile string
	var filterReason string

	flag.StringVar(&serverURL, "server", "", "Escape analysis server URL (default: http://localhost:8500)")
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.IntVar(&threshold, "threshold", defaultThreshold, "Warning threshold for escape count")
	flag.BoolVar(&detailed, "detailed", false, "Use -gcflags=\"-m=2\" for more detailed output")
	flag.StringVar(&filterFile, "file", "", "Filter by file name")
	flag.StringVar(&filterReason, "reason", "", "Filter by escape reason")
	flag.Parse()

	if serverURL == "" {
		serverURL = os.Getenv("ESCAPE_ANALYZE_SERVER")
	}
	if serverURL == "" {
		serverURL = "http://localhost:8500"
	}

	sourceDir := "."
	if flag.NArg() > 0 {
		sourceDir = flag.Arg(0)
	}

	absDir, err := filepath.Abs(sourceDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving source directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Analyzing escape in: %s\n", absDir)

	gcflags := "-m"
	if detailed {
		gcflags = "-m=2"
	}

	output, err := getEscapeOutput(absDir, gcflags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting escape output: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Println("\n=== Raw compiler output ===")
		fmt.Println(output)
		fmt.Println("\n=== Analyze report ===")
	}

	var filter *api.FilterOptions
	if filterFile != "" || filterReason != "" {
		filter = &api.FilterOptions{
			FileName:     filterFile,
			EscapeReason: filterReason,
		}
	}

	report, err := analyzeOutput(serverURL, output, filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error analyzing output: %v\n", err)
		os.Exit(1)
	}

	printReport(report, threshold)
}

func getEscapeOutput(dir, gcflags string) (string, error) {
	args := []string{"build", "-gcflags", "all=" + gcflags, "./..."}

	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	outputBytes, err := io.ReadAll(stderr)
	if err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return string(outputBytes), nil
		}
		return "", err
	}

	return string(outputBytes), nil
}

func analyzeOutput(serverURL, output string, filter *api.FilterOptions) (*api.EscapeReport, error) {
	req := api.AnalyzeRequest{
		OutputText: output,
		Filter:     filter,
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/analyze", "application/json", bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var analyzeResp api.AnalyzeResponse
	if err := json.Unmarshal(body, &analyzeResp); err != nil {
		return nil, err
	}

	if !analyzeResp.Success {
		return nil, fmt.Errorf("server error: %s", analyzeResp.Error)
	}

	return analyzeResp.Report, nil
}

func printReport(report *api.EscapeReport, threshold int) {
	if report.TotalEscapes == 0 {
		fmt.Println("未检测到堆逃逸")
		return
	}

	fmt.Printf("\n总共检测到 %d 个逃逸变量\n", report.TotalEscapes)

	if report.TotalEscapes > threshold {
		fmt.Printf("\n⚠️  警告: 检测到 %d 个逃逸变量，超过阈值 %d\n", report.TotalEscapes, threshold)
	}

	fmt.Println(strings.Repeat("-", 60))

	for _, fileReport := range report.Files {
		fmt.Printf("\n📄 %s (%d 个逃逸)\n", fileReport.FilePath, fileReport.EscapeCount)
		fmt.Println(strings.Repeat("-", 40))

		for _, rec := range fileReport.EscapeRecords {
			var tags []string
			if rec.IsGenerated {
				tags = append(tags, "编译器生成")
			}
			if rec.IsReturnValue {
				tags = append(tags, "返回值")
			}
			if rec.IsClosureVar {
				tags = append(tags, "闭包捕获")
			}

			tagStr := ""
			if len(tags) > 0 {
				tagStr = fmt.Sprintf(" [%s]", strings.Join(tags, ", "))
			}

			fmt.Printf("  行 %d: %s  →  原因: %s%s\n",
				rec.LineNumber,
				rec.VariableName,
				rec.EscapeReason,
				tagStr)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("提示: 使用 -file 按文件过滤，-reason 按原因过滤\n")
	fmt.Printf("      使用 -detailed 获取更详细的输出\n")
	fmt.Println(strings.Repeat("=", 60))
}
