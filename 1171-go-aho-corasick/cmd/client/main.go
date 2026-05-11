package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"ac-service/internal/api"
)

var baseURL string

func init() {
	baseURL = os.Getenv("SERVER_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8300"
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "add":
		handleAdd(args)
	case "search":
		handleSearch(args)
	case "dict":
		handleDict(args)
	case "clear":
		handleClear(args)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("AC自动机多模式匹配服务客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client <command> [arguments]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  add      添加模式串到词典")
	fmt.Println("  search   在文本中搜索匹配的模式串")
	fmt.Println("  dict     查看当前词典状态")
	fmt.Println("  clear    清空当前词典")
	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  SERVER_URL   服务端地址 (默认: http://localhost:8300)")
}

func handleAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	pattern := fs.String("p", "", "要添加的模式串")
	file := fs.String("f", "", "从文件批量导入模式串，每行一个")
	fs.Parse(args)

	if *pattern == "" && *file == "" {
		fmt.Println("错误: 必须指定 -p 或 -f 参数")
		fmt.Println()
		fmt.Println("用法:")
		fmt.Println("  client add -p <pattern>")
		fmt.Println("  client add -f <file>")
		os.Exit(1)
	}

	if *pattern != "" {
		addSinglePattern(*pattern)
	}

	if *file != "" {
		addPatternsFromFile(*file)
	}
}

func addSinglePattern(pattern string) {
	reqBody, _ := json.Marshal(api.AddPatternRequest{Pattern: pattern})
	resp, err := http.Post(baseURL+"/api/pattern/add", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.Unmarshal(body, &errResp)
		fmt.Printf("错误: %s\n", errResp.Error)
		os.Exit(1)
	}

	var addResp api.AddPatternResponse
	json.Unmarshal(body, &addResp)
	if addResp.Success {
		fmt.Println("模式串添加成功")
	} else {
		fmt.Printf("错误: %s\n", addResp.Message)
		os.Exit(1)
	}
}

func addPatternsFromFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("错误: 无法打开文件 %s: %v\n", filename, err)
		os.Exit(1)
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			patterns = append(patterns, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("错误: 读取文件失败: %v\n", err)
		os.Exit(1)
	}

	reqBody, _ := json.Marshal(api.AddPatternsRequest{Patterns: patterns})
	resp, err := http.Post(baseURL+"/api/pattern/add-batch", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.Unmarshal(body, &errResp)
		fmt.Printf("错误: %s\n", errResp.Error)
		os.Exit(1)
	}

	var addResp api.AddPatternsResponse
	json.Unmarshal(body, &addResp)
	if addResp.Success {
		fmt.Printf("批量添加完成: 成功添加 %d 个，跳过 %d 个\n", addResp.AddedCount, addResp.SkippedCount)
	} else {
		fmt.Printf("错误: %s\n", addResp.Message)
		os.Exit(1)
	}
}

func handleSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	text := fs.String("t", "", "要搜索的文本")
	file := fs.String("f", "", "从文件读取要搜索的文本")
	fs.Parse(args)

	if *text == "" && *file == "" {
		fmt.Println("错误: 必须指定 -t 或 -f 参数")
		fmt.Println()
		fmt.Println("用法:")
		fmt.Println("  client search -t <text>")
		fmt.Println("  client search -f <file>")
		os.Exit(1)
	}

	var searchText string
	if *text != "" {
		searchText = *text
	} else {
		fileContent, err := os.ReadFile(*file)
		if err != nil {
			fmt.Printf("错误: 无法读取文件 %s: %v\n", *file, err)
			os.Exit(1)
		}
		searchText = string(fileContent)
	}

	reqBody, _ := json.Marshal(api.SearchRequest{Text: searchText})
	resp, err := http.Post(baseURL+"/api/search", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.Unmarshal(body, &errResp)
		fmt.Printf("错误: %s\n", errResp.Error)
		os.Exit(1)
	}

	var searchResp api.SearchResponse
	json.Unmarshal(body, &searchResp)

	if !searchResp.Success {
		fmt.Printf("错误: %s\n", searchResp.Message)
		os.Exit(1)
	}

	if len(searchResp.Matches) == 0 {
		fmt.Println("未找到匹配的模式串")
		return
	}

	fmt.Printf("找到 %d 个匹配:\n", len(searchResp.Matches))
	fmt.Println()
	for i, match := range searchResp.Matches {
		fmt.Printf("%d. 模式串: %s, 位置: [%d, %d)\n", i+1, match.Pattern, match.Start, match.End)
	}
}

func handleDict(args []string) {
	resp, err := http.Get(baseURL + "/api/dict")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.Unmarshal(body, &errResp)
		fmt.Printf("错误: %s\n", errResp.Error)
		os.Exit(1)
	}

	var dictResp api.DictStatusResponse
	json.Unmarshal(body, &dictResp)

	if dictResp.Success {
		fmt.Printf("当前词典包含 %d 个模式串\n", dictResp.Count)
	} else {
		fmt.Printf("错误: %s\n", dictResp.Message)
		os.Exit(1)
	}
}

func handleClear(args []string) {
	reqBody, _ := json.Marshal(struct{}{})
	resp, err := http.Post(baseURL+"/api/clear", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.Unmarshal(body, &errResp)
		fmt.Printf("错误: %s\n", errResp.Error)
		os.Exit(1)
	}

	var clearResp api.ClearResponse
	json.Unmarshal(body, &clearResp)

	if clearResp.Success {
		fmt.Println("词典已清空")
	} else {
		fmt.Printf("错误: %s\n", clearResp.Message)
		os.Exit(1)
	}
}
