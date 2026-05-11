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

	"cardcheck/internal/models"
)

const (
	defaultServer = "http://localhost:8500"
	redText       = "\033[31m"
	greenText     = "\033[32m"
	yellowText    = "\033[33m"
	resetText     = "\033[0m"
)

func main() {
	var cardNumber string
	var filePath string
	var serverURL string

	flag.StringVar(&cardNumber, "c", "", "单个银行卡号")
	flag.StringVar(&filePath, "f", "", "批量卡号文件路径(每行一个卡号)")
	flag.StringVar(&serverURL, "s", defaultServer, "服务端地址")
	flag.Usage = customUsage
	flag.Parse()

	if cardNumber == "" && filePath == "" {
		flag.Usage()
		os.Exit(1)
	}

	if cardNumber != "" {
		err := validateSingleCard(cardNumber, serverURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if filePath != "" {
		err := validateBatchFile(filePath, serverURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}
	}
}

func customUsage() {
	fmt.Fprintf(os.Stderr, "银行卡号校验客户端\n\n")
	fmt.Fprintf(os.Stderr, "用法:\n")
	fmt.Fprintf(os.Stderr, "  client -c <银行卡号> [-s 服务端地址]\n")
	fmt.Fprintf(os.Stderr, "  client -f <卡号文件路径> [-s 服务端地址]\n\n")
	fmt.Fprintf(os.Stderr, "参数:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n示例:\n")
	fmt.Fprintf(os.Stderr, "  client -c 6222021234567890123\n")
	fmt.Fprintf(os.Stderr, "  client -f cards.txt\n")
}

func validateSingleCard(cardNumber, serverURL string) error {
	reqBody, err := json.Marshal(models.ValidateRequest{CardNumber: cardNumber})
	if err != nil {
		return fmt.Errorf("构建请求失败: %w", err)
	}

	resp, err := http.Post(serverURL+"/validate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("请求服务端失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	var result models.ValidateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	printSingleResult(result)
	return nil
}

func validateBatchFile(filePath, serverURL string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	var cardNumbers []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			cardNumbers = append(cardNumbers, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	if len(cardNumbers) == 0 {
		return fmt.Errorf("文件中没有有效的卡号")
	}

	reqBody, err := json.Marshal(models.BatchValidateRequest{CardNumbers: cardNumbers})
	if err != nil {
		return fmt.Errorf("构建请求失败: %w", err)
	}

	resp, err := http.Post(serverURL+"/batch-validate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("请求服务端失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	var result models.BatchValidateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	printBatchResults(result.Results)
	return nil
}

func printSingleResult(result models.ValidateResponse) {
	fmt.Println("==================================================")
	fmt.Printf("原始卡号: %s\n", result.CardNumber)
	if result.CleanNumber != "" {
		fmt.Printf("标准卡号: %s\n", result.CleanNumber)
	}
	
	if result.IsValid {
		fmt.Printf("校验状态: %s✓ 有效%s\n", greenText, resetText)
	} else {
		fmt.Printf("校验状态: %s✗ 无效%s\n", redText, resetText)
		fmt.Printf("失败原因: %s%s%s\n", redText, result.ErrorReason, resetText)
	}
	
	if result.CardOrganization != "" && result.CardOrganization != models.CardOrgUnknown {
		fmt.Printf("卡组织:   %s\n", result.CardOrganization)
	}
	
	if result.BankName != "" && result.BankName != "未知" {
		fmt.Printf("发卡行:   %s\n", result.BankName)
	}
	
	if result.CardType != "" && result.CardType != models.CardTypeUnknown {
		fmt.Printf("卡种:     %s\n", result.CardType)
	}
	
	if result.BIN != "" {
		fmt.Printf("BIN号:    %s\n", result.BIN)
	}
	fmt.Println("==================================================")
}

func printBatchResults(results []models.ValidateResponse) {
	fmt.Println("==================================================")
	fmt.Println("批量校验结果")
	fmt.Println("==================================================")

	validCount := 0
	invalidCount := 0

	for i, result := range results {
		fmt.Printf("\n[%d/%d] 卡号: %s\n", i+1, len(results), result.CardNumber)
		
		if result.IsValid {
			validCount++
			fmt.Printf("  状态: %s有效%s\n", greenText, resetText)
		} else {
			invalidCount++
			fmt.Printf("  状态: %s无效%s\n", redText, resetText)
			fmt.Printf("  原因: %s%s%s\n", redText, result.ErrorReason, resetText)
		}
		
		if result.CardOrganization != "" {
			fmt.Printf("  卡组织: %s\n", result.CardOrganization)
		}
		
		if result.BankName != "" {
			fmt.Printf("  发卡行: %s\n", result.BankName)
		}
		
		if result.CardType != "" {
			fmt.Printf("  卡种: %s\n", result.CardType)
		}
	}

	fmt.Println("\n==================================================")
	fmt.Printf("总计: %d 张, %s有效%d张%s, %s无效%d张%s\n", 
		len(results), greenText, validCount, resetText, redText, invalidCount, resetText)
	fmt.Println("==================================================")
}
