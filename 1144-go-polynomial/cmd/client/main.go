package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"polynomial-service/api"
)

const serverURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	operation := os.Args[1]

	switch operation {
	case "add", "sub", "mul", "div":
		handleBinaryOp(operation)
	case "eval":
		handleEval()
	case "factor", "deriv":
		handleSingleOp(operation)
	default:
		fmt.Printf("未知操作: %s\n", operation)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("多项式运算客户端工具")
	fmt.Println("用法:")
	fmt.Println("  poly add \"<多项式1>\" \"<多项式2>\"")
	fmt.Println("  poly sub \"<多项式1>\" \"<多项式2>\"")
	fmt.Println("  poly mul \"<多项式1>\" \"<多项式2>\"")
	fmt.Println("  poly div \"<被除数>\" \"<除数>\"")
	fmt.Println("  poly eval \"<多项式>\" <x值>")
	fmt.Println("  poly factor \"<多项式>\"")
	fmt.Println("  poly deriv \"<多项式>\"")
	fmt.Println()
	fmt.Println("多项式格式: [系数数组，从低次到高次]")
	fmt.Println("示例:")
	fmt.Println("  poly add \"[1,2,3]\" \"[4,5]\"")
	fmt.Println("  poly eval \"[1,0,-2]\" 1.414")
}

func handleBinaryOp(operation string) {
	if len(os.Args) != 4 {
		fmt.Printf("%s 需要两个多项式参数\n", operation)
		os.Exit(1)
	}

	a, err := parsePolynomial(os.Args[2])
	if err != nil {
		fmt.Printf("解析第一个多项式失败: %v\n", err)
		os.Exit(1)
	}

	b, err := parsePolynomial(os.Args[3])
	if err != nil {
		fmt.Printf("解析第二个多项式失败: %v\n", err)
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/%s", serverURL, operation)
	req := api.BinaryOpRequest{A: a, B: b}

	if operation == "div" {
		var resp api.DivResponse
		if err := sendRequest(url, req, &resp); err != nil {
			fmt.Printf("请求失败: %v\n", err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Printf("错误: %s\n", resp.Error)
			os.Exit(1)
		}
		fmt.Printf("商: %v\n", resp.Quotient)
		fmt.Printf("余数: %v\n", resp.Remainder)
		return
	}

	var resp api.PolynomialResponse
	if err := sendRequest(url, req, &resp); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}
	if !resp.Success {
		fmt.Printf("错误: %s\n", resp.Error)
		os.Exit(1)
	}
	fmt.Printf("结果: %v\n", resp.Result)
}

func handleEval() {
	if len(os.Args) != 4 {
		fmt.Println("eval 需要两个参数: 多项式和x值")
		os.Exit(1)
	}

	p, err := parsePolynomial(os.Args[2])
	if err != nil {
		fmt.Printf("解析多项式失败: %v\n", err)
		os.Exit(1)
	}

	x, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		fmt.Printf("解析x值失败: %v\n", err)
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/eval", serverURL)
	req := api.EvalRequest{Polynomial: p, X: x}

	var resp api.EvalResponse
	if err := sendRequest(url, req, &resp); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}
	if !resp.Success {
		fmt.Printf("错误: %s\n", resp.Error)
		os.Exit(1)
	}
	fmt.Printf("结果: %v\n", resp.Result)
}

func handleSingleOp(operation string) {
	if len(os.Args) != 3 {
		fmt.Printf("%s 需要一个多项式参数\n", operation)
		os.Exit(1)
	}

	p, err := parsePolynomial(os.Args[2])
	if err != nil {
		fmt.Printf("解析多项式失败: %v\n", err)
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/%s", serverURL, operation)
	req := api.SinglePolynomialRequest{Polynomial: p}

	if operation == "factor" {
		var resp api.FactorResponse
		if err := sendRequest(url, req, &resp); err != nil {
			fmt.Printf("请求失败: %v\n", err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Printf("错误: %s\n", resp.Error)
			os.Exit(1)
		}
		fmt.Printf("结果: %s\n", resp.Result)
		return
	}

	var resp api.PolynomialResponse
	if err := sendRequest(url, req, &resp); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}
	if !resp.Success {
		fmt.Printf("错误: %s\n", resp.Error)
		os.Exit(1)
	}
	fmt.Printf("结果: %v\n", resp.Result)
}

func parsePolynomial(s string) ([]float64, error) {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '[' || s[len(s)-1] != ']' {
		return nil, fmt.Errorf("格式错误，应为 [系数列表]")
	}

	content := s[1 : len(s)-1]
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("空数组")
	}

	parts := strings.Split(content, ",")
	result := make([]float64, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		val, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return nil, fmt.Errorf("第 %d 个系数不是有效数字: %s", i+1, part)
		}
		result[i] = val
	}

	return result, nil
}

func sendRequest(url string, req interface{}, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpResp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	return json.NewDecoder(httpResp.Body).Decode(resp)
}
