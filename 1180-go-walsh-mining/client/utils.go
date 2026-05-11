package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var serverURL = "http://localhost:8404"

func init() {
	if envURL := os.Getenv("SERVER_URL"); envURL != "" {
		serverURL = envURL
	}
}

func httpPost(endpoint string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	resp, err := http.Post(serverURL+endpoint, "application/json", bodyReader)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func httpGet(endpoint string) ([]byte, error) {
	resp, err := http.Get(serverURL + endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func printUsage() {
	fmt.Println("WALS Mining 命令行客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  wals-client <command> [options]")
	fmt.Println()
	fmt.Println("子命令:")
	fmt.Println("  import <file>     导入交易数据（JSON格式，每行一条交易）")
	fmt.Println("  weight <json>     设置商品权重（JSON格式，如 {\"A\":0.8, \"B\":0.5}）")
	fmt.Println("  mine <minSup> <minConf>  执行挖掘，输出前20条置信度最高的规则")
	fmt.Println("  rules             查看所有关联规则")
	fmt.Println("  items             查看所有频繁项集")
	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  SERVER_URL        服务端地址（默认: http://localhost:8404）")
	os.Exit(1)
}

func formatRule(antecedent, consequent []string, confidence, support float64) string {
	return fmt.Sprintf("%s → %s (置信度: %.2f, 支持度: %.2f)",
		strings.Join(antecedent, ", "),
		strings.Join(consequent, ", "),
		confidence,
		support)
}
