package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"qc-process/internal/config"
	"qc-process/internal/models"
	"qc-process/internal/service"
	"qc-process/internal/storage"
)

var rootCmd = &cobra.Command{
	Use:   "qc-process",
	Short: "质检流程管理系统",
	Long:  `一个基于 Go 和 Cobra 构建的质检流程管理系统，支持 AQL 抽样、状态流转、检测记录和报告导出。`,
}

var batchNo string
var batchSize int
var inspectionLevel string
var configFile string
var reportID string
var outputFile string
var itemName string
var actualValue float64
var unqualifiedData string
var recheckItemsData string
var reworkReason string

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(samplingCmd)
	rootCmd.AddCommand(testingCmd)
	rootCmd.AddCommand(recordCmd)
	rootCmd.AddCommand(completeTestCmd)
	rootCmd.AddCommand(evaluateCmd)
	rootCmd.AddCommand(recheckCmd)
	rootCmd.AddCommand(performRecheckCmd)
	rootCmd.AddCommand(finalCheckCmd)
	rootCmd.AddCommand(reworkCmd)
	rootCmd.AddCommand(completeReworkCmd)
	rootCmd.AddCommand(archiveCmd)
	rootCmd.AddCommand(exportCmd)

	createCmd.Flags().StringVarP(&batchNo, "batch", "b", "", "批次号 (必填)")
	createCmd.Flags().IntVarP(&batchSize, "size", "s", 0, "批量大小 (必填，必须大于 0)")
	createCmd.Flags().StringVarP(&inspectionLevel, "level", "l", "II", "检验级别 (I, II, III, S1, S2, S3, S4)")
	createCmd.Flags().StringVarP(&configFile, "config", "c", "config/items.yaml", "检测项目配置文件 (YAML)")
	createCmd.MarkFlagRequired("batch")
	createCmd.MarkFlagRequired("size")

	recordCmd.Flags().StringVarP(&reportID, "id", "i", "", "质检单 ID (必填)")
	recordCmd.Flags().StringVarP(&itemName, "item", "n", "", "检测项目名称 (必填)")
	recordCmd.Flags().Float64VarP(&actualValue, "value", "v", 0, "实际检测值 (必填)")
	recordCmd.MarkFlagRequired("id")
	recordCmd.MarkFlagRequired("item")
	recordCmd.MarkFlagRequired("value")

	completeTestCmd.Flags().StringVarP(&reportID, "id", "i", "", "质检单 ID (必填)")
	completeTestCmd.Flags().StringVarP(&unqualifiedData, "unqualified", "u", "", "不合格项 JSON 数据 (选填)")
	completeTestCmd.MarkFlagRequired("id")

	recheckCmd.Flags().StringVarP(&reportID, "id", "i", "", "质检单 ID (必填)")
	recheckCmd.MarkFlagRequired("id")

	performRecheckCmd.Flags().StringVarP(&reportID, "id", "i", "", "质检单 ID (必填)")
	performRecheckCmd.Flags().StringVarP(&recheckItemsData, "items", "t", "", "复检项目 JSON 数据 (必填)")
	performRecheckCmd.MarkFlagRequired("id")
	performRecheckCmd.MarkFlagRequired("items")

	finalCheckCmd.Flags().StringVarP(&reportID, "id", "i", "", "质检单 ID (必填)")
	finalCheckCmd.MarkFlagRequired("id")

	reworkCmd.Flags().StringVarP(&reportID, "id", "i", "", "质检单 ID (必填)")
	reworkCmd.Flags().StringVarP(&reworkReason, "reason", "r", "", "返工原因 (必填)")
	reworkCmd.MarkFlagRequired("id")
	reworkCmd.MarkFlagRequired("reason")

	completeReworkCmd.Flags().StringVarP(&reportID, "id", "i", "", "质检单 ID (必填)")
	completeReworkCmd.MarkFlagRequired("id")

	samplingCmd.Flags().StringVarP(&reportID, "id", "i", "", "质检单 ID (必填)")
	samplingCmd.MarkFlagRequired("id")

	testingCmd.Flags().StringVarP(&reportID, "id", "i", "", "质检单 ID (必填)")
	testingCmd.MarkFlagRequired("id")

	evaluateCmd.Flags().StringVarP(&reportID, "id", "i", "", "质检单 ID (必填)")
	evaluateCmd.MarkFlagRequired("id")

	archiveCmd.Flags().StringVarP(&reportID, "id", "i", "", "质检单 ID (必填)")
	archiveCmd.MarkFlagRequired("id")

	showCmd.Flags().StringVarP(&reportID, "id", "i", "", "质检单 ID (必填)")
	showCmd.MarkFlagRequired("id")

	exportCmd.Flags().StringVarP(&reportID, "id", "i", "", "质检单 ID (必填)")
	exportCmd.Flags().StringVarP(&outputFile, "output", "o", "", "输出文件路径 (必填)")
	exportCmd.MarkFlagRequired("id")
	exportCmd.MarkFlagRequired("output")
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "创建新的质检单",
	RunE: func(cmd *cobra.Command, args []string) error {
		items, err := config.LoadItemsFromYAML(configFile)
		if err != nil {
			return err
		}

		report, err := service.CreateQCReport(batchNo, batchSize, inspectionLevel, items)
		if err != nil {
			return err
		}

		if err := storage.SaveReport(report); err != nil {
			return err
		}

		fmt.Printf("质检单创建成功！\nID: %s\n批次号: %s\n批量大小: %d\n检验级别: %s\n抽样数量: %d\n当前状态: %s\n",
			report.ID, report.BatchNo, report.BatchSize, report.InspectionLevel, report.SampleSize, report.Status)
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有质检单",
	RunE: func(cmd *cobra.Command, args []string) error {
		ids, err := storage.ListReports()
		if err != nil {
			return err
		}

		if len(ids) == 0 {
			fmt.Println("暂无质检单")
			return nil
		}

		fmt.Println("质检单列表:")
		for _, id := range ids {
			report, err := storage.LoadReport(id)
			if err != nil {
				continue
			}
			fmt.Printf("ID: %s, 批次号: %s, 状态: %s\n", report.ID, report.BatchNo, report.Status)
		}
		return nil
	},
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "显示质检单详情",
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := storage.LoadReport(reportID)
		if err != nil {
			return err
		}

		data, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(data))
		return nil
	},
}

var samplingCmd = &cobra.Command{
	Use:   "sampling",
	Short: "开始抽样",
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := storage.LoadReport(reportID)
		if err != nil {
			return err
		}

		if err := service.TransitionToSampling(report); err != nil {
			return err
		}

		if err := storage.SaveReport(report); err != nil {
			return err
		}

		fmt.Printf("已进入抽样中状态。当前状态: %s\n", report.Status)
		return nil
	},
}

var testingCmd = &cobra.Command{
	Use:   "testing",
	Short: "开始检测",
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := storage.LoadReport(reportID)
		if err != nil {
			return err
		}

		if err := service.TransitionToTesting(report); err != nil {
			return err
		}

		if err := storage.SaveReport(report); err != nil {
			return err
		}

		fmt.Printf("已进入检测中状态。当前状态: %s\n", report.Status)
		return nil
	},
}

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "记录检测结果",
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := storage.LoadReport(reportID)
		if err != nil {
			return err
		}

		if err := service.RecordTestResult(report, itemName, actualValue); err != nil {
			return err
		}

		if err := storage.SaveReport(report); err != nil {
			return err
		}

		fmt.Printf("检测结果已记录。项目: %s, 实际值: %.2f\n", itemName, actualValue)
		return nil
	},
}

var completeTestCmd = &cobra.Command{
	Use:   "complete-test",
	Short: "完成检测",
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := storage.LoadReport(reportID)
		if err != nil {
			return err
		}

		var unqualifiedRecords []models.UnqualifiedRecord
		if unqualifiedData != "" {
			if err := json.Unmarshal([]byte(unqualifiedData), &unqualifiedRecords); err != nil {
				return fmt.Errorf("错误：不合格项 JSON 格式错误: %v", err)
			}
		}

		if err := service.CompleteTesting(report, unqualifiedRecords); err != nil {
			return err
		}

		if err := storage.SaveReport(report); err != nil {
			return err
		}

		fmt.Printf("检测已完成。当前状态: %s\n", report.Status)
		return nil
	},
}

var evaluateCmd = &cobra.Command{
	Use:   "evaluate",
	Short: "评估检测结果",
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := storage.LoadReport(reportID)
		if err != nil {
			return err
		}

		status, err := service.EvaluateResult(report)
		if err != nil {
			return err
		}

		if err := storage.SaveReport(report); err != nil {
			return err
		}

		fmt.Printf("评估完成。状态: %s, 最终结果: %s\n", status, report.FinalResult)
		return nil
	},
}

var recheckCmd = &cobra.Command{
	Use:   "recheck",
	Short: "申请复检",
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := storage.LoadReport(reportID)
		if err != nil {
			return err
		}

		if err := service.RequestRecheck(report); err != nil {
			return err
		}

		if err := storage.SaveReport(report); err != nil {
			return err
		}

		fmt.Printf("复检申请已提交。当前状态: %s\n", report.Status)
		return nil
	},
}

var performRecheckCmd = &cobra.Command{
	Use:   "perform-recheck",
	Short: "执行复检",
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := storage.LoadReport(reportID)
		if err != nil {
			return err
		}

		var recheckItems []models.InspectionItem
		if err := json.Unmarshal([]byte(recheckItemsData), &recheckItems); err != nil {
			return fmt.Errorf("错误：复检项目 JSON 格式错误: %v", err)
		}

		if err := service.PerformRecheck(report, recheckItems); err != nil {
			return err
		}

		if err := storage.SaveReport(report); err != nil {
			return err
		}

		fmt.Printf("复检已完成。当前状态: %s\n", report.Status)
		return nil
	},
}

var finalCheckCmd = &cobra.Command{
	Use:   "final-check",
	Short: "执行终检",
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := storage.LoadReport(reportID)
		if err != nil {
			return err
		}

		status, err := service.FinalCheck(report)
		if err != nil {
			return err
		}

		if err := storage.SaveReport(report); err != nil {
			return err
		}

		fmt.Printf("终检完成。状态: %s, 最终结果: %s\n", status, report.FinalResult)
		return nil
	},
}

var reworkCmd = &cobra.Command{
	Use:   "rework",
	Short: "创建返工工单",
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := storage.LoadReport(reportID)
		if err != nil {
			return err
		}

		if err := service.CreateReworkOrder(report, reworkReason); err != nil {
			return err
		}

		if err := storage.SaveReport(report); err != nil {
			return err
		}

		fmt.Printf("返工工单已创建。返工单号: %s\n", report.ReworkOrder.ID)
		return nil
	},
}

var completeReworkCmd = &cobra.Command{
	Use:   "complete-rework",
	Short: "完成返工",
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := storage.LoadReport(reportID)
		if err != nil {
			return err
		}

		if err := service.CompleteRework(report); err != nil {
			return err
		}

		if err := storage.SaveReport(report); err != nil {
			return err
		}

		fmt.Printf("返工已完成，已重置为待抽样状态。当前状态: %s\n", report.Status)
		return nil
	},
}

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "归档质检单",
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := storage.LoadReport(reportID)
		if err != nil {
			return err
		}

		if err := service.Archive(report); err != nil {
			return err
		}

		if err := storage.SaveReport(report); err != nil {
			return err
		}

		fmt.Printf("质检单已归档。当前状态: %s\n", report.Status)
		return nil
	},
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "导出质检报告为 JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := storage.LoadReport(reportID)
		if err != nil {
			return err
		}

		if err := storage.ExportReportToJSON(report, outputFile); err != nil {
			return err
		}

		fmt.Printf("报告已导出到: %s\n", outputFile)
		return nil
	},
}

func parseUnqualifiedData(data string) ([]models.UnqualifiedRecord, error) {
	if data == "" {
		return nil, nil
	}

	var records []models.UnqualifiedRecord
	if err := json.Unmarshal([]byte(data), &records); err != nil {
		return nil, fmt.Errorf("错误：不合格项 JSON 格式错误: %v", err)
	}

	for i, record := range records {
		records[i].Reason = strings.TrimSpace(record.Reason)
		if records[i].Reason == "" {
			return nil, fmt.Errorf("错误：不合格项 %d 未提供原因", i+1)
		}

		severity := strings.ToLower(strings.TrimSpace(string(record.Severity)))
		switch severity {
		case "轻微":
			records[i].Severity = models.SeverityMinor
		case "一般":
			records[i].Severity = models.SeverityGeneral
		case "严重":
			records[i].Severity = models.SeverityMajor
		default:
			return nil, fmt.Errorf("错误：不合格项 %d 严重程度 %q 无效，支持: 轻微/一般/严重", i+1, record.Severity)
		}
	}

	return records, nil
}

func parseFloatFlagToFloat64(val string) (float64, error) {
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, err
	}
	return f, nil
}
