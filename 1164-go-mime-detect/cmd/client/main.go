package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/example/mimedetect/api"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverAddr := flag.String("server", "http://localhost:8080", "MIME检测服务地址")
	flag.Parse()

	filePath := flag.Arg(0)
	if filePath == "" {
		printUsage()
		os.Exit(1)
	}

	if _, err := os.Stat(filePath); err != nil {
		fmt.Printf("错误: 无法访问文件 %s: %v\n", filePath, err)
		os.Exit(1)
	}

	reqBody, err := json.Marshal(api.DetectRequest{FilePath: filePath})
	if err != nil {
		fmt.Printf("错误: 构建请求失败: %v\n", err)
		os.Exit(1)
	}

	url := *serverAddr + "/detect"
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("错误: 无法连接到服务端: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("错误: 读取响应失败: %v\n", err)
		os.Exit(1)
	}

	var detectResp api.DetectResponse
	if err := json.Unmarshal(body, &detectResp); err != nil {
		fmt.Printf("错误: 解析响应失败: %v\n", err)
		os.Exit(1)
	}

	if !detectResp.Success {
		fmt.Printf("检测失败: %s\n", detectResp.Error)
		os.Exit(1)
	}

	printResult(detectResp)
}

func printUsage() {
	fmt.Println("使用方法: mimedetect-client [选项] <文件路径>")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -server string    MIME检测服务地址 (默认: http://localhost:8080)")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  mimedetect-client ./test.pdf")
	fmt.Println("  mimedetect-client -server http://192.168.1.100:8080 ./document.docx")
}

func printResult(resp api.DetectResponse) {
	fmt.Printf("文件: %s\n", resp.FilePath)
	fmt.Println("----------------------------------------")
	fmt.Printf("基于扩展名: %s\n", resp.ExtensionMIME)
	fmt.Printf("基于魔数:   %s\n", resp.MagicMIME)
	fmt.Printf("最终结论:   %s\n", resp.FinalMIME)
	fmt.Printf("可信度:     %s\n", resp.Confidence)

	if resp.HasWarning {
		fmt.Printf("警告:       %s\n", resp.Warning)
	}

	if !resp.IsConsistent && resp.MagicMIME != "application/octet-stream" {
		fmt.Println()
		fmt.Println("⚠️  注意: 扩展名与魔数不匹配，可能是伪装文件！")
	}
}
