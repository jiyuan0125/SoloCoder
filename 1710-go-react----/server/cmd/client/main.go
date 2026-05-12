package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const baseURL = "http://localhost:8300/api"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "create-protocol":
		createProtocol(args)
	case "enroll-subject":
		enrollSubject(args)
	case "record-visit":
		recordVisit(args)
	case "export-data":
		exportData(args)
	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("用法: trial-cli <命令> [参数]")
	fmt.Println("")
	fmt.Println("命令:")
	fmt.Println("  create-protocol   创建试验方案")
	fmt.Println("  enroll-subject    入组受试者")
	fmt.Println("  record-visit      录入访视数据")
	fmt.Println("  export-data       导出数据")
	fmt.Println("")
	fmt.Println("示例:")
	fmt.Println("  trial-cli create-protocol CT-2026-001 新药X 高血压 II期 100 2026-01-01 2027-01-01 纳入标准 排除标准 2:1")
	fmt.Println("  trial-cli enroll-subject <协议ID> <中心ID> S001 张XX 男 1990-01-01 筛选中")
	fmt.Println("  trial-cli record-visit <协议ID> <受试者ID> <访视ID> 2026-01-15")
	fmt.Println("  trial-cli export-data <协议ID>")
}

func createProtocol(args []string) {
	if len(args) < 10 {
		fmt.Println("用法: trial-cli create-protocol <方案编号> <药物名称> <适应症> <试验阶段> <计划入组人数> <开始日期> <结束日期> <纳入标准> <排除标准> <分组比例>")
		return
	}

	planned, _ := strconv.Atoi(args[4])
	startDate, _ := time.Parse("2006-01-02", args[5])
	endDate, _ := time.Parse("2006-01-02", args[6])

	req := map[string]interface{}{
		"protocol_number":    args[0],
		"drug_name":          args[1],
		"indication":         args[2],
		"trial_phase":        args[3],
		"planned_enrollment": planned,
		"start_date":         startDate,
		"end_date":           endDate,
		"inclusion_criteria": args[7],
		"exclusion_criteria": args[8],
		"group_ratio":        args[9],
		"status":             "筹备中",
		"sites":              []interface{}{},
		"visits":             []interface{}{},
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(baseURL+"/protocols", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	result, _ := io.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n响应: %s\n", resp.StatusCode, string(result))
}

func enrollSubject(args []string) {
	if len(args) < 7 {
		fmt.Println("用法: trial-cli enroll-subject <方案ID> <中心ID> <筛选编号> <姓名缩写> <性别> <出生日期> <初始状态>")
		return
	}

	birthDate, _ := time.Parse("2006-01-02", args[5])

	req := map[string]interface{}{
		"protocol_id":      args[0],
		"site_id":          args[1],
		"screening_number": args[2],
		"name_initials":    args[3],
		"gender":           args[4],
		"birth_date":       birthDate,
		"initial_status":   args[6],
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(baseURL+"/subjects", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	result, _ := io.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n响应: %s\n", resp.StatusCode, string(result))
}

func recordVisit(args []string) {
	if len(args) < 4 {
		fmt.Println("用法: trial-cli record-visit <方案ID> <受试者ID> <访视ID> <实际日期> [生命体征] [实验室检查] [其他数据]")
		return
	}

	actualDate, _ := time.Parse("2006-01-02", args[3])

	req := map[string]interface{}{
		"protocol_id": args[0],
		"subject_id":  args[1],
		"visit_id":    args[2],
		"actual_date": actualDate,
	}

	if len(args) > 4 {
		req["vital_signs"] = args[4]
	}
	if len(args) > 5 {
		req["lab_tests"] = args[5]
	}
	if len(args) > 6 {
		req["other_data"] = args[6]
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(baseURL+"/visits/records", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	result, _ := io.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n响应: %s\n", resp.StatusCode, string(result))
}

func exportData(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: trial-cli export-data <方案ID> [中心ID] [类型(subjects/ae/sae/visits)]")
		return
	}

	protocolID := args[0]
	siteID := ""
	exportType := "subjects"

	if len(args) > 1 {
		siteID = args[1]
	}
	if len(args) > 2 {
		exportType = args[2]
	}

	url := baseURL + "/export/data?protocol_id=" + protocolID
	if siteID != "" {
		url += "&site_id=" + siteID
	}
	if exportType != "subjects" {
		url += "&type=" + exportType
	}

	if exportType == "visits" {
		url = baseURL + "/export/visits?protocol_id=" + protocolID
		if siteID != "" {
			url += "&site_id=" + siteID
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	result, _ := io.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	if resp.StatusCode == 200 {
		filename := fmt.Sprintf("export_%s_%s.csv", exportType, time.Now().Format("20060102"))
		os.WriteFile(filename, result, 0644)
		fmt.Printf("已保存到: %s\n", filename)
	} else {
		fmt.Println("响应:", string(result))
	}
}
