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
)

const baseURL = "http://localhost:8300/api"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "register-enterprise":
		registerEnterprise(args)
	case "schedule-exam":
		scheduleExam(args)
	case "record-result":
		recordResult(args)
	case "export-report":
		exportReport(args)
	default:
		fmt.Printf("未知命令: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`职业病防治管理系统 CLI

用法:
  client <command> [arguments]

命令:
  register-enterprise    登记企业
  schedule-exam          安排体检
  record-result          录入体检结果
  export-report          导出报告

示例:
  client register-enterprise -name "XX公司" -uscc 91310000MA1K31XW1M -industry 制造业 -region 上海市
  client schedule-exam -worker 1 -enterprise 1 -type 在岗期间体检 -date 2026-05-12
  client record-result -exam 1 -items CHST_XRAY:正常,PFT:正常
  client export-report -id 1 -output report.txt
`)
}

func registerEnterprise(args []string) {
	fs := flag.NewFlagSet("register-enterprise", flag.ExitOnError)
	name := fs.String("name", "", "企业名称")
	uscc := fs.String("uscc", "", "统一社会信用代码")
	industry := fs.String("industry", "制造业", "行业")
	region := fs.String("region", "", "地区")
	contact := fs.String("contact", "", "联系人")
	phone := fs.String("phone", "", "联系电话")
	address := fs.String("address", "", "地址")
	fs.Parse(args)

	if *name == "" || *uscc == "" {
		fmt.Println("错误: -name 和 -uscc 是必需的")
		os.Exit(1)
	}

	data := map[string]interface{}{
		"name":                *name,
		"unified_social_code": *uscc,
		"industry":            *industry,
		"region":              *region,
		"contact_person":      *contact,
		"contact_phone":       *phone,
		"address":             *address,
	}

	resp, err := httpPost("/enterprises", data)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("企业登记成功:\n%s\n", resp)
}

func scheduleExam(args []string) {
	fs := flag.NewFlagSet("schedule-exam", flag.ExitOnError)
	workerID := fs.Uint("worker", 0, "劳动者ID")
	entID := fs.Uint("enterprise", 0, "企业ID")
	examType := fs.String("type", "在岗期间体检", "体检类型: 上岗前体检/在岗期间体检/离岗时体检")
	date := fs.String("date", "", "安排日期 (YYYY-MM-DD)")
	fs.Parse(args)

	if *workerID == 0 || *entID == 0 {
		fmt.Println("错误: -worker 和 -enterprise 是必需的")
		os.Exit(1)
	}

	data := map[string]interface{}{
		"worker_id":      *workerID,
		"enterprise_id":  *entID,
		"exam_type":      *examType,
		"scheduled_date": *date,
	}

	resp, err := httpPost("/examinations", data)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("体检安排成功:\n%s\n", resp)
}

func recordResult(args []string) {
	fs := flag.NewFlagSet("record-result", flag.ExitOnError)
	examID := fs.Uint("exam", 0, "体检记录ID")
	items := fs.String("items", "", "体检项目结果，格式: CODE:结果,CODE2:结果2")
	abnormal := fs.String("abnormal", "", "异常项目，格式: CODE,CODE2")
	suspected := fs.String("suspected", "", "疑似职业病项目，格式: CODE,CODE2")
	date := fs.String("date", "", "体检日期 (YYYY-MM-DD)")
	fs.Parse(args)

	if *examID == 0 {
		fmt.Println("错误: -exam 是必需的")
		os.Exit(1)
	}

	examItems := make([]map[string]interface{}, 0)
	itemResults := parseItems(*items)
	abnormalSet := parseSet(*abnormal)
	suspectedSet := parseSet(*suspected)

	for code, result := range itemResults {
		item := map[string]interface{}{
			"item_code":   code,
			"result":      result,
			"is_abnormal": abnormalSet[code],
			"is_suspected": suspectedSet[code],
		}
		examItems = append(examItems, item)
	}

	data := map[string]interface{}{
		"exam_items": examItems,
		"exam_date":  *date,
	}

	resp, err := httpPut(fmt.Sprintf("/examinations/%d/results", *examID), data)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("结果录入成功:\n%s\n", resp)
}

func parseItems(itemsStr string) map[string]string {
	result := make(map[string]string)
	if itemsStr == "" {
		return result
	}
	for _, part := range splitIgnoreEmpty(itemsStr, ",") {
		kv := splitIgnoreEmpty(part, ":")
		if len(kv) >= 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}

func parseSet(str string) map[string]bool {
	result := make(map[string]bool)
	if str == "" {
		return result
	}
	for _, item := range splitIgnoreEmpty(str, ",") {
		result[item] = true
	}
	return result
}

func splitIgnoreEmpty(s, sep string) []string {
	var result []string
	for _, part := range splitAll(s, sep) {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func splitAll(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i+len(sep) <= len(s); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

func exportReport(args []string) {
	fs := flag.NewFlagSet("export-report", flag.ExitOnError)
	reportID := fs.Uint("id", 0, "报告ID")
	output := fs.String("output", "", "输出文件路径")
	fs.Parse(args)

	if *reportID == 0 {
		fmt.Println("错误: -id 是必需的")
		os.Exit(1)
	}

	resp, err := http.Get(baseURL + fmt.Sprintf("/reports/%d/export", *reportID))
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("服务器返回错误 (HTTP %d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	content, _ := io.ReadAll(resp.Body)
	if *output != "" {
		err = os.WriteFile(*output, content, 0644)
		if err != nil {
			fmt.Printf("写入文件失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("报告已导出到: %s\n", *output)
	} else {
		fmt.Println(string(content))
	}
}

func httpPost(path string, data interface{}) (string, error) {
	body, _ := json.Marshal(data)
	resp, err := http.Post(baseURL+path, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return string(respBody), nil
}

func httpPut(path string, data interface{}) (string, error) {
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest(http.MethodPut, baseURL+path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return string(respBody), nil
}

func _unused() {
	_ = strconv.Itoa(0)
}
