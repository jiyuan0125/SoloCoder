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

	"regex-generator/pkg/api"
)

type exampleFlag []string

func (e *exampleFlag) String() string {
	return strings.Join(*e, ", ")
}

func (e *exampleFlag) Set(value string) error {
	*e = append(*e, value)
	return nil
}

func readExamplesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var examples []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			examples = append(examples, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return examples, nil
}

func makeRequest(url string, body interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

func handleGenerateCommand() {
	var examples exampleFlag
	var fileFlag string
	var serverURL string
	var shortestMatch bool
	var allowOptional bool

	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	fs.Var(&examples, "examples", "example strings (can be used multiple times)")
	fs.StringVar(&fileFlag, "file", "", "file containing examples (one per line)")
	fs.StringVar(&serverURL, "server", "http://localhost:8601", "server URL")
	fs.BoolVar(&shortestMatch, "shortest", false, "generate shortest matching regex")
	fs.BoolVar(&allowOptional, "optional", false, "allow optional groups")

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "error parsing flags:", err)
		os.Exit(1)
	}

	var allExamples []string

	if fileFlag != "" {
		fileExamples, err := readExamplesFromFile(fileFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error reading file:", err)
			os.Exit(1)
		}
		allExamples = append(allExamples, fileExamples...)
	}

	if len(examples) > 0 {
		allExamples = append(allExamples, examples...)
	}

	if len(allExamples) == 0 {
		fmt.Fprintln(os.Stderr, "error: no examples provided. Use --examples or --file flag.")
		os.Exit(1)
	}

	seen := make(map[string]bool)
	uniqueExamples := make([]string, 0, len(allExamples))
	for _, ex := range allExamples {
		if !seen[ex] {
			seen[ex] = true
			uniqueExamples = append(uniqueExamples, ex)
		}
	}

	req := api.GenerateRequest{
		Examples:      uniqueExamples,
		ShortestMatch: shortestMatch,
		AllowOptional: allowOptional,
	}

	resp, err := makeRequest(serverURL+"/generate", req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error making request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading response:", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errResp); err == nil {
			fmt.Fprintln(os.Stderr, "server error:", errResp.Error)
		} else {
			fmt.Fprintln(os.Stderr, "server returned status:", resp.StatusCode)
			fmt.Fprintln(os.Stderr, "response body:", string(bodyBytes))
		}
		os.Exit(1)
	}

	var genResp api.GenerateResponse
	if err := json.Unmarshal(bodyBytes, &genResp); err != nil {
		fmt.Fprintln(os.Stderr, "error parsing response:", err)
		os.Exit(1)
	}

	fmt.Println("正则表达式:", genResp.Regex)
	fmt.Println("说明:", genResp.Explanation)
}

func handleValidateCommand() {
	var regex string
	var testStr string
	var serverURL string

	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	fs.StringVar(&regex, "regex", "", "regex pattern to validate")
	fs.StringVar(&testStr, "test", "", "test string to match against")
	fs.StringVar(&serverURL, "server", "http://localhost:8601", "server URL")

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "error parsing flags:", err)
		os.Exit(1)
	}

	if regex == "" || testStr == "" {
		fmt.Fprintln(os.Stderr, "error: --regex and --test flags are required")
		os.Exit(1)
	}

	req := api.ValidateRequest{
		Regex:   regex,
		TestStr: testStr,
	}

	resp, err := makeRequest(serverURL+"/validate", req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error making request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading response:", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errResp); err == nil {
			fmt.Fprintln(os.Stderr, "server error:", errResp.Error)
		} else {
			fmt.Fprintln(os.Stderr, "server returned status:", resp.StatusCode)
			fmt.Fprintln(os.Stderr, "response body:", string(bodyBytes))
		}
		os.Exit(1)
	}

	var valResp api.ValidateResponse
	if err := json.Unmarshal(bodyBytes, &valResp); err != nil {
		fmt.Fprintln(os.Stderr, "error parsing response:", err)
		os.Exit(1)
	}

	if valResp.Matched {
		fmt.Println("匹配成功: 是的，字符串与正则表达式匹配")
	} else {
		fmt.Println("匹配失败: 否，字符串与正则表达式不匹配")
	}
}

func printUsage() {
	fmt.Println("正则表达式生成器客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client generate --examples <string> [--examples <string>...]")
	fmt.Println("  client generate --file <filename>")
	fmt.Println("  client validate --regex <pattern> --test <string>")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  generate    从示例字符串生成正则表达式")
	fmt.Println("  validate    验证正则表达式是否匹配测试字符串")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  client generate --examples \"test123\" --examples \"test456\"")
	fmt.Println("  client generate --file examples.txt")
	fmt.Println("  client validate --regex \"test\\d+\" --test \"test123\"")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate":
		handleGenerateCommand()
	case "validate":
		handleValidateCommand()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintln(os.Stderr, "未知命令:", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}
