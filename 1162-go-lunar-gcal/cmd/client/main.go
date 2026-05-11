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
	"strings"

	"go-lunar-gcal/pkg/api"
)

var serverURL = "http://localhost:8100"

func getServerURL() string {
	if url := os.Getenv("SERVER_URL"); url != "" {
		return url
	}
	return serverURL
}

func httpPost(endpoint string, body interface{}, result interface{}) error {
	url := getServerURL() + endpoint
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("响应解析失败: %v", err)
	}

	return nil
}

func httpGet(endpoint string, result interface{}) error {
	url := getServerURL() + endpoint
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, result)
	if err != nil {
		return fmt.Errorf("响应解析失败: %v", err)
	}

	return nil
}

func solarToLunar(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: solar-to-lunar <日期>")
		fmt.Println("示例: solar-to-lunar 2024-02-10")
		os.Exit(1)
	}

	req := api.SolarToLunarRequest{
		Date: args[0],
	}

	var resp api.SolarToLunarResponse
	if err := httpPost("/api/solar-to-lunar", req, &resp); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("错误: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Printf("公历: %s\n", args[0])
	fmt.Printf("农历: %s\n", resp.Data.FullText)
	fmt.Printf("\n详细信息:\n")
	fmt.Printf("  农历年份: %d年\n", resp.Data.Year)
	fmt.Printf("  天干地支: %s\n", resp.Data.Ganzhi)
	fmt.Printf("  生肖: %s\n", resp.Data.Shengxiao)
	fmt.Printf("  月份: %s (%d月%s)\n", resp.Data.MonthName, resp.Data.Month, formatLeap(resp.Data.IsLeapMonth))
	fmt.Printf("  日期: %s (%d日)\n", resp.Data.DayName, resp.Data.Day)
}

func formatLeap(isLeap bool) string {
	if isLeap {
		return "，闰月"
	}
	return ""
}

func lunarToSolar(args []string) {
	if len(args) < 3 {
		fmt.Println("用法: lunar-to-solar <年份> <月份> <日期>")
		fmt.Println("示例: lunar-to-solar 2024 正月 初一")
		fmt.Println("      lunar-to-solar 2024 闰二月 十五")
		fmt.Println("      lunar-to-solar 2024 2月 15")
		os.Exit(1)
	}

	yearStr := args[0]
	monthStr := args[1]
	dayStr := args[2]

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		fmt.Printf("年份格式错误: %s\n", yearStr)
		os.Exit(1)
	}

	req := api.LunarToSolarRequest{
		Year:     year,
		MonthStr: monthStr,
		DayStr:   dayStr,
	}

	var resp api.LunarToSolarResponse
	if err := httpPost("/api/lunar-to-solar", req, &resp); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("错误: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Printf("农历: %s年 %s %s\n", yearStr, monthStr, dayStr)
	fmt.Printf("公历: %s\n", resp.Data.FullText)
	fmt.Printf("\n详细信息:\n")
	fmt.Printf("  公历日期: %s\n", resp.Data.DateStr)
	fmt.Printf("  星期: %s\n", resp.Data.Weekday)
}

func jieqi(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: jieqi <年份>")
		fmt.Println("示例: jieqi 2024")
		os.Exit(1)
	}

	yearStr := args[0]
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		fmt.Printf("年份格式错误: %s\n", yearStr)
		os.Exit(1)
	}

	var resp api.JieqiResponse
	if err := httpGet(fmt.Sprintf("/api/jieqi?year=%d", year), &resp); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("错误: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Printf("%d 年二十四节气:\n", year)
	fmt.Println(strings.Repeat("-", 40))
	for i, item := range resp.Data {
		fmt.Printf("%2d. %-4s %s\n", i+1, item.Name, item.DateTime)
	}
}

func printUsage() {
	fmt.Println("农历公历互转工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  lunar-to-solar <年份> <月份> <日期>  - 农历转公历")
	fmt.Println("  solar-to-lunar <日期>               - 公历转农历")
	fmt.Println("  jieqi <年份>                        - 查询二十四节气")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  solar-to-lunar 2024-02-10")
	fmt.Println("  lunar-to-solar 2024 正月 初一")
	fmt.Println("  lunar-to-solar 2024 闰二月 十五")
	fmt.Println("  lunar-to-solar 2024 2月 15")
	fmt.Println("  jieqi 2024")
	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  SERVER_URL  - 服务端地址，默认 http://localhost:8100")
}

func main() {
	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := strings.ToLower(args[0])
	rest := args[1:]

	switch command {
	case "lunar-to-solar":
		lunarToSolar(rest)
	case "solar-to-lunar":
		solarToLunar(rest)
	case "jieqi":
		jieqi(rest)
	default:
		fmt.Printf("未知命令: %s\n", command)
		fmt.Println()
		printUsage()
		os.Exit(1)
	}
}
