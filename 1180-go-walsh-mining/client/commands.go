package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"walsh-mining/common"
)

func handleImport(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "错误: 请指定要导入的文件路径")
		os.Exit(1)
	}

	filePath := args[0]
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法打开文件 %s: %v\n", filePath, err)
		os.Exit(1)
	}
	defer file.Close()

	var transactions [][]string
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var items []string
		if err := json.Unmarshal([]byte(line), &items); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 第 %d 行解析失败: %v\n", lineNum, err)
			os.Exit(1)
		}
		transactions = append(transactions, items)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 读取文件失败: %v\n", err)
		os.Exit(1)
	}

	req := common.ImportRequest{Transactions: transactions}
	respBody, err := httpPost("/import", req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 发送请求失败: %v\n", err)
		os.Exit(1)
	}

	var resp common.ImportResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 解析响应失败: %v\n", err)
		fmt.Fprintf(os.Stderr, "响应内容: %s\n", string(respBody))
		os.Exit(1)
	}

	if !resp.Success {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		fmt.Fprintf(os.Stderr, "错误: %s\n", errResp.Error)
		os.Exit(1)
	}

	fmt.Println(resp.Message)
}

func handleWeight(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "错误: 请指定权重JSON")
		fmt.Fprintln(os.Stderr, "用法: weight '{\"A\":0.8, \"B\":0.5}'")
		os.Exit(1)
	}

	jsonStr := args[0]
	weights := make(map[string]float64)
	if err := json.Unmarshal([]byte(jsonStr), &weights); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 解析权重JSON失败: %v\n", err)
		os.Exit(1)
	}

	req := common.SetWeightsRequest{Weights: weights}
	respBody, err := httpPost("/weight", req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 发送请求失败: %v\n", err)
		os.Exit(1)
	}

	var resp common.SetWeightsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 解析响应失败: %v\n", err)
		fmt.Fprintf(os.Stderr, "响应内容: %s\n", string(respBody))
		os.Exit(1)
	}

	if !resp.Success {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		fmt.Fprintf(os.Stderr, "错误: %s\n", errResp.Error)
		os.Exit(1)
	}

	fmt.Println(resp.Message)
}

func handleMine(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "错误: 请指定 minSup 和 minConf")
		fmt.Fprintln(os.Stderr, "用法: mine <minSup> <minConf>")
		fmt.Fprintln(os.Stderr, "示例: mine 0.1 0.5")
		os.Exit(1)
	}

	minSup, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: minSup 必须是数字: %v\n", err)
		os.Exit(1)
	}

	minConf, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: minConf 必须是数字: %v\n", err)
		os.Exit(1)
	}

	req := common.MineRequest{MinSup: minSup, MinConf: minConf}
	respBody, err := httpPost("/mine", req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 发送请求失败: %v\n", err)
		os.Exit(1)
	}

	var resp common.MineResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		if errResp.Error != "" {
			fmt.Fprintf(os.Stderr, "错误: %s\n", errResp.Error)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "错误: 解析响应失败: %v\n", err)
		fmt.Fprintf(os.Stderr, "响应内容: %s\n", string(respBody))
		os.Exit(1)
	}

	fmt.Printf("挖掘完成！\n")
	fmt.Printf("  频繁项集数量: %d\n", resp.FrequentCount)
	fmt.Printf("  关联规则数量: %d\n", resp.RulesCount)
	fmt.Printf("  扫描次数: %d\n", resp.Stats.ScansCount)
	fmt.Printf("  候选集总数: %d\n", resp.Stats.CandidatesCount)
	fmt.Println()

	if len(resp.TopRules) > 0 {
		fmt.Println("前20条置信度最高的关联规则:")
		for i, rule := range resp.TopRules {
			fmt.Printf("  %d. %s\n", i+1, formatRule(rule.Antecedent, rule.Consequent, rule.Confidence, rule.WeightedSupport))
		}
	} else {
		fmt.Println("未找到满足条件的关联规则。")
	}
}

func handleRules(args []string) {
	respBody, err := httpGet("/rules")
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 发送请求失败: %v\n", err)
		os.Exit(1)
	}

	var resp common.RulesResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		if errResp.Error != "" {
			fmt.Fprintf(os.Stderr, "错误: %s\n", errResp.Error)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "错误: 解析响应失败: %v\n", err)
		fmt.Fprintf(os.Stderr, "响应内容: %s\n", string(respBody))
		os.Exit(1)
	}

	if len(resp.Rules) == 0 {
		fmt.Println("未找到关联规则。")
		return
	}

	fmt.Printf("共找到 %d 条关联规则（按置信度降序）:\n", len(resp.Rules))
	for i, rule := range resp.Rules {
		fmt.Printf("  %d. %s\n", i+1, formatRule(rule.Antecedent, rule.Consequent, rule.Confidence, rule.WeightedSupport))
	}
}

func handleItems(args []string) {
	respBody, err := httpGet("/items")
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 发送请求失败: %v\n", err)
		os.Exit(1)
	}

	var resp common.FrequentItemSetResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		if errResp.Error != "" {
			fmt.Fprintf(os.Stderr, "错误: %s\n", errResp.Error)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "错误: 解析响应失败: %v\n", err)
		fmt.Fprintf(os.Stderr, "响应内容: %s\n", string(respBody))
		os.Exit(1)
	}

	if len(resp.FrequentItemSets) == 0 {
		fmt.Println("未找到频繁项集。")
		return
	}

	fmt.Printf("共找到 %d 个频繁项集:\n", len(resp.FrequentItemSets))
	for i, fis := range resp.FrequentItemSets {
		fmt.Printf("  %d. {%s} (加权支持度: %.4f, 出现次数: %d)\n",
			i+1, strings.Join(fis.Items, ", "), fis.WeightedSupport, fis.RawCount)
	}
}
