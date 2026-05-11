package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/example/avmetadata/pkg/api"
)

func main() {
	var serverAddr string
	var filePath string
	var timeout int

	flag.StringVar(&serverAddr, "server", "", "服务端地址")
	flag.StringVar(&filePath, "file", "", "要解析的音视频文件路径")
	flag.IntVar(&timeout, "timeout", 30, "请求超时时间（秒）")
	flag.Parse()

	if serverAddr == "" {
		serverAddr = os.Getenv("AVMETA_SERVER")
		if serverAddr == "" {
			serverAddr = "http://127.0.0.1:8080"
		}
	}

	if filePath == "" {
		if flag.NArg() > 0 {
			filePath = flag.Arg(0)
		} else {
			fmt.Fprintln(os.Stderr, "用法: avmeta-client [options] <文件路径>")
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "选项:")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "错误: 文件不存在: %s\n", filePath)
		os.Exit(1)
	}

	reqBody := api.MetadataRequest{
		FilePath: filePath,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 序列化请求失败: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	url := fmt.Sprintf("%s/api/metadata", serverAddr)
	resp, err := client.Post(url, "application/json", bytes.NewReader(jsonData))
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

	var apiResp api.MetadataResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 解析响应失败: %v\n", err)
		fmt.Fprintf(os.Stderr, "原始响应: %s\n", string(body))
		os.Exit(1)
	}

	if !apiResp.Success {
		fmt.Fprintf(os.Stderr, "错误: 服务端返回失败: %s\n", apiResp.Message)
		os.Exit(1)
	}

	printMetadata(apiResp.Metadata)
}

func printMetadata(meta *api.Metadata) {
	if meta == nil {
		return
	}

	fmt.Println("=== 音视频元数据 ===")
	fmt.Printf("格式: %s\n", meta.Format)
	fmt.Printf("文件大小: %d 字节\n", meta.FileSize)
	if meta.Duration > 0 {
		fmt.Printf("时长: %.2f 秒\n", meta.Duration)
	}
	if meta.Bitrate > 0 {
		fmt.Printf("比特率: %d bps (%.1f kbps)\n", meta.Bitrate, float64(meta.Bitrate)/1000.0)
	}

	if len(meta.AudioTags) > 0 {
		fmt.Println("\n音频标签:")
		for key, value := range meta.AudioTags {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	if len(meta.FormatTags) > 0 {
		fmt.Println("\n格式标签:")
		for key, value := range meta.FormatTags {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}
}
