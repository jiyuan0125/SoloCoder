package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"segment-tree/model"
)

const (
	defaultServerURL = "http://localhost:8080/process"
)

func main() {
	serverURL := flag.String("server", defaultServerURL, "服务端地址 (例如: http://localhost:8080/process)")
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println("用法: client [--server=http://host:port/process] <JSON输入>")
		fmt.Println("示例:")
		fmt.Println(`  client '{"init_values":[1,2,3,4,5],"actions":[{"type":"range_sum","left":0,"right":4}]}'`)
		fmt.Println()
		fmt.Println("JSON 格式:")
		fmt.Println(`  {`)
		fmt.Println(`    "init_values": [1, 2, 3, 4, 5],`)
		fmt.Println(`    "actions": [`)
		fmt.Println(`      {"type": "range_add", "left": 0, "right": 2, "value": 10},`)
		fmt.Println(`      {"type": "range_sum", "left": 0, "right": 4},`)
		fmt.Println(`      {"type": "range_max", "left": 3, "right": 4}`)
		fmt.Println(`    ]`)
		fmt.Println(`  }`)
		os.Exit(1)
	}

	jsonInput := flag.Arg(0)

	var req model.Request
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		fmt.Fprintf(os.Stderr, "JSON 解析失败: %v\n", err)
		os.Exit(1)
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "请求序列化失败: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(*serverURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取响应失败: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp model.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			fmt.Fprintf(os.Stderr, "服务端错误: %s\n", errResp.Error)
		} else {
			fmt.Fprintf(os.Stderr, "HTTP 错误 %d: %s\n", resp.StatusCode, string(body))
		}
		os.Exit(1)
	}

	var result model.Response
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Fprintf(os.Stderr, "解析响应失败: %v\n", err)
		os.Exit(1)
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "输出序列化失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}
