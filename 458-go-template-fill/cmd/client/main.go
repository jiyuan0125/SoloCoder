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
	"template-engine/api"
)

const (
	DefaultServerURL = "http://localhost:8080"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	
	command := os.Args[1]
	
	switch command {
	case "render":
		handleRender()
	case "render-file":
		handleRenderFile()
	case "health":
		handleHealth()
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleRender() {
	fs := flag.NewFlagSet("render", flag.ExitOnError)
	serverURL := fs.String("server", DefaultServerURL, "服务端URL")
	templateStr := fs.String("template", "", "模板字符串 (必需)")
	dataJSON := fs.String("data", "{}", "JSON格式的数据")
	
	fs.Parse(os.Args[2:])
	
	if *templateStr == "" {
		fmt.Fprintln(os.Stderr, "错误: -template 参数是必需的")
		fs.Usage()
		os.Exit(1)
	}
	
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(*dataJSON), &data); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无效的JSON数据: %v\n", err)
		os.Exit(1)
	}
	
	req := api.RenderRequest{
		Template: *templateStr,
		Data:     data,
	}
	
	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 序列化请求失败: %v\n", err)
		os.Exit(1)
	}
	
	url := strings.TrimRight(*serverURL, "/") + "/render"
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 读取响应失败: %v\n", err)
		os.Exit(1)
	}
	
	var result api.RenderResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 解析响应失败: %v\n", err)
		fmt.Fprintf(os.Stderr, "原始响应: %s\n", string(body))
		os.Exit(1)
	}
	
	if !result.Success {
		fmt.Fprintf(os.Stderr, "渲染失败: %s\n", result.Error)
		os.Exit(1)
	}
	
	fmt.Println(result.Content)
	
	if len(result.Warnings) > 0 {
		fmt.Fprintln(os.Stderr, "\n警告:")
		for _, w := range result.Warnings {
			fmt.Fprintf(os.Stderr, "  - %s\n", w)
		}
	}
}

func handleRenderFile() {
	fs := flag.NewFlagSet("render-file", flag.ExitOnError)
	serverURL := fs.String("server", DefaultServerURL, "服务端URL")
	templatePath := fs.String("path", "", "模板文件路径 (必需)")
	dataJSON := fs.String("data", "{}", "JSON格式的数据")
	
	fs.Parse(os.Args[2:])
	
	if *templatePath == "" {
		fmt.Fprintln(os.Stderr, "错误: -path 参数是必需的")
		fs.Usage()
		os.Exit(1)
	}
	
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(*dataJSON), &data); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无效的JSON数据: %v\n", err)
		os.Exit(1)
	}
	
	req := api.RenderFileRequest{
		TemplatePath: *templatePath,
		Data:         data,
	}
	
	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 序列化请求失败: %v\n", err)
		os.Exit(1)
	}
	
	url := strings.TrimRight(*serverURL, "/") + "/render-file"
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 读取响应失败: %v\n", err)
		os.Exit(1)
	}
	
	var result api.RenderFileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 解析响应失败: %v\n", err)
		fmt.Fprintf(os.Stderr, "原始响应: %s\n", string(body))
		os.Exit(1)
	}
	
	if !result.Success {
		fmt.Fprintf(os.Stderr, "渲染失败: %s\n", result.Error)
		os.Exit(1)
	}
	
	fmt.Println(result.Content)
	
	if len(result.Warnings) > 0 {
		fmt.Fprintln(os.Stderr, "\n警告:")
		for _, w := range result.Warnings {
			fmt.Fprintf(os.Stderr, "  - %s\n", w)
		}
	}
}

func handleHealth() {
	fs := flag.NewFlagSet("health", flag.ExitOnError)
	serverURL := fs.String("server", DefaultServerURL, "服务端URL")
	
	fs.Parse(os.Args[2:])
	
	url := strings.TrimRight(*serverURL, "/") + "/health"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 健康检查失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 读取响应失败: %v\n", err)
		os.Exit(1)
	}
	
	var result api.HealthResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 解析响应失败: %v\n", err)
		fmt.Fprintf(os.Stderr, "原始响应: %s\n", string(body))
		os.Exit(1)
	}
	
	fmt.Printf("状态: %s\n", result.Status)
	fmt.Printf("版本: %s\n", result.Version)
}

func printUsage() {
	fmt.Println(`模板引擎客户端

用法:
  template-client <command> [options]

命令:
  render       渲染字符串模板
  render-file  从文件渲染模板
  health       检查服务端健康状态
  help         显示帮助信息

render 命令选项:
  -template string   模板字符串 (必需)
  -data string       JSON格式的数据 (默认: "{}")
  -server string     服务端URL (默认: "http://localhost:8080")

render-file 命令选项:
  -path string       模板文件路径 (必需)
  -data string       JSON格式的数据 (默认: "{}")
  -server string     服务端URL (默认: "http://localhost:8080")

health 命令选项:
  -server string     服务端URL (默认: "http://localhost:8080")

示例:
  # 渲染字符串模板
  template-client render -template "Hello, {{name}}!" -data '{"name": "World"}'
  
  # 从文件渲染模板
  template-client render-file -path ./template.txt -data '{"name": "World"}'
  
  # 检查服务端状态
  template-client health
`)
}
