package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"phonetic-search/api"
)

const defaultServerURL = "http://localhost:8304"

func getServerURL() string {
	if envURL := os.Getenv("PHONETIC_SERVER_URL"); envURL != "" {
		return envURL
	}
	return defaultServerURL
}

func postJSON(url string, body interface{}, resp interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return err
	}

	respHTTP, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer respHTTP.Body.Close()

	data, err := io.ReadAll(respHTTP.Body)
	if err != nil {
		return err
	}

	if resp != nil {
		return json.Unmarshal(data, resp)
	}
	return nil
}

func getJSON(url string, resp interface{}) error {
	respHTTP, err := http.Get(url)
	if err != nil {
		return err
	}
	defer respHTTP.Body.Close()

	data, err := io.ReadAll(respHTTP.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, resp)
}

func cmdEncode(name string, serverURL string) {
	var resp api.EncodeResponse
	if err := postJSON(serverURL+"/encode", api.EncodeRequest{Name: name}, &resp); err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("人名: %s\n", resp.Name)
	fmt.Printf("Soundex: %s\n", resp.Soundex)
	fmt.Printf("Metaphone: %s\n", resp.Metaphone)
	if resp.Message != "" {
		fmt.Printf("提示: %s\n", resp.Message)
	}
}

func cmdSearch(name string, serverURL string) {
	var resp api.SearchResponse
	if err := postJSON(serverURL+"/search", api.SearchRequest{Name: name}, &resp); err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("查询人名: %s\n", resp.QueryName)
	fmt.Printf("Soundex: %s\n", resp.Soundex)
	fmt.Printf("Metaphone: %s\n", resp.Metaphone)
	if resp.Message != "" {
		fmt.Printf("提示: %s\n", resp.Message)
	}
	fmt.Println()
	fmt.Printf("找到 %d 个发音相似的人名:\n", len(resp.Matches))
	for i, name := range resp.Matches {
		fmt.Printf("  %d. %s\n", i+1, name)
	}
}

func cmdAdd(name string, serverURL string) {
	var resp api.AddResponse
	if err := postJSON(serverURL+"/add", api.AddRequest{Name: name}, &resp); err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	if resp.Added {
		fmt.Printf("已添加人名: %s\n", resp.Name)
	} else {
		fmt.Printf("添加失败: %s - %s\n", resp.Name, resp.Message)
	}
}

func cmdImport(filename string, serverURL string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("无法打开文件: %v\n", err)
		return
	}
	defer file.Close()

	names := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			names = append(names, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("读取文件错误: %v\n", err)
		return
	}

	var resp api.ImportResponse
	if err := postJSON(serverURL+"/import", api.ImportRequest{Names: names}, &resp); err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("导入完成:\n")
	fmt.Printf("  总数: %d\n", resp.Total)
	fmt.Printf("  新增: %d\n", resp.Added)
	fmt.Printf("  已存在: %d\n", resp.Existing)
	if len(resp.Invalid) > 0 {
		fmt.Printf("  无效: %d\n", len(resp.Invalid))
	}
}

func cmdStats(serverURL string) {
	var resp api.StatsResponse
	if err := getJSON(serverURL+"/stats", &resp); err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("人名库统计:\n")
	fmt.Printf("  总人数: %d\n", resp.TotalNames)
	fmt.Printf("  不同Soundex编码数: %d\n", resp.UniqueSoundex)
	fmt.Printf("  不同Metaphone编码数: %d\n", resp.UniqueMetaphone)

	fmt.Println()
	fmt.Println("Soundex编码冲突最多的前5个:")
	for i, code := range resp.TopSoundexCodes {
		fmt.Printf("  %d. %s (%d人): %s\n", i+1, code.Code, code.Count, strings.Join(code.Names, ", "))
	}

	fmt.Println()
	fmt.Println("Metaphone编码冲突最多的前5个:")
	for i, code := range resp.TopMetaphoneCodes {
		fmt.Printf("  %d. %s (%d人): %s\n", i+1, code.Code, code.Count, strings.Join(code.Names, ", "))
	}
}

func printUsage() {
	fmt.Println("用法: phonetic-client <命令> [参数]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  encode <人名>     对人名进行编码，返回Soundex和Metaphone")
	fmt.Println("  search <人名>     在人名库中搜索发音相似的人名")
	fmt.Println("  add <人名>        添加人名到人名库")
	fmt.Println("  import <文件路径> 从文件批量导入人名（每行一个）")
	fmt.Println("  stats             查看人名库统计信息")
	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  PHONETIC_SERVER_URL  服务端地址（默认: http://localhost:8304）")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := getServerURL()
	cmd := os.Args[1]

	switch cmd {
	case "encode":
		if len(os.Args) < 3 {
			fmt.Println("错误: 请提供要编码的人名")
			os.Exit(1)
		}
		cmdEncode(strings.Join(os.Args[2:], " "), serverURL)
	case "search":
		if len(os.Args) < 3 {
			fmt.Println("错误: 请提供要搜索的人名")
			os.Exit(1)
		}
		cmdSearch(strings.Join(os.Args[2:], " "), serverURL)
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("错误: 请提供要添加的人名")
			os.Exit(1)
		}
		cmdAdd(strings.Join(os.Args[2:], " "), serverURL)
	case "import":
		if len(os.Args) < 3 {
			fmt.Println("错误: 请提供要导入的文件路径")
			os.Exit(1)
		}
		cmdImport(os.Args[2], serverURL)
	case "stats":
		cmdStats(serverURL)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
