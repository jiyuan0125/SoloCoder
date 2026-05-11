package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"pprof-analyzer/internal/apimodels"
	"pprof-analyzer/internal/profiled"
)

func main() {
	var (
		topN       int
		serverAddr string
		localMode  bool
		history    bool
	)

	flag.IntVar(&topN, "top", 10, "显示前N个函数")
	flag.StringVar(&serverAddr, "server", "http://localhost:8080", "服务端地址")
	flag.BoolVar(&localMode, "local", false, "使用本地模式分析（不调用服务端）")
	flag.BoolVar(&history, "history", false, "查看历史分析记录")
	flag.Parse()

	if history {
		if err := printHistory(serverAddr); err != nil {
			fmt.Fprintf(os.Stderr, "获取历史记录失败: %v\n", err)
			os.Exit(1)
		}
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("使用方法:")
		fmt.Println("  本地分析: pprof-cli --local <pprof文件路径> [--top N]")
		fmt.Println("  远程分析: pprof-cli <pprof文件路径> [--server <地址>] [--top N]")
		fmt.Println("  查看历史: pprof-cli --history [--server <地址>]")
		os.Exit(1)
	}

	filePath := args[0]

	var response *apimodels.AnalysisResponse
	var err error

	if localMode {
		response, err = analyzeLocal(filePath)
	} else {
		response, err = analyzeRemote(filePath, serverAddr)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "分析失败: %v\n", err)
		os.Exit(1)
	}

	printResult(response, topN)
}

func analyzeLocal(filePath string) (*apimodels.AnalysisResponse, error) {
	result, err := profiled.AnalyzeFile(filePath)
	if err != nil {
		return nil, err
	}

	functions := make([]apimodels.FunctionInfo, len(result.Functions))
	for i, f := range result.Functions {
		functions[i] = apimodels.FunctionInfo{
			FullName:    f.FullName,
			PackageName: f.PackageName,
			FuncName:    f.FuncName,
			Bytes:       f.Bytes,
			Objects:     f.Objects,
			Percentage:  f.Percentage,
		}
	}

	return &apimodels.AnalysisResponse{
		ProfileType:  apimodels.ProfileType(result.ProfileType),
		TotalBytes:   result.TotalBytes,
		TotalObjects: result.TotalObjects,
		Functions:    functions,
	}, nil
}

func analyzeRemote(filePath, serverAddr string) (*apimodels.AnalysisResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("创建表单失败: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("复制文件失败: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("关闭writer失败: %w", err)
	}

	url := strings.TrimRight(serverAddr, "/") + "/analyze"
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("服务端返回错误: %s - %s", resp.Status, string(body))
	}

	var response apimodels.AnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &response, nil
}

func printHistory(serverAddr string) error {
	url := strings.TrimRight(serverAddr, "/") + "/history"
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("服务端返回错误: %s - %s", resp.Status, string(body))
	}

	var response apimodels.HistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if len(response.Records) == 0 {
		fmt.Println("暂无历史记录")
		return nil
	}

	fmt.Printf("%-30s %-20s %-15s %-15s %s\n", "ID", "文件名", "类型", "总字节数", "分析时间")
	fmt.Println(strings.Repeat("-", 100))
	for _, record := range response.Records {
		fmt.Printf("%-30s %-20s %-15s %-15s %s\n",
			truncate(record.ID, 30),
			truncate(record.FileName, 20),
			record.ProfileType,
			formatBytes(record.TotalBytes),
			record.AnalyzedAt)
	}

	return nil
}

func printResult(response *apimodels.AnalysisResponse, topN int) {
	fmt.Printf("Profile类型: %s\n", response.ProfileType)
	fmt.Printf("总采样字节数: %s (%d bytes)\n", formatBytes(response.TotalBytes), response.TotalBytes)
	fmt.Printf("总采样对象数: %d\n", response.TotalObjects)
	fmt.Println()

	functions := response.Functions
	var other *apimodels.FunctionInfo

	if topN > 0 && topN < len(functions) {
		other = &apimodels.FunctionInfo{
			FullName: "其他",
			FuncName: "其他",
		}
		for _, f := range functions[topN:] {
			other.Bytes += f.Bytes
			other.Objects += f.Objects
		}
		if response.TotalBytes > 0 {
			other.Percentage = float64(other.Bytes) / float64(response.TotalBytes) * 100.0
		}
		functions = functions[:topN]
	}

	maxNameLen := 40
	for _, f := range functions {
		if len(f.FullName) > maxNameLen {
			maxNameLen = len(f.FullName)
		}
	}
	if maxNameLen > 80 {
		maxNameLen = 80
	}

	headerFormat := fmt.Sprintf("%%-%ds | %%15s | %%15s | %%10s | %%15s\n", maxNameLen)
	rowFormat := fmt.Sprintf("%%-%ds | %%15s | %%15d | %%9.2f%%%% | %%15s\n", maxNameLen)
	otherFormat := fmt.Sprintf("%%-%ds | %%15s | %%15d | %%9.2f%%%% | %%15s\n", maxNameLen)

	fmt.Printf(headerFormat, "函数名", "字节数", "对象数", "占比", "累计占比")
	fmt.Println(strings.Repeat("-", maxNameLen+15+15+10+15+12))

	var cumulative float64
	for _, f := range functions {
		cumulative += f.Percentage
		fmt.Printf(rowFormat,
			truncate(f.FullName, maxNameLen),
			formatBytes(f.Bytes),
			f.Objects,
			f.Percentage,
			fmt.Sprintf("%.2f%%", cumulative))
	}

	if other != nil {
		cumulative += other.Percentage
		fmt.Printf(otherFormat,
			"其他",
			formatBytes(other.Bytes),
			other.Objects,
			other.Percentage,
			fmt.Sprintf("%.2f%%", cumulative))
	}
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
