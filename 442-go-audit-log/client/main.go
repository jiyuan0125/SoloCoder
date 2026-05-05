package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"audit-log/common"
)

const defaultServerURL = "http://localhost:8080"

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: serverURL,
	}
}

func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, c.serverURL+endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return respBody, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	return respBody, nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := os.Getenv("AUDIT_SERVER_URL")
	if serverURL == "" {
		serverURL = defaultServerURL
	}

	client := NewClient(serverURL)

	command := os.Args[1]

	switch command {
	case "help":
		printUsage()
	case "health":
		handleHealth(client)
	case "log":
		handleLog(client, os.Args[2:])
	case "get":
		handleGet(client, os.Args[2:])
	case "query":
		handleQuery(client, os.Args[2:])
	case "login":
		handleLogin(client, os.Args[2:])
	case "lock-status":
		handleLockStatus(client, os.Args[2:])
	case "export-approval":
		handleExportApproval(client, os.Args[2:])
	case "archive-trigger":
		handleArchiveTrigger(client)
	case "archive-list":
		handleArchiveList(client)
	case "storage-status":
		handleStorageStatus(client)
	case "statistics":
		handleStatistics(client, os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Audit Log Client - Command Line Interface

Usage:
  client <command> [options]

Commands:
  help                          Show this help message
  health                        Check server health
  
  log [options]                 Create a new audit log
    -user_id string      User ID (required)
    -user_name string    User name
    -ip string           IP address
    -type string         Operation type: create/update/delete/login/logout/export/import/query (required)
    -desc string         Description
    -result string       Result: success/failed (default: success)
    -record_ids string   Comma-separated record IDs
    -snapshot string     JSON string of deleted data snapshot (for delete operations)
    -export_range string JSON string of export range (for export operations)
    -approver string     Approver user ID (for sensitive operations)
  
  get <log_id>                  Get a specific log by ID
  
  query [options]               Query audit logs
    -user_id string      Filter by user ID
    -type string         Filter by operation type
    -ip string           Filter by IP address
    -start string        Start time (RFC3339 format)
    -end string          End time (RFC3339 format)
    -abnormal            Filter only abnormal logs
    -page int            Page number (default: 1)
    -size int            Page size (default: 20)
  
  login [options]               Record login attempt
    -user_id string      User ID (required)
    -user_name string    User name
    -ip string           IP address
    -failed              Mark as failed login
  
  lock-status -user_id <id>    Check user lock status
  
  export-approval [options]    Request or grant export approval
    -user_id string      User ID (required)
    -approver_id string  Approver user ID (for granting approval)
    -grant               Grant approval (requires approver_id)
  
  archive-trigger               Trigger archive of old logs
  archive-list                  List archived logs
  storage-status                Check storage status
  
  statistics [options]          Get audit statistics
    -start string        Start time (RFC3339 format)
    -end string          End time (RFC3339 format)

Environment:
  AUDIT_SERVER_URL      Server URL (default: http://localhost:8080)`)
}

func handleHealth(client *Client) {
	body, err := client.doRequest(http.MethodGet, "/health", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		if len(body) > 0 {
			fmt.Println(string(body))
		}
		os.Exit(1)
	}
	fmt.Println(string(body))
}

func handleLog(client *Client, args []string) {
	fs := flag.NewFlagSet("log", flag.ExitOnError)
	userID := fs.String("user_id", "", "User ID")
	userName := fs.String("user_name", "", "User name")
	ip := fs.String("ip", "", "IP address")
	opType := fs.String("type", "", "Operation type")
	desc := fs.String("desc", "", "Description")
	result := fs.String("result", common.ResultSuccess, "Result")
	recordIDsStr := fs.String("record_ids", "", "Comma-separated record IDs")
	snapshotStr := fs.String("snapshot", "", "JSON snapshot")
	exportRangeStr := fs.String("export_range", "", "JSON export range")
	approver := fs.String("approver", "", "Approver ID")

	fs.Parse(args)

	if *userID == "" || *opType == "" {
		fmt.Println("Error: -user_id and -type are required")
		os.Exit(1)
	}

	var recordIDs []string
	if *recordIDsStr != "" {
		var ids []string
		if err := json.Unmarshal([]byte(*recordIDsStr), &ids); err == nil {
			recordIDs = ids
		}
	}

	var snapshot map[string]interface{}
	if *snapshotStr != "" {
		json.Unmarshal([]byte(*snapshotStr), &snapshot)
	}

	var exportRange *common.ExportRange
	if *exportRangeStr != "" {
		json.Unmarshal([]byte(*exportRangeStr), &exportRange)
	}

	req := common.CreateLogRequest{
		UserID:        *userID,
		UserName:      *userName,
		IPAddress:     *ip,
		OperationType: *opType,
		Description:   *desc,
		Result:        *result,
		RecordIDs:     recordIDs,
		Snapshot:      snapshot,
		ExportRange:   exportRange,
		Approver:      *approver,
	}

	body, err := client.doRequest(http.MethodPost, "/api/logs", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		if len(body) > 0 {
			fmt.Println(string(body))
		}
		os.Exit(1)
	}
	fmt.Println(string(body))
}

func handleGet(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Error: log_id is required")
		os.Exit(1)
	}
	logID := args[0]

	body, err := client.doRequest(http.MethodGet, "/api/logs/"+logID, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		if len(body) > 0 {
			fmt.Println(string(body))
		}
		os.Exit(1)
	}
	fmt.Println(string(body))
}

func handleQuery(client *Client, args []string) {
	fs := flag.NewFlagSet("query", flag.ExitOnError)
	userID := fs.String("user_id", "", "User ID")
	opType := fs.String("type", "", "Operation type")
	ip := fs.String("ip", "", "IP address")
	startStr := fs.String("start", "", "Start time")
	endStr := fs.String("end", "", "End time")
	abnormal := fs.Bool("abnormal", false, "Only abnormal")
	page := fs.Int("page", 1, "Page number")
	size := fs.Int("size", 20, "Page size")

	fs.Parse(args)

	var startTime, endTime time.Time
	if *startStr != "" {
		startTime, _ = time.Parse(time.RFC3339, *startStr)
	}
	if *endStr != "" {
		endTime, _ = time.Parse(time.RFC3339, *endStr)
	}

	req := common.QueryLogsRequest{
		UserID:        *userID,
		OperationType: *opType,
		IPAddress:     *ip,
		StartTime:     startTime,
		EndTime:       endTime,
		IsAbnormal:    *abnormal,
		Page:          *page,
		PageSize:      *size,
	}

	body, err := client.doRequest(http.MethodPost, "/api/logs/query", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		if len(body) > 0 {
			fmt.Println(string(body))
		}
		os.Exit(1)
	}
	fmt.Println(string(body))
}

func handleLogin(client *Client, args []string) {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	userID := fs.String("user_id", "", "User ID")
	userName := fs.String("user_name", "", "User name")
	ip := fs.String("ip", "", "IP address")
	failed := fs.Bool("failed", false, "Failed login")

	fs.Parse(args)

	if *userID == "" {
		fmt.Println("Error: -user_id is required")
		os.Exit(1)
	}

	req := common.LoginRequest{
		UserID:    *userID,
		UserName:  *userName,
		IPAddress: *ip,
		Success:   !*failed,
	}

	body, err := client.doRequest(http.MethodPost, "/api/login", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		if len(body) > 0 {
			fmt.Println(string(body))
		}
		os.Exit(1)
	}
	fmt.Println(string(body))
}

func handleLockStatus(client *Client, args []string) {
	fs := flag.NewFlagSet("lock-status", flag.ExitOnError)
	userID := fs.String("user_id", "", "User ID")
	fs.Parse(args)

	if *userID == "" {
		fmt.Println("Error: -user_id is required")
		os.Exit(1)
	}

	body, err := client.doRequest(http.MethodGet, "/api/user/lock-status?user_id="+*userID, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		if len(body) > 0 {
			fmt.Println(string(body))
		}
		os.Exit(1)
	}
	fmt.Println(string(body))
}

func handleExportApproval(client *Client, args []string) {
	fs := flag.NewFlagSet("export-approval", flag.ExitOnError)
	userID := fs.String("user_id", "", "User ID")
	approverID := fs.String("approver_id", "", "Approver ID")
	grant := fs.Bool("grant", false, "Grant approval")
	fs.Parse(args)

	if *userID == "" {
		fmt.Println("Error: -user_id is required")
		os.Exit(1)
	}

	req := common.ExportApprovalRequest{
		UserID:     *userID,
		ApproverID: *approverID,
		Approval:   *grant,
	}

	body, err := client.doRequest(http.MethodPost, "/api/export/approval", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		if len(body) > 0 {
			fmt.Println(string(body))
		}
		os.Exit(1)
	}
	fmt.Println(string(body))
}

func handleArchiveTrigger(client *Client) {
	body, err := client.doRequest(http.MethodPost, "/api/archive/trigger", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		if len(body) > 0 {
			fmt.Println(string(body))
		}
		os.Exit(1)
	}
	fmt.Println(string(body))
}

func handleArchiveList(client *Client) {
	body, err := client.doRequest(http.MethodGet, "/api/archive/list", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		if len(body) > 0 {
			fmt.Println(string(body))
		}
		os.Exit(1)
	}
	fmt.Println(string(body))
}

func handleStorageStatus(client *Client) {
	body, err := client.doRequest(http.MethodGet, "/api/storage/status", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		if len(body) > 0 {
			fmt.Println(string(body))
		}
		os.Exit(1)
	}
	fmt.Println(string(body))
}

func handleStatistics(client *Client, args []string) {
	fs := flag.NewFlagSet("statistics", flag.ExitOnError)
	startStr := fs.String("start", "", "Start time")
	endStr := fs.String("end", "", "End time")
	fs.Parse(args)

	var startTime, endTime time.Time
	if *startStr != "" {
		startTime, _ = time.Parse(time.RFC3339, *startStr)
	}
	if *endStr != "" {
		endTime, _ = time.Parse(time.RFC3339, *endStr)
	}

	req := common.StatisticsRequest{
		StartTime: startTime,
		EndTime:   endTime,
	}

	body, err := client.doRequest(http.MethodPost, "/api/statistics", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		if len(body) > 0 {
			fmt.Println(string(body))
		}
		os.Exit(1)
	}
	fmt.Println(string(body))
}
