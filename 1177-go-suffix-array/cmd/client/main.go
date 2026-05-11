package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"suffixarray/pkg/api"
)

const defaultServerURL = "http://localhost:8401"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := getServerURL()

	command := os.Args[1]

	switch command {
	case "build":
		handleBuild(serverURL, os.Args[2:])
	case "lrs":
		handleLRS(serverURL, os.Args[2:])
	case "count":
		handleCount(serverURL, os.Args[2:])
	case "demo":
		handleDemo(serverURL, os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func getServerURL() string {
	if envURL := os.Getenv("SUFFIXARRAY_SERVER_URL"); envURL != "" {
		return envURL
	}
	return defaultServerURL
}

func printUsage() {
	fmt.Println("后缀数组客户端工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  suffixarray-client <command> [options]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  build   构建后缀数组和LCP数组")
	fmt.Println("  lrs     查询最长重复子串")
	fmt.Println("  count   查询子串出现次数")
	fmt.Println("  demo    演示所有功能")
	fmt.Println("  help    显示帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  suffixarray-client build --text \"banana\"")
	fmt.Println("  suffixarray-client lrs --text \"abracadabra\"")
	fmt.Println("  suffixarray-client count --text \"banana\" --pattern \"ana\"")
	fmt.Println("  suffixarray-client demo")
	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  SUFFIXARRAY_SERVER_URL  服务端URL (默认: http://localhost:8401)")
}

func handleBuild(serverURL string, args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	text := fs.String("text", "", "要处理的文本")
	fs.Parse(args)

	if *text == "" {
		fmt.Println("错误: 必须提供 --text 参数")
		os.Exit(1)
	}

	reqBody, _ := json.Marshal(api.BuildRequest{Text: *text})
	resp, err := http.Post(serverURL+"/build", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result api.BuildResponse
	json.Unmarshal(body, &result)

	if !result.Success {
		fmt.Printf("错误: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("文本: %s\n", result.Text)
	fmt.Printf("后缀数组 (SA): %v\n", result.SA)
	fmt.Printf("Rank数组: %v\n", result.Rank)
	fmt.Printf("LCP数组: %v\n", result.LCP)
}

func handleLRS(serverURL string, args []string) {
	fs := flag.NewFlagSet("lrs", flag.ExitOnError)
	text := fs.String("text", "", "要处理的文本")
	fs.Parse(args)

	if *text == "" {
		fmt.Println("错误: 必须提供 --text 参数")
		os.Exit(1)
	}

	reqBody, _ := json.Marshal(api.LRSRequest{Text: *text})
	resp, err := http.Post(serverURL+"/lrs", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result api.LRSResponse
	json.Unmarshal(body, &result)

	if !result.Success {
		fmt.Printf("错误: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("文本: %s\n", result.Text)
	fmt.Printf("最长重复子串长度: %d\n", result.MaxLength)
	if len(result.Substrings) == 0 {
		fmt.Println("没有找到重复子串")
	} else {
		fmt.Println("最长重复子串:")
		for i, s := range result.Substrings {
			fmt.Printf("  %d. \"%s\" 出现位置: %v\n", i+1, s.Substring, s.Positions)
		}
	}
}

func handleCount(serverURL string, args []string) {
	fs := flag.NewFlagSet("count", flag.ExitOnError)
	text := fs.String("text", "", "要处理的文本")
	pattern := fs.String("pattern", "", "要查询的子串")
	fs.Parse(args)

	if *text == "" || *pattern == "" {
		fmt.Println("错误: 必须提供 --text 和 --pattern 参数")
		os.Exit(1)
	}

	reqBody, _ := json.Marshal(api.CountRequest{Text: *text, Pattern: *pattern})
	resp, err := http.Post(serverURL+"/count", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result api.CountResponse
	json.Unmarshal(body, &result)

	if !result.Success {
		fmt.Printf("错误: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("文本: %s\n", result.Text)
	fmt.Printf("模式串: \"%s\"\n", result.Pattern)
	fmt.Printf("出现次数: %d\n", result.Count)
	if result.Count > 0 {
		fmt.Printf("出现位置: %v\n", result.Positions)
	}
}

func handleDemo(serverURL string, args []string) {
	resp, err := http.Get(serverURL + "/demo")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result api.DemoResponse
	json.Unmarshal(body, &result)

	if !result.Success {
		fmt.Printf("错误: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Println("=== 后缀数组服务演示 ===")
	fmt.Println()
	fmt.Printf("示例文本: %s\n", result.Text)
	fmt.Println()
	fmt.Printf("后缀数组 (SA) 长度: %d\n", len(result.SA))
	fmt.Printf("LCP数组 长度: %d\n", len(result.LCP))
	fmt.Println()

	fmt.Println("--- 最长重复子串 ---")
	if len(result.LongestRepeated) == 0 {
		fmt.Println("没有找到重复子串")
	} else {
		for i, s := range result.LongestRepeated {
			fmt.Printf("  %d. \"%s\" 出现位置: %v\n", i+1, s.Substring, s.Positions)
		}
	}
	fmt.Println()

	fmt.Println("--- 示例查询 ---")
	for _, q := range result.ExampleQueries {
		fmt.Printf("  \"%s\" 出现 %d 次，位置: %v\n", q.Pattern, q.Count, q.Positions)
	}
}
