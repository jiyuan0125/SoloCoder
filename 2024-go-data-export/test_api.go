//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const baseURL = "http://localhost:9800/api"

func main() {
	client := &http.Client{Timeout: 30 * time.Second}

	fmt.Println("=== 1. 测试获取表列表 ===")
	resp, err := client.Get(baseURL + "/tables")
	checkError(err)
	printResponse(resp)

	fmt.Println("\n=== 2. 测试获取 users 表结构 ===")
	resp, err = client.Get(baseURL + "/tables/users")
	checkError(err)
	printResponse(resp)

	fmt.Println("\n=== 3. 测试不存在的表 (应返回 404) ===")
	resp, err = client.Get(baseURL + "/tables/nonexistent_table")
	checkError(err)
	fmt.Printf("Status: %d\n", resp.StatusCode)

	fmt.Println("\n=== 4. 测试提交 CSV 导出任务 ===")
	reqBody := map[string]interface{}{
		"user_id":    "test_user_001",
		"table_name": "users",
		"fields":     []string{"id", "name", "email", "phone", "age"},
		"filter":     map[string]string{"conditions": "age > 20"},
		"format":     "csv",
	}
	body, _ := json.Marshal(reqBody)
	resp, err = client.Post(baseURL+"/export", "application/json", bytes.NewReader(body))
	checkError(err)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("Response: %v\n", result)

	taskID := uint(result["task_id"].(float64))
	fmt.Printf("Task ID: %d\n", taskID)

	fmt.Println("\n=== 5. 测试重复提交同一用户 (应返回 409) ===")
	resp, err = client.Post(baseURL+"/export", "application/json", bytes.NewReader(body))
	checkError(err)
	fmt.Printf("Status: %d (expect 409)\n", resp.StatusCode)
	resp.Body.Close()

	time.Sleep(2 * time.Second)

	fmt.Println("\n=== 6. 测试查询任务进度 ===")
	resp, err = client.Get(fmt.Sprintf("%s/export/%d/progress", baseURL, taskID))
	checkError(err)
	printResponse(resp)

	fmt.Println("\n=== 7. 等待导出完成... ===")
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		resp, err = client.Get(fmt.Sprintf("%s/export/%d/progress", baseURL, taskID))
		checkError(err)
		var progress map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&progress)
		resp.Body.Close()
		fmt.Printf("  Progress: status=%v, progress=%v\n", progress["status"], progress["progress"])
		if progress["status"] == "completed" {
			break
		}
	}

	fmt.Println("\n=== 8. 测试下载导出文件 ===")
	resp, err = client.Get(fmt.Sprintf("%s/export/%d/download", baseURL, taskID))
	checkError(err)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	if resp.StatusCode == 200 {
		data, _ := io.ReadAll(resp.Body)
		fmt.Println("File content (first 500 chars):")
		if len(data) > 500 {
			fmt.Println(string(data[:500]))
		} else {
			fmt.Println(string(data))
		}
	}
	resp.Body.Close()

	fmt.Println("\n=== 9. 测试不支持的格式 (应返回 400) ===")
	reqBody2 := map[string]interface{}{
		"user_id":    "test_user_002",
		"table_name": "users",
		"fields":     []string{"name", "email"},
		"format":     "pdf",
	}
	body2, _ := json.Marshal(reqBody2)
	resp, err = client.Post(baseURL+"/export", "application/json", bytes.NewReader(body2))
	checkError(err)
	fmt.Printf("Status: %d (expect 400)\n", resp.StatusCode)

	fmt.Println("\n=== 10. 测试 Excel 导出 ===")
	reqBody3 := map[string]interface{}{
		"user_id":    "test_user_003",
		"table_name": "orders",
		"fields":     []string{"id", "order_no", "amount", "status"},
		"format":     "excel",
	}
	body3, _ := json.Marshal(reqBody3)
	resp, err = client.Post(baseURL+"/export", "application/json", bytes.NewReader(body3))
	checkError(err)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("Excel Task: %v\n", result)

	fmt.Println("\n=== 全部测试完成 ===")
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func printResponse(resp *http.Response) {
	defer resp.Body.Close()
	fmt.Printf("Status: %d\n", resp.StatusCode)
	data, _ := io.ReadAll(resp.Body)
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, data, "", "  ")
	fmt.Println(prettyJSON.String())
}
