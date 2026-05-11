package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go-enum-generator/pkg/api"
	"go-enum-generator/pkg/enum"
)

func main() {
	localMode := flag.Bool("local", true, "本地模式（直接在本地生成，不调用服务端）")
	serverURL := flag.String("server", "http://localhost:8080", "服务端地址（仅在非本地模式下使用）")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "用法: go-enum-client [选项] <源文件路径>")
		fmt.Fprintln(os.Stderr, "选项:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	sourceFile := flag.Arg(0)

	sourceBytes, err := os.ReadFile(sourceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取源文件失败: %v\n", err)
		os.Exit(1)
	}

	pkgName, err := extractPackageName(sourceFile, sourceBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析包名失败: %v\n", err)
		os.Exit(1)
	}

	var generatedCode []byte

	if *localMode {
		enums, parseErr := enum.Parse(bytes.NewReader(sourceBytes), sourceFile)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "解析源文件失败: %v\n", parseErr)
			os.Exit(1)
		}

		if len(enums) == 0 {
			fmt.Println("没有找到需要生成的枚举")
			return
		}

		generatedCode, err = enum.Generate(pkgName, enums)
		if err != nil {
			fmt.Fprintf(os.Stderr, "生成代码失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		req := api.GenerateRequest{
			SourceCode: string(sourceBytes),
			Filename:   sourceFile,
		}
		reqBody, _ := json.Marshal(req)

		resp, httpErr := http.Post(*serverURL+"/api/generate", "application/json", bytes.NewReader(reqBody))
		if httpErr != nil {
			fmt.Fprintf(os.Stderr, "调用服务端失败: %v\n", httpErr)
			os.Exit(1)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "服务端返回错误: %s\n", string(respBody))
			os.Exit(1)
		}

		var genResp api.GenerateResponse
		if err = json.Unmarshal(respBody, &genResp); err != nil {
			fmt.Fprintf(os.Stderr, "解析响应失败: %v\n", err)
			os.Exit(1)
		}

		if genResp.Message != "" {
			fmt.Println(genResp.Message)
			return
		}

		generatedCode = []byte(genResp.GeneratedCode)
	}

	outputFile := generateOutputPath(sourceFile)
	if err = os.WriteFile(outputFile, generatedCode, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "写入输出文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已生成: %s\n", outputFile)
}

func extractPackageName(filename string, content []byte) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, content, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}
	return file.Name.Name, nil
}

func generateOutputPath(sourceFile string) string {
	ext := filepath.Ext(sourceFile)
	base := strings.TrimSuffix(sourceFile, ext)
	return base + "_enum.go"
}
