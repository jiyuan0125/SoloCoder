package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultServer = "http://localhost:8100"
)

type Config struct {
	Mode      string
	Input     string
	Output    string
	Server    string
	LocalMode bool
}

func main() {
	cfg := parseFlags()

	if cfg.LocalMode {
		runLocalMode(cfg)
	} else {
		runServerMode(cfg)
	}
}

func parseFlags() Config {
	cfg := Config{}

	decodeCmd := flag.NewFlagSet("decode", flag.ExitOnError)
	encodeCmd := flag.NewFlagSet("encode", flag.ExitOnError)

	var decodeInput, decodeOutput, decodeServer string
	var decodeLocal bool
	decodeCmd.StringVar(&decodeInput, "input", "", "输入文件路径（支持.bencode和.torrent）")
	decodeCmd.StringVar(&decodeInput, "i", "", "输入文件路径（简写）")
	decodeCmd.StringVar(&decodeOutput, "output", "", "输出文件路径，默认stdout")
	decodeCmd.StringVar(&decodeOutput, "o", "", "输出文件路径（简写）")
	decodeCmd.StringVar(&decodeServer, "server", defaultServer, "服务端地址")
	decodeCmd.StringVar(&decodeServer, "s", defaultServer, "服务端地址（简写）")
	decodeCmd.BoolVar(&decodeLocal, "local", false, "本地处理模式，不调用服务端")
	decodeCmd.BoolVar(&decodeLocal, "l", false, "本地处理模式（简写）")

	var encodeInput, encodeOutput, encodeServer string
	var encodeLocal bool
	encodeCmd.StringVar(&encodeInput, "input", "", "输入JSON文件路径")
	encodeCmd.StringVar(&encodeInput, "i", "", "输入JSON文件路径（简写）")
	encodeCmd.StringVar(&encodeOutput, "output", "", "输出文件路径，默认stdout")
	encodeCmd.StringVar(&encodeOutput, "o", "", "输出文件路径（简写）")
	encodeCmd.StringVar(&encodeServer, "server", defaultServer, "服务端地址")
	encodeCmd.StringVar(&encodeServer, "s", defaultServer, "服务端地址（简写）")
	encodeCmd.BoolVar(&encodeLocal, "local", false, "本地处理模式，不调用服务端")
	encodeCmd.BoolVar(&encodeLocal, "l", false, "本地处理模式（简写）")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "decode":
		decodeCmd.Parse(os.Args[2:])
		cfg.Mode = "decode"
		cfg.Input = decodeInput
		cfg.Output = decodeOutput
		cfg.Server = decodeServer
		cfg.LocalMode = decodeLocal
	case "encode":
		encodeCmd.Parse(os.Args[2:])
		cfg.Mode = "encode"
		cfg.Input = encodeInput
		cfg.Output = encodeOutput
		cfg.Server = encodeServer
		cfg.LocalMode = encodeLocal
	case "-h", "--help", "help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}

	return cfg
}

func printUsage() {
	fmt.Println("Bencode编解码工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  bencode decode [options] - 从Bencode解码到JSON")
	fmt.Println("  bencode encode [options] - 从JSON编码到Bencode")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -i, --input <path>   输入文件路径（不指定则从stdin读取）")
	fmt.Println("  -o, --output <path>  输出文件路径（不指定则输出到stdout）")
	fmt.Println("  -s, --server <url>   服务端地址（默认: http://localhost:8100）")
	fmt.Println("  -l, --local          本地处理模式，不调用服务端")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  bencode decode -i test.torrent -o test.json")
	fmt.Println("  cat test.torrent | bencode decode > test.json")
	fmt.Println("  bencode encode -i test.json -o test.torrent")
	fmt.Println("  cat test.json | bencode encode > test.torrent")
}

func runLocalMode(cfg Config) {
	input, err := readInput(cfg.Input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		os.Exit(1)
	}

	var result []byte

	if cfg.Mode == "decode" {
		result, err = decodeLocal(input)
	} else {
		result, err = encodeLocal(input)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "处理失败: %v\n", err)
		os.Exit(1)
	}

	if err := writeOutput(result, cfg.Output); err != nil {
		fmt.Fprintf(os.Stderr, "写入输出失败: %v\n", err)
		os.Exit(1)
	}
}

func runServerMode(cfg Config) {
	var result []byte
	var err error

	if cfg.Mode == "decode" {
		result, err = decodeRemote(cfg)
	} else {
		result, err = encodeRemote(cfg)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "处理失败: %v\n", err)
		os.Exit(1)
	}

	if err := writeOutput(result, cfg.Output); err != nil {
		fmt.Fprintf(os.Stderr, "写入输出失败: %v\n", err)
		os.Exit(1)
	}
}

func readInput(path string) ([]byte, error) {
	if path == "" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func writeOutput(data []byte, path string) error {
	if path == "" {
		_, err := os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func decodeLocal(input []byte) ([]byte, error) {
	value, err := decodeBencode(input)
	if err != nil {
		return nil, err
	}
	return toJSON(value)
}

func encodeLocal(input []byte) ([]byte, error) {
	value, err := fromJSON(input)
	if err != nil {
		return nil, err
	}
	return encodeBencode(value)
}

func decodeRemote(cfg Config) ([]byte, error) {
	url := strings.TrimSuffix(cfg.Server, "/") + "/decode"

	if cfg.Input != "" {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		file, err := os.Open(cfg.Input)
		if err != nil {
			return nil, fmt.Errorf("打开文件失败: %v", err)
		}
		defer file.Close()

		part, err := writer.CreateFormFile("file", filepath.Base(cfg.Input))
		if err != nil {
			return nil, fmt.Errorf("创建表单部分失败: %v", err)
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return nil, fmt.Errorf("复制文件内容失败: %v", err)
		}

		writer.Close()

		req, err := http.NewRequest("POST", url, body)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("发送请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("服务端返回错误: %s", string(body))
		}

		return io.ReadAll(resp.Body)
	}

	input, err := readInput("")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("服务端返回错误: %s", string(body))
	}

	return io.ReadAll(resp.Body)
}

func encodeRemote(cfg Config) ([]byte, error) {
	url := strings.TrimSuffix(cfg.Server, "/") + "/encode"

	input, err := readInput(cfg.Input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("服务端返回错误: %s", string(body))
	}

	return io.ReadAll(resp.Body)
}
