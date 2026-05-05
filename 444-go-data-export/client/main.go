package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"data-export/pkg/common"
)

var (
	serverAddr = "http://localhost:8080"
	currentUser = "user_001"
)

type APIClient struct {
	baseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{baseURL: baseURL}
}

func (c *APIClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *APIClient) ListTemplates() error {
	body, err := c.doRequest("GET", "/api/templates", nil)
	if err != nil {
		return err
	}

	var resp common.APIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if resp.Code != common.ErrCodeSuccess {
		return fmt.Errorf("[%d] %s", resp.Code, resp.Message)
	}

	var templates []common.Template
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &templates)

	fmt.Println("\n=== Templates List ===")
	for i, t := range templates {
		fmt.Printf("%d. ID: %s\n", i+1, t.ID)
		fmt.Printf("   Name: %s\n", t.Name)
		fmt.Printf("   Description: %s\n", t.Description)
		fmt.Printf("   DataSource: %s\n", t.DataSource)
		fmt.Printf("   Format: %s\n", t.OutputFormat)
		fmt.Printf("   Fields: %d\n", len(t.FieldMappings))
		fmt.Println()
	}

	return nil
}

func (c *APIClient) CreateTemplate() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter template name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Enter description (optional): ")
	desc, _ := reader.ReadString('\n')
	desc = strings.TrimSpace(desc)

	fmt.Println("\nSelect data source:")
	fmt.Println("1. user_behavior (用户行为)")
	fmt.Println("2. transactions (交易明细)")
	fmt.Println("3. feature_usage (功能使用)")
	fmt.Print("Enter choice: ")
	dsChoice, _ := reader.ReadString('\n')
	dsChoice = strings.TrimSpace(dsChoice)

	var dataSource string
	switch dsChoice {
	case "1":
		dataSource = "user_behavior"
	case "2":
		dataSource = "transactions"
	case "3":
		dataSource = "feature_usage"
	default:
		dataSource = "user_behavior"
	}

	fmt.Println("\nSelect output format:")
	fmt.Println("1. CSV")
	fmt.Println("2. JSON")
	fmt.Println("3. Excel")
	fmt.Print("Enter choice: ")
	fmtChoice, _ := reader.ReadString('\n')
	fmtChoice = strings.TrimSpace(fmtChoice)

	var format common.ExportFormat
	switch fmtChoice {
	case "1":
		format = common.FormatCSV
	case "2":
		format = common.FormatJSON
	case "3":
		format = common.FormatExcel
	default:
		format = common.FormatCSV
	}

	fieldMappings := getDefaultFieldMappings(dataSource)

	req := common.CreateTemplateRequest{
		Name:           name,
		Description:    desc,
		DataSource:     dataSource,
		QueryCondition: []common.QueryCondition{},
		OutputFormat:   format,
		FieldMappings:  fieldMappings,
	}

	body, err := c.doRequest("POST", "/api/templates", req)
	if err != nil {
		return err
	}

	var resp common.APIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if resp.Code != common.ErrCodeSuccess {
		return fmt.Errorf("[%d] %s", resp.Code, resp.Message)
	}

	var tpl common.Template
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &tpl)

	fmt.Printf("\nTemplate created successfully!\n")
	fmt.Printf("ID: %s\n", tpl.ID)
	fmt.Printf("Name: %s\n", tpl.Name)

	return nil
}

func getDefaultFieldMappings(dataSource string) []common.FieldMapping {
	switch dataSource {
	case "user_behavior":
		return []common.FieldMapping{
			{SourceField: "user_id", TargetField: "用户ID", FieldType: common.FieldTypeString},
			{SourceField: "user_name", TargetField: "用户名", FieldType: common.FieldTypeString},
			{SourceField: "phone", TargetField: "手机号", FieldType: common.FieldTypePhone, Sensitive: true},
			{SourceField: "idcard", TargetField: "身份证号", FieldType: common.FieldTypeIDCard, Sensitive: true},
			{SourceField: "action", TargetField: "行为", FieldType: common.FieldTypeString},
			{SourceField: "page", TargetField: "页面", FieldType: common.FieldTypeString},
			{SourceField: "timestamp", TargetField: "时间", FieldType: common.FieldTypeDate, FormatPattern: "2006-01-02 15:04:05"},
			{SourceField: "duration_ms", TargetField: "耗时(毫秒)", FieldType: common.FieldTypeInteger},
		}
	case "transactions":
		return []common.FieldMapping{
			{SourceField: "transaction_id", TargetField: "交易ID", FieldType: common.FieldTypeString},
			{SourceField: "user_id", TargetField: "用户ID", FieldType: common.FieldTypeString},
			{SourceField: "user_name", TargetField: "用户名", FieldType: common.FieldTypeString},
			{SourceField: "phone", TargetField: "手机号", FieldType: common.FieldTypePhone, Sensitive: true},
			{SourceField: "amount", TargetField: "金额", FieldType: common.FieldTypeFloat, FormatPattern: "%.2f"},
			{SourceField: "status", TargetField: "状态", FieldType: common.FieldTypeString},
			{SourceField: "payment_method", TargetField: "支付方式", FieldType: common.FieldTypeString},
			{SourceField: "created_at", TargetField: "创建时间", FieldType: common.FieldTypeDate, FormatPattern: "2006-01-02 15:04:05"},
		}
	case "feature_usage":
		return []common.FieldMapping{
			{SourceField: "feature_id", TargetField: "功能ID", FieldType: common.FieldTypeString},
			{SourceField: "feature_name", TargetField: "功能名称", FieldType: common.FieldTypeString},
			{SourceField: "user_id", TargetField: "用户ID", FieldType: common.FieldTypeString},
			{SourceField: "user_name", TargetField: "用户名", FieldType: common.FieldTypeString},
			{SourceField: "phone", TargetField: "手机号", FieldType: common.FieldTypePhone, Sensitive: true},
			{SourceField: "usage_count", TargetField: "使用次数", FieldType: common.FieldTypeInteger},
			{SourceField: "last_used_at", TargetField: "最后使用时间", FieldType: common.FieldTypeDate, FormatPattern: "2006-01-02 15:04:05"},
		}
	default:
		return []common.FieldMapping{}
	}
}

func (c *APIClient) DeleteTemplate() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter template ID to delete: ")
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	body, err := c.doRequest("DELETE", "/api/templates/"+id, nil)
	if err != nil {
		return err
	}

	var resp common.APIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if resp.Code != common.ErrCodeSuccess {
		return fmt.Errorf("[%d] %s", resp.Code, resp.Message)
	}

	fmt.Println("Template deleted successfully!")
	return nil
}

func (c *APIClient) CreateExportTask() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter template ID: ")
	tplID, _ := reader.ReadString('\n')
	tplID = strings.TrimSpace(tplID)

	fmt.Printf("Current user: %s\n", currentUser)
	fmt.Print("Use different user ID? (leave blank to use current): ")
	userID, _ := reader.ReadString('\n')
	userID = strings.TrimSpace(userID)
	if userID == "" {
		userID = currentUser
	}

	req := common.CreateTaskRequest{
		TemplateID: tplID,
		UserID:     userID,
		Params:     map[string]interface{}{},
	}

	body, err := c.doRequest("POST", "/api/tasks", req)
	if err != nil {
		return err
	}

	var resp common.APIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if resp.Code != common.ErrCodeSuccess {
		return fmt.Errorf("[%d] %s", resp.Code, resp.Message)
	}

	var result map[string]interface{}
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)

	fmt.Printf("\nExport task submitted!\n")
	fmt.Printf("Task ID: %v\n", result["task_id"])
	fmt.Printf("Status: %v\n", result["status"])

	return nil
}

func (c *APIClient) CheckTaskProgress() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter task ID: ")
	taskID, _ := reader.ReadString('\n')
	taskID = strings.TrimSpace(taskID)

	body, err := c.doRequest("GET", "/api/tasks/"+taskID+"/progress", nil)
	if err != nil {
		return err
	}

	var resp common.APIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if resp.Code != common.ErrCodeSuccess {
		return fmt.Errorf("[%d] %s", resp.Code, resp.Message)
	}

	var progress common.TaskProgressResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &progress)

	fmt.Println("\n=== Task Progress ===")
	fmt.Printf("Task ID: %s\n", progress.TaskID)
	fmt.Printf("Status: %s\n", progress.Status)
	fmt.Printf("Progress: %d%%\n", progress.Progress)
	fmt.Printf("Total Records: %d\n", progress.TotalRecords)
	fmt.Printf("Processed: %d\n", progress.Processed)

	if progress.ErrorMessage != "" {
		fmt.Printf("Error: %s\n", progress.ErrorMessage)
	}

	if progress.Status == common.TaskStatusCompleted {
		fmt.Printf("\n=== Download Info ===\n")
		fmt.Printf("File Name: %s\n", progress.FileName)
		fmt.Printf("File Size: %d bytes\n", progress.FileSize)
		if progress.ExpireAt != nil {
			fmt.Printf("Expire At: %s\n", progress.ExpireAt.Format("2006-01-02 15:04:05"))
		}
	}

	return nil
}

func (c *APIClient) DownloadTaskFile() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter task ID: ")
	taskID, _ := reader.ReadString('\n')
	taskID = strings.TrimSpace(taskID)

	resp, err := http.Get(serverAddr + "/api/tasks/" + taskID + "/download")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var apiResp common.APIResponse
		json.Unmarshal(body, &apiResp)
		return fmt.Errorf("[%d] %s", apiResp.Code, apiResp.Message)
	}

	contentDisp := resp.Header.Get("Content-Disposition")
	var fileName string
	if idx := strings.Index(contentDisp, "filename*=UTF-8''"); idx >= 0 {
		fileName = contentDisp[idx+len("filename*=UTF-8''"):]
	}
	if fileName == "" {
		fileName = "export_" + time.Now().Format("20060102150405") + ".csv"
	}

	fileName = strings.ReplaceAll(fileName, "%20", "_")
	fileName = strings.ReplaceAll(fileName, "%", "_")

	fmt.Print("Save as (leave blank to use default): ")
	saveName, _ := reader.ReadString('\n')
	saveName = strings.TrimSpace(saveName)
	if saveName == "" {
		saveName = fileName
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := os.WriteFile(saveName, data, 0644); err != nil {
		return err
	}

	fmt.Printf("File saved to: %s (%d bytes)\n", saveName, len(data))
	return nil
}

func (c *APIClient) GetStats() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Number of days to show (default 7): ")
	daysStr, _ := reader.ReadString('\n')
	daysStr = strings.TrimSpace(daysStr)
	days := 7
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil {
			days = d
		}
	}

	fmt.Print("Number of top templates (default 10): ")
	topStr, _ := reader.ReadString('\n')
	topStr = strings.TrimSpace(topStr)
	topN := 10
	if topStr != "" {
		if t, err := strconv.Atoi(topStr); err == nil {
			topN = t
		}
	}

	body, err := c.doRequest("GET", fmt.Sprintf("/api/stats?days=%d&top=%d", days, topN), nil)
	if err != nil {
		return err
	}

	var resp common.APIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if resp.Code != common.ErrCodeSuccess {
		return fmt.Errorf("[%d] %s", resp.Code, resp.Message)
	}

	var stats common.ExportStatsResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &stats)

	fmt.Println("\n=== Export Statistics ===")
	fmt.Printf("\n--- Overall Stats ---\n")
	fmt.Printf("Total Tasks: %d\n", stats.OverallStats.TotalTasks)
	fmt.Printf("Success Rate: %s\n", stats.OverallStats.SuccessRate)
	fmt.Printf("Average Duration: %d ms\n", stats.OverallStats.AvgDurationMs)

	fmt.Printf("\n--- Daily Stats (Last %d days) ---\n", days)
	for _, ds := range stats.DailyStats {
		fmt.Printf("  %s: Total=%d, Success=%d, Failed=%d, AvgDuration=%dms\n",
			ds.Date, ds.TotalTasks, ds.SuccessTasks, ds.FailedTasks, ds.AvgDurationMs)
	}

	fmt.Printf("\n--- Top %d Templates ---\n", topN)
	for i, ts := range stats.TopTemplates {
		name := ts.TemplateName
		if name == "" {
			name = ts.TemplateID
		}
		fmt.Printf("  %d. %s: Usage=%d, AvgDuration=%dms\n",
			i+1, name, ts.UsageCount, ts.AvgDurationMs)
	}

	return nil
}

func showHelp() {
	fmt.Println("\n=== Data Export Client Commands ===")
	fmt.Println("  template list        - List all templates")
	fmt.Println("  template create      - Create a new template")
	fmt.Println("  template delete      - Delete a template")
	fmt.Println("  export create        - Submit an export task")
	fmt.Println("  export status        - Check task progress")
	fmt.Println("  export download      - Download exported file")
	fmt.Println("  stats                - View export statistics")
	fmt.Println("  user <user_id>       - Switch current user")
	fmt.Println("  server <url>         - Set server address")
	fmt.Println("  help                 - Show this help")
	fmt.Println("  exit                 - Exit the client")
	fmt.Println()
}

func main() {
	client := NewAPIClient(serverAddr)
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== Data Export Client ===")
	fmt.Printf("Server: %s\n", serverAddr)
	fmt.Printf("Current User: %s\n", currentUser)
	fmt.Println("Type 'help' for available commands")

	for {
		fmt.Print("\n> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		cmd := strings.ToLower(parts[0])

		var err error
		switch cmd {
		case "exit", "quit":
			fmt.Println("Goodbye!")
			return

		case "help":
			showHelp()

		case "template":
			if len(parts) < 2 {
				fmt.Println("Usage: template <list|create|delete>")
				continue
			}
			subCmd := strings.ToLower(parts[1])
			switch subCmd {
			case "list":
				err = client.ListTemplates()
			case "create":
				err = client.CreateTemplate()
			case "delete":
				err = client.DeleteTemplate()
			default:
				fmt.Println("Unknown template command. Use: list, create, delete")
			}

		case "export":
			if len(parts) < 2 {
				fmt.Println("Usage: export <create|status|download>")
				continue
			}
			subCmd := strings.ToLower(parts[1])
			switch subCmd {
			case "create":
				err = client.CreateExportTask()
			case "status":
				err = client.CheckTaskProgress()
			case "download":
				err = client.DownloadTaskFile()
			default:
				fmt.Println("Unknown export command. Use: create, status, download")
			}

		case "stats":
			err = client.GetStats()

		case "user":
			if len(parts) < 2 {
				fmt.Printf("Current user: %s\n", currentUser)
				continue
			}
			currentUser = parts[1]
			fmt.Printf("Switched to user: %s\n", currentUser)

		case "server":
			if len(parts) < 2 {
				fmt.Printf("Current server: %s\n", serverAddr)
				continue
			}
			serverAddr = parts[1]
			client = NewAPIClient(serverAddr)
			fmt.Printf("Server set to: %s\n", serverAddr)

		default:
			fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", cmd)
		}

		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}
