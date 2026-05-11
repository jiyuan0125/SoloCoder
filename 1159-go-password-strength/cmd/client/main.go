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

	"github.com/passwordstrength/common"
)

const defaultServerURL = "http://localhost:8603"

func maskPassword(pwd string) string {
	if len(pwd) == 0 {
		return ""
	}
	return strings.Repeat("*", len(pwd))
}

func evaluatePassword(serverURL, password string) (*common.EvaluateResponse, error) {
	reqBody := common.EvaluateRequest{Password: password}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}
	resp, err := http.Post(serverURL+"/api/evaluate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}
	var result common.EvaluateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}

func printSingleResult(password string, resp *common.EvaluateResponse) {
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("密码: %s (长度=%d)\n", maskPassword(password), len(password))
	fmt.Printf("总分: %.1f/100\n", resp.TotalScore)
	fmt.Printf("强度等级: %s\n", resp.Level)
	fmt.Printf("是否常见弱密码: %v\n", resp.IsCommon)
	fmt.Printf("理论熵: %.1f bit\n", resp.Entropy.Theoretical)
	fmt.Printf("估计熵: %.1f bit\n", resp.Entropy.Estimated)
	if len(resp.Dimensions) > 0 {
		fmt.Println("\n各维度得分:")
		for _, d := range resp.Dimensions {
			fmt.Printf("  - %s: %.1f/%.1f\n", d.Name, d.Score, d.Max)
		}
	}
	if len(resp.Suggestions) > 0 {
		fmt.Println("\n改进建议:")
		for _, s := range resp.Suggestions {
			fmt.Printf("  - %s\n", s)
		}
	}
	fmt.Println("=" + strings.Repeat("=", 50))
}

func processBatch(serverURL, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	var passwords []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			passwords = append(passwords, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	if len(passwords) == 0 {
		return fmt.Errorf("no passwords found in file")
	}
	summary := common.BatchSummary{TotalCount: len(passwords)}
	var totalScore float64
	var strongestScore, weakestScore float64 = -1, 101
	var strongestMasked, weakestMasked string
	for i, pwd := range passwords {
		resp, err := evaluatePassword(serverURL, pwd)
		if err != nil {
			return fmt.Errorf("failed to evaluate password at line %d: %w", i+1, err)
		}
		masked := maskPassword(pwd)
		totalScore += resp.TotalScore
		switch resp.Level {
		case common.LevelWeak:
			summary.WeakCount++
		case common.LevelMedium:
			summary.MediumCount++
		case common.LevelStrong:
			summary.StrongCount++
		case common.LevelVeryStrong:
			summary.VeryStrongCount++
		}
		if strongestScore < 0 || resp.TotalScore > strongestScore {
			strongestScore = resp.TotalScore
			strongestMasked = masked
		}
		if resp.TotalScore < weakestScore {
			weakestScore = resp.TotalScore
			weakestMasked = masked
		}
	}
	summary.AverageScore = totalScore / float64(len(passwords))
	summary.StrongestExample = strongestMasked
	summary.WeakestExample = weakestMasked
	printBatchSummary(&summary, strongestScore, weakestScore)
	return nil
}

func printBatchSummary(s *common.BatchSummary, strongest, weakest float64) {
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Println("批量评估汇总")
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Printf("总密码数: %d\n", s.TotalCount)
	fmt.Printf("平均分数: %.2f\n", s.AverageScore)
	fmt.Println()
	fmt.Println("强度分布:")
	fmt.Printf("  弱:     %d 个\n", s.WeakCount)
	fmt.Printf("  中:     %d 个\n", s.MediumCount)
	fmt.Printf("  强:     %d 个\n", s.StrongCount)
	fmt.Printf("  很强:   %d 个\n", s.VeryStrongCount)
	fmt.Println()
	fmt.Printf("最强密码示例: %s (分数: %.1f)\n", s.StrongestExample, strongest)
	fmt.Printf("最弱密码示例: %s (分数: %.1f)\n", s.WeakestExample, weakest)
	fmt.Println("=" + strings.Repeat("=", 60))
}

func printUsage() {
	fmt.Println("Password Strength Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client -p <password>     评估单个密码")
	fmt.Println("  client -f <file>         批量评估文件中的密码（每行一个）")
	fmt.Println("  client -s <url>          指定服务端URL（默认 http://localhost:8603）")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client -p \"MyPassword123!\"")
	fmt.Println("  client -f passwords.txt -s http://myserver:8603")
}

func main() {
	var password string
	var filePath string
	var serverURL string
	flag.StringVar(&password, "p", "", "Single password to evaluate")
	flag.StringVar(&filePath, "f", "", "File containing passwords (one per line)")
	flag.StringVar(&serverURL, "s", defaultServerURL, "Server URL")
	flag.Usage = printUsage
	flag.Parse()
	if password == "" && filePath == "" {
		printUsage()
		os.Exit(1)
	}
	if password != "" {
		resp, err := evaluatePassword(serverURL, password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		printSingleResult(password, resp)
	}
	if filePath != "" {
		if err := processBatch(serverURL, filePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
