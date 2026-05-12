package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const baseURL = "http://localhost:8080/api"

type Indicator struct {
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Formula      string  `json:"formula"`
	SourceDept   string  `json:"source_dept"`
	TargetValue  float64 `json:"target_value"`
	WarningValue float64 `json:"warning_value"`
	Category     string  `json:"category"`
}

type PDCA struct {
	Name        string `json:"name"`
	IndicatorID *uint  `json:"indicator_id"`
	Responsible string `json:"responsible"`
	StartDate   string `json:"start_date"`
}

type IndicatorData struct {
	IndicatorID uint    `json:"indicator_id"`
	Month       string  `json:"month"`
	Value       float64 `json:"value"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mqs-cli",
		Short: "医疗质量管理系统CLI",
	}

	var code, name, formula, dept, category string
	var target, warning float64

	var addIndicatorCmd = &cobra.Command{
		Use:   "add-indicator",
		Short: "添加质量指标",
		Run: func(cmd *cobra.Command, args []string) {
			indicator := Indicator{
				Code:         code,
				Name:         name,
				Formula:      formula,
				SourceDept:   dept,
				TargetValue:  target,
				WarningValue: warning,
				Category:     category,
			}

			body, _ := json.Marshal(indicator)
			resp, err := http.Post(baseURL+"/indicators", "application/json", bytes.NewBuffer(body))
			if err != nil {
				fmt.Println("请求失败:", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == 201 {
				fmt.Println("指标添加成功!")
			} else {
				data, _ := io.ReadAll(resp.Body)
				fmt.Printf("添加失败: %s\n", string(data))
			}
		},
	}

	addIndicatorCmd.Flags().StringVarP(&code, "code", "c", "", "指标编号 (必填)")
	addIndicatorCmd.Flags().StringVarP(&name, "name", "n", "", "指标名称 (必填)")
	addIndicatorCmd.Flags().StringVarP(&formula, "formula", "f", "", "计算公式")
	addIndicatorCmd.Flags().StringVarP(&dept, "dept", "d", "", "数据来源科室")
	addIndicatorCmd.Flags().Float64VarP(&target, "target", "t", 0, "目标值 (必填)")
	addIndicatorCmd.Flags().Float64VarP(&warning, "warning", "w", 0, "预警阈值 (必填)")
	addIndicatorCmd.Flags().StringVarP(&category, "category", "C", "", "指标类别")
	addIndicatorCmd.MarkFlagRequired("code")
	addIndicatorCmd.MarkFlagRequired("name")
	addIndicatorCmd.MarkFlagRequired("target")
	addIndicatorCmd.MarkFlagRequired("warning")

	var pdcaName, responsible, startDate string
	var indicatorID uint

	var createPDCACmd = &cobra.Command{
		Use:   "create-pdca",
		Short: "创建PDCA项目",
		Run: func(cmd *cobra.Command, args []string) {
			pdca := PDCA{
				Name:        pdcaName,
				Responsible: responsible,
				StartDate:   startDate,
			}
			if indicatorID > 0 {
				pdca.IndicatorID = &indicatorID
			}

			body, _ := json.Marshal(pdca)
			resp, err := http.Post(baseURL+"/pdca", "application/json", bytes.NewBuffer(body))
			if err != nil {
				fmt.Println("请求失败:", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == 201 {
				fmt.Println("PDCA项目创建成功!")
			} else {
				data, _ := io.ReadAll(resp.Body)
				fmt.Printf("创建失败: %s\n", string(data))
			}
		},
	}

	createPDCACmd.Flags().StringVarP(&pdcaName, "name", "n", "", "项目名称 (必填)")
	createPDCACmd.Flags().UintVarP(&indicatorID, "indicator", "i", 0, "关联指标ID")
	createPDCACmd.Flags().StringVarP(&responsible, "responsible", "r", "", "负责人 (必填)")
	createPDCACmd.Flags().StringVarP(&startDate, "date", "d", "", "开始日期 (必填)")
	createPDCACmd.MarkFlagRequired("name")
	createPDCACmd.MarkFlagRequired("responsible")
	createPDCACmd.MarkFlagRequired("date")

	var id uint
	var month string
	var value float64

	var recordDataCmd = &cobra.Command{
		Use:   "record-data",
		Short: "录入指标数据",
		Run: func(cmd *cobra.Command, args []string) {
			data := IndicatorData{
				IndicatorID: id,
				Month:       month,
				Value:       value,
			}

			body, _ := json.Marshal(data)
			resp, err := http.Post(fmt.Sprintf("%s/indicators/%d/data", baseURL, id), "application/json", bytes.NewBuffer(body))
			if err != nil {
				fmt.Println("请求失败:", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == 201 {
				fmt.Println("数据录入成功!")
			} else {
				respData, _ := io.ReadAll(resp.Body)
				fmt.Printf("录入失败: %s\n", string(respData))
			}
		},
	}

	recordDataCmd.Flags().UintVarP(&id, "id", "i", 0, "指标ID (必填)")
	recordDataCmd.Flags().StringVarP(&month, "month", "m", "", "月份 (YYYY-MM, 必填)")
	recordDataCmd.Flags().Float64VarP(&value, "value", "v", 0, "数值 (必填)")
	recordDataCmd.MarkFlagRequired("id")
	recordDataCmd.MarkFlagRequired("month")
	recordDataCmd.MarkFlagRequired("value")

	var aggType, aggMonth, aggPeriod string

	var showAggCmd = &cobra.Command{
		Use:   "show-aggregation",
		Short: "查看聚合结果",
		Run: func(cmd *cobra.Command, args []string) {
			var url string
			switch strings.ToLower(aggType) {
			case "department", "dept":
				url = baseURL + "/aggregations/department"
				if aggMonth != "" {
					url += "?month=" + aggMonth
				}
			case "category", "cat":
				url = baseURL + "/aggregations/category"
			case "time":
				url = baseURL + "/aggregations/time"
				if aggPeriod != "" {
					url += "?period=" + aggPeriod
				}
			default:
				fmt.Println("请指定聚合类型: --type department|category|time")
				return
			}

			resp, err := http.Get(url)
			if err != nil {
				fmt.Println("请求失败:", err)
				return
			}
			defer resp.Body.Close()

			data, _ := io.ReadAll(resp.Body)
			var prettyJSON bytes.Buffer
			json.Indent(&prettyJSON, data, "", "  ")
			fmt.Println(prettyJSON.String())
		},
	}

	showAggCmd.Flags().StringVarP(&aggType, "type", "t", "department", "聚合类型: department|category|time")
	showAggCmd.Flags().StringVarP(&aggMonth, "month", "m", "", "月份 (YYYY-MM)")
	showAggCmd.Flags().StringVarP(&aggPeriod, "period", "p", "", "时间周期: month|quarter|year")

	rootCmd.AddCommand(addIndicatorCmd, createPDCACmd, recordDataCmd, showAggCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
