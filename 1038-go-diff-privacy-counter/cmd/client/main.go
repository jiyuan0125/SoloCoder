package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"diff-privacy-counter/pkg/models"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

func (c *Client) QueryCount(category string, epsilon float64) (*models.CountQueryResponse, error) {
	reqBody := models.CountQueryRequest{
		Category: category,
		Epsilon:  epsilon,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.baseURL+"/api/count", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("server error: %s", string(body))
		}
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var countResp models.CountQueryResponse
	if err := json.Unmarshal(body, &countResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &countResp, nil
}

func (c *Client) GetBudget() (*models.BudgetResponse, error) {
	resp, err := http.Get(c.baseURL + "/api/budget")
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("server error: %s", string(body))
		}
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var budgetResp models.BudgetResponse
	if err := json.Unmarshal(body, &budgetResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &budgetResp, nil
}

func (c *Client) RechargeBudget(amount float64) (*models.BudgetResponse, error) {
	reqBody := models.RechargeRequest{Amount: amount}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.baseURL+"/api/budget/recharge", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("server error: %s", string(body))
		}
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var budgetResp models.BudgetResponse
	if err := json.Unmarshal(body, &budgetResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &budgetResp, nil
}

func printUsage() {
	fmt.Println("差分隐私计数服务客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client query <category> [epsilon]   查询某个类别的计数")
	fmt.Println("  client budget                        查看剩余隐私预算")
	fmt.Println("  client recharge <amount>             充值隐私预算（测试用）")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  client query apple 1.0      # 查询apple的计数，使用epsilon=1.0")
	fmt.Println("  client query apple 0        # 查询apple的精确计数（不消耗预算）")
	fmt.Println("  client budget               # 查看剩余预算")
	fmt.Println("  client recharge 5.0         # 充值5.0的预算")
	fmt.Println()
	fmt.Println("说明:")
	fmt.Println("  - epsilon=0: 返回精确计数，不消耗预算，可无限次查询")
	fmt.Println("  - epsilon>0: 返回加噪计数，消耗相应的预算")
	fmt.Println("  - epsilon越小，隐私保护越强，结果准确度越低")
}

func main() {
	serverURL := "http://localhost:8080"
	if envURL := os.Getenv("SERVER_URL"); envURL != "" {
		serverURL = envURL
	}

	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(serverURL)

	switch args[0] {
	case "query":
		if len(args) < 2 {
			fmt.Println("错误: 缺少category参数")
			printUsage()
			os.Exit(1)
		}

		category := args[1]
		epsilon := 1.0

		if len(args) >= 3 {
			eps, err := strconv.ParseFloat(args[2], 64)
			if err != nil {
				fmt.Printf("错误: 无效的epsilon值: %s\n", args[2])
				os.Exit(1)
			}
			epsilon = eps
		}

		resp, err := client.QueryCount(category, epsilon)
		if err != nil {
			fmt.Printf("查询失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("类别: %s\n", category)
		fmt.Printf("计数: %d\n", resp.Count)
		if resp.Message != "" {
			fmt.Printf("说明: %s\n", resp.Message)
		}
		if resp.Epsilon > 0 {
			fmt.Printf("消耗预算: %.2f\n", resp.Epsilon)
		}

	case "budget":
		resp, err := client.GetBudget()
		if err != nil {
			fmt.Printf("查询失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("总预算: %.2f\n", resp.TotalBudget)
		fmt.Printf("剩余预算: %.2f\n", resp.RemainingBudget)
		if resp.Message != "" {
			fmt.Printf("说明: %s\n", resp.Message)
		}

	case "recharge":
		if len(args) < 2 {
			fmt.Println("错误: 缺少amount参数")
			printUsage()
			os.Exit(1)
		}

		amount, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			fmt.Printf("错误: 无效的充值金额: %s\n", args[1])
			os.Exit(1)
		}

		resp, err := client.RechargeBudget(amount)
		if err != nil {
			fmt.Printf("充值失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("充值成功!\n")
		fmt.Printf("总预算: %.2f\n", resp.TotalBudget)
		fmt.Printf("剩余预算: %.2f\n", resp.RemainingBudget)
		if resp.Message != "" {
			fmt.Printf("说明: %s\n", resp.Message)
		}

	default:
		fmt.Printf("未知命令: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}
}
