package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"buildtag/api"
)

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	return &Client{serverURL: strings.TrimSuffix(serverURL, "/")}
}

func (c *Client) Analyze(request api.AnalyzeRequest) (*api.AnalyzeResponse, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	resp, err := http.Post(c.serverURL+"/analyze", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}
	
	var response api.AnalyzeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &response, nil
}

func readGoFiles(dir string) ([]api.File, error) {
	var files []api.File
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}
		
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			relPath = path
		}
		
		files = append(files, api.File{
			Name:    relPath,
			Content: string(content),
		})
		
		return nil
	})
	
	return files, err
}

func printResponse(response *api.AnalyzeResponse) {
	fmt.Println("=== 文件包含/排除分析 ===")
	fmt.Println()
	
	fmt.Println("已包含的文件:")
	for _, f := range response.IncludedFiles {
		fmt.Printf("  [✓] %s\n", f.FileName)
		fmt.Printf("      原因: %s\n", f.Reason)
	}
	fmt.Println()
	
	fmt.Println("已排除的文件:")
	for _, f := range response.ExcludedFiles {
		fmt.Printf("  [✗] %s\n", f.FileName)
		fmt.Printf("      原因: %s\n", f.Reason)
	}
	fmt.Println()
	
	if len(response.Warnings) > 0 {
		fmt.Println("警告:")
		for _, w := range response.Warnings {
			fmt.Printf("  [!] %s\n", w)
		}
		fmt.Println()
	}
	
	fmt.Println("=== Generate 指令 ===")
	if len(response.GenerateDirectives) == 0 {
		fmt.Println("未找到任何 go:generate 指令")
	} else {
		for _, dir := range response.GenerateDirectives {
			fmt.Printf("  [%s:%d] %s\n", dir.FileName, dir.Position, dir.Command)
		}
	}
}

func main() {
	var osFlag, archFlag, serverURL string
	var tagsFlag string
	var dir string
	
	flag.StringVar(&osFlag, "os", "", "目标操作系统 (GOOS)")
	flag.StringVar(&archFlag, "arch", "", "目标架构 (GOARCH)")
	flag.StringVar(&tagsFlag, "tags", "", "自定义标签，用逗号分隔")
	flag.StringVar(&serverURL, "server", "http://localhost:8400", "服务端 URL")
	
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] <directory>\n", filepath.Base(os.Args[0]))
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
	}
	
	flag.Parse()
	
	args := flag.Args()
	if len(args) == 0 {
		dir = "."
	} else {
		dir = args[0]
	}
	
	files, err := readGoFiles(dir)
	if err != nil {
		log.Fatalf("Failed to read Go files: %v", err)
	}
	
	if len(files) == 0 {
		log.Fatalf("No Go files found in directory: %s", dir)
	}
	
	var tags []string
	if tagsFlag != "" {
		for _, t := range strings.Split(tagsFlag, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}
	
	request := api.AnalyzeRequest{
		Files: files,
		Target: api.TargetPlatform{
			GOOS:   osFlag,
			GOARCH: archFlag,
			Tags:   tags,
		},
	}
	
	client := NewClient(serverURL)
	response, err := client.Analyze(request)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}
	
	printResponse(response)
}
