package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"recycling/api"
	"strconv"
)

var serverURL string

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:9016", "服务端地址")
	flag.Parse()

	if len(flag.Args()) == 0 {
		printHelp()
		return
	}

	command := flag.Args()[0]
	switch command {
	case "categories":
		listCategories()
	case "update-price":
		updatePrice()
	case "records":
		listRecords()
	case "create-record":
		createRecord()
	case "export-details":
		exportDetails()
	case "export-summary":
		exportSummary()
	case "help":
		printHelp()
	default:
		fmt.Printf("未知命令: %s\n", command)
		printHelp()
	}
}

func printHelp() {
	fmt.Println("废品回收管理平台 - 命令行客户端")
	fmt.Println("")
	fmt.Println("用法: client [options] <command> [arguments]")
	fmt.Println("")
	fmt.Println("选项:")
	fmt.Println("  -server <url>    服务端地址，默认 http://localhost:9016")
	fmt.Println("")
	fmt.Println("命令:")
	fmt.Println("  categories                          列出所有废品分类")
	fmt.Println("  update-price <id> <price>           更新分类单价")
	fmt.Println("  records                             列出所有回收记录")
	fmt.Println("  create-record <customer_id> <customer_name> <customer_type> <settlement> <category_id> <weight> ...")
	fmt.Println("                                      创建回收记录")
	fmt.Println("  export-details <start_date> <end_date> [category_id]")
	fmt.Println("                                      导出明细CSV")
	fmt.Println("  export-summary <year> <month>       导出月度汇总CSV")
	fmt.Println("  help                                显示帮助信息")
	fmt.Println("")
	fmt.Println("示例:")
	fmt.Println("  client categories")
	fmt.Println("  client update-price paper-box 1.50")
	fmt.Println("  client records")
	fmt.Println("  client create-record C001 张三 resident immediate paper-box 10.5 plastic-pet 5.2")
	fmt.Println("  client export-details 2024-01-01 2024-01-31")
	fmt.Println("  client export-details 2024-01-01 2024-01-31 paper-box")
	fmt.Println("  client export-summary 2024 1")
}

func listCategories() {
	resp, err := http.Get(serverURL + "/api/categories")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("请求失败: %s\n", string(body))
		return
	}

	var categories []api.Category
	json.NewDecoder(resp.Body).Decode(&categories)

	fmt.Println("废品分类列表:")
	fmt.Println("----------------------------------------")
	for _, c := range categories {
		fmt.Printf("ID: %s, 名称: %s, 父类: %s, 单价: %.2f 元/kg\n",
			c.ID, c.Name, c.ParentID, c.Price)
	}
}

func updatePrice() {
	args := flag.Args()[1:]
	if len(args) < 2 {
		fmt.Println("用法: update-price <id> <price>")
		return
	}

	id := args[0]
	price, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		fmt.Printf("无效的价格: %s\n", args[1])
		return
	}

	req := api.UpdatePriceRequest{
		CategoryID: id,
		NewPrice:   price,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/categories/update", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("请求失败: %s\n", string(errBody))
		return
	}

	fmt.Println("价格更新成功")
}

func listRecords() {
	resp, err := http.Get(serverURL + "/api/records")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("请求失败: %s\n", string(body))
		return
	}

	var records []api.Record
	json.NewDecoder(resp.Body).Decode(&records)

	fmt.Println("回收记录列表:")
	fmt.Println("----------------------------------------")
	for _, r := range records {
		fmt.Printf("ID: %s\n", r.ID)
		fmt.Printf("  时间: %s\n", r.DateTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("  客户: %s (%s, %s)\n", r.CustomerName, r.CustomerID, r.CustomerType)
		fmt.Printf("  结算方式: %s\n", r.Settlement)
		fmt.Printf("  明细:\n")
		for _, item := range r.Items {
			fmt.Printf("    %s: %.2f kg * %.2f 元/kg = %.2f 元\n",
				item.CategoryName, item.Weight, item.Price, item.Amount)
		}
		fmt.Printf("  总金额: %.2f 元\n", r.TotalAmount)
		fmt.Println("----------------------------------------")
	}
}

func createRecord() {
	args := flag.Args()[1:]
	if len(args) < 6 {
		fmt.Println("用法: create-record <customer_id> <customer_name> <customer_type> <settlement> <category_id> <weight> [category_id weight ...]")
		return
	}

	customerID := args[0]
	customerName := args[1]
	customerType := args[2]
	settlement := args[3]

	items := make([]api.RecordItem, 0)
	for i := 4; i < len(args); i += 2 {
		if i+1 >= len(args) {
			break
		}
		categoryID := args[i]
		weight, err := strconv.ParseFloat(args[i+1], 64)
		if err != nil {
			fmt.Printf("无效的重量: %s\n", args[i+1])
			return
		}
		items = append(items, api.RecordItem{
			CategoryID: categoryID,
			Weight:     weight,
		})
	}

	req := api.CreateRecordRequest{
		CustomerID:   customerID,
		CustomerName: customerName,
		CustomerType: customerType,
		Settlement:   settlement,
		Items:        items,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/records/create", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("请求失败: %s\n", string(errBody))
		return
	}

	var record api.Record
	json.NewDecoder(resp.Body).Decode(&record)
	fmt.Printf("记录创建成功，ID: %s，总金额: %.2f 元\n", record.ID, record.TotalAmount)
}

func exportDetails() {
	args := flag.Args()[1:]
	if len(args) < 2 {
		fmt.Println("用法: export-details <start_date> <end_date> [category_id]")
		return
	}

	startDate := args[0]
	endDate := args[1]
	categoryID := ""
	if len(args) > 2 {
		categoryID = args[2]
	}

	url := fmt.Sprintf("%s/api/export/details?start_date=%s&end_date=%s",
		serverURL, startDate, endDate)
	if categoryID != "" {
		url += "&category_id=" + categoryID
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("请求失败: %s\n", string(body))
		return
	}

	filename := fmt.Sprintf("records_%s_%s.csv", startDate, endDate)
	body, _ := io.ReadAll(resp.Body)
	os.WriteFile(filename, body, 0644)
	fmt.Printf("导出成功，保存到: %s\n", filename)
}

func exportSummary() {
	args := flag.Args()[1:]
	if len(args) < 2 {
		fmt.Println("用法: export-summary <year> <month>")
		return
	}

	year, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("无效的年份: %s\n", args[0])
		return
	}

	month, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("无效的月份: %s\n", args[1])
		return
	}

	url := fmt.Sprintf("%s/api/export/summary?year=%d&month=%d", serverURL, year, month)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("请求失败: %s\n", string(body))
		return
	}

	filename := fmt.Sprintf("summary_%d_%02d.csv", year, month)
	body, _ := io.ReadAll(resp.Body)
	os.WriteFile(filename, body, 0644)
	fmt.Printf("导出成功，保存到: %s\n", filename)
}
