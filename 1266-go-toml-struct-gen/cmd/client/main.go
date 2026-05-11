package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/example/toml2struct/pkg/common"
)

func main() {
	var outputFile string
	var packageName string
	var serverURL string

	flag.StringVar(&outputFile, "o", "", "输出文件路径 (默认输出到stdout)")
	flag.StringVar(&packageName, "pkg", "main", "Go包名")
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "服务端地址")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "用法: toml2struct <input.toml> [-o output.go] [-pkg package]")
		os.Exit(1)
	}

	inputFile := args[0]

	tomlContent, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取文件失败: %v\n", err)
		os.Exit(1)
	}

	req := common.GenerateRequest{
		TomlContent: string(tomlContent),
		PackageName: packageName,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "序列化请求失败: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/codegen/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "请求服务端失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取响应失败: %v\n", err)
		os.Exit(1)
	}

	var respData common.GenerateResponse
	if err := json.Unmarshal(body, &respData); err != nil {
		fmt.Fprintf(os.Stderr, "解析响应失败: %v\n", err)
		fmt.Fprintf(os.Stderr, "响应内容: %s\n", string(body))
		os.Exit(1)
	}

	if !respData.Success {
		fmt.Fprintf(os.Stderr, "生成失败: %s\n", respData.Error)
		os.Exit(1)
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(respData.Code), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "写入文件失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("成功写入: %s\n", outputFile)
	} else {
		fmt.Print(respData.Code)
	}
}
