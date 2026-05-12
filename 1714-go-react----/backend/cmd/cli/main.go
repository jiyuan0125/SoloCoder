package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

const baseURL = "http://localhost:8080/api"

func main() {
	rootCmd := &cobra.Command{
		Use:   "hsc",
		Short: "卫生监督管理系统CLI",
	}

	rootCmd.AddCommand(createInspectionCmd())
	rootCmd.AddCommand(issuePenaltyCmd())
	rootCmd.AddCommand(publishNoticeCmd())
	rootCmd.AddCommand(viewAuditCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createInspectionCmd() *cobra.Command {
	var (
		unitID         uint
		inspectionDate string
		inspectors     string
		inspectionType string
		items          []string
	)

	cmd := &cobra.Command{
		Use:   "create-inspection",
		Short: "创建监督检查记录",
		Run: func(cmd *cobra.Command, args []string) {
			if inspectionDate == "" {
				inspectionDate = time.Now().Format("2006-01-02")
			}

			parsedItems := []map[string]interface{}{}
			for _, item := range items {
				parsedItems = append(parsedItems, map[string]interface{}{
					"item_name": item,
					"result":    "合格",
				})
			}

			data := map[string]interface{}{
				"unit_id":         unitID,
				"inspection_date": inspectionDate,
				"inspectors":      inspectors,
				"inspection_type": inspectionType,
				"items":           parsedItems,
			}

			resp, err := makeRequest("POST", "/inspections", data)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println(resp)
		},
	}

	cmd.Flags().UintVar(&unitID, "unit-id", 0, "被监督单位ID (必填)")
	cmd.Flags().StringVar(&inspectionDate, "date", "", "检查日期 (默认今天)")
	cmd.Flags().StringVar(&inspectors, "inspectors", "", "检查人员 (用逗号分隔，至少2人)")
	cmd.Flags().StringVar(&inspectionType, "type", "日常监督", "检查类型: 日常监督/专项检查/投诉举报核查")
	cmd.Flags().StringSliceVar(&items, "items", []string{}, "检查项列表")

	cmd.MarkFlagRequired("unit-id")
	cmd.MarkFlagRequired("inspectors")

	return cmd
}

func issuePenaltyCmd() *cobra.Command {
	var (
		opinionID   uint
		penaltyType string
		fineAmount  float64
		fineReason  string
		remarks     string
	)

	cmd := &cobra.Command{
		Use:   "issue-penalty",
		Short: "下达行政处罚",
		Run: func(cmd *cobra.Command, args []string) {
			data := map[string]interface{}{
				"opinion_id":    opinionID,
				"penalty_types": penaltyType,
				"fine_amount":   fineAmount,
				"fine_reason":   fineReason,
				"remarks":       remarks,
			}

			resp, err := makeRequest("POST", "/penalties", data)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println(resp)
		},
	}

	cmd.Flags().UintVar(&opinionID, "opinion-id", 0, "卫生监督意见书ID (必填)")
	cmd.Flags().StringVar(&penaltyType, "type", "", "处罚类型: 警告/罚款/停业整顿/吊销许可证")
	cmd.Flags().Float64Var(&fineAmount, "fine", 0, "罚款金额")
	cmd.Flags().StringVar(&fineReason, "reason", "", "罚款原因")
	cmd.Flags().StringVar(&remarks, "remarks", "", "备注")

	cmd.MarkFlagRequired("opinion-id")
	cmd.MarkFlagRequired("type")

	return cmd
}

func publishNoticeCmd() *cobra.Command {
	var (
		unitID       uint
		title        string
		content      string
		noticeType   string
		durationDays int
	)

	cmd := &cobra.Command{
		Use:   "publish-notice",
		Short: "发布公示",
		Run: func(cmd *cobra.Command, args []string) {
			if durationDays == 0 {
				durationDays = 30
			}

			data := map[string]interface{}{
				"unit_id":       unitID,
				"title":         title,
				"content":       content,
				"notice_type":   noticeType,
				"duration_days": durationDays,
			}

			resp, err := makeRequest("POST", "/notices", data)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println(resp)
		},
	}

	cmd.Flags().UintVar(&unitID, "unit-id", 0, "被监督单位ID (必填)")
	cmd.Flags().StringVar(&title, "title", "", "公示标题 (必填)")
	cmd.Flags().StringVar(&content, "content", "", "公示内容 (必填)")
	cmd.Flags().StringVar(&noticeType, "type", "检查结果", "公示类型")
	cmd.Flags().IntVar(&durationDays, "duration", 30, "公示天数 (默认30天)")

	cmd.MarkFlagRequired("unit-id")
	cmd.MarkFlagRequired("title")
	cmd.MarkFlagRequired("content")

	return cmd
}

func viewAuditCmd() *cobra.Command {
	var (
		action   string
		resource string
	)

	cmd := &cobra.Command{
		Use:   "view-audit",
		Short: "查看审计日志",
		Run: func(cmd *cobra.Command, args []string) {
			url := "/audit"
			query := ""
			if action != "" {
				query += "?action=" + action
			}
			if resource != "" {
				if query != "" {
					query += "&"
				} else {
					query += "?"
				}
				query += "resource=" + resource
			}

			resp, err := makeRequest("GET", url+query, nil)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println(resp)
		},
	}

	cmd.Flags().StringVar(&action, "action", "", "按操作类型筛选")
	cmd.Flags().StringVar(&resource, "resource", "", "按资源类型筛选")

	return cmd
}

func makeRequest(method, path string, data interface{}) (string, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+path, body)
	if err != nil {
		return "", err
	}

	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, respBody, "", "  "); err == nil {
		return prettyJSON.String(), nil
	}
	return string(respBody), nil
}
