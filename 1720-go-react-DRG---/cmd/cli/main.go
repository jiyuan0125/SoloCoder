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

	"drg-system/internal/model"

	"github.com/urfave/cli/v2"
)

const baseURL = "http://localhost:8300/api"

func main() {
	app := &cli.App{
		Name:  "drg-cli",
		Usage: "DRG付费管理系统命令行工具",
		Commands: []*cli.Command{
			{
				Name:  "submit-settlement",
				Usage: "提交结算数据",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "hospital", Aliases: []string{"H"}, Required: true, Usage: "医院ID"},
					&cli.StringFlag{Name: "period", Aliases: []string{"p"}, Required: true, Usage: "结算周期 (格式: YYYY-MM)"},
					&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Usage: "病案数据JSON文件路径"},
				},
				Action: submitSettlement,
			},
			{
				Name:  "calculate-payment",
				Usage: "重新计算结算支付",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "settlement", Aliases: []string{"s"}, Required: true, Usage: "结算ID"},
				},
				Action: calculatePayment,
			},
			{
				Name:  "export-settlement",
				Usage: "导出结算数据CSV",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "settlement", Aliases: []string{"s"}, Required: true, Usage: "结算ID"},
					&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Required: true, Usage: "输出文件路径"},
				},
				Action: exportSettlement,
			},
			{
				Name:  "list-settlements",
				Usage: "列出所有结算",
				Action: listSettlements,
			},
			{
				Name:  "get-settlement",
				Usage: "获取结算详情",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "settlement", Aliases: []string{"s"}, Required: true, Usage: "结算ID"},
				},
				Action: getSettlement,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func httpPost(url string, body interface{}) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	return respBody, resp.StatusCode, err
}

func submitSettlement(c *cli.Context) error {
	hospitalID := c.String("hospital")
	period := c.String("period")
	filePath := c.String("file")

	var records []model.MedicalRecord
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("读取文件失败: %v", err)
		}
		if err := json.Unmarshal(data, &records); err != nil {
			return fmt.Errorf("解析JSON失败: %v", err)
		}
	} else {
		records = generateSampleRecords()
	}

	reqBody := map[string]interface{}{
		"hospital_id": hospitalID,
		"period":      period,
		"records":     records,
	}

	body, status, err := httpPost(baseURL+"/settlements", reqBody)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}

	if status != http.StatusCreated {
		return fmt.Errorf("提交失败 (状态码: %d): %s", status, string(body))
	}

	var settlement model.Settlement
	if err := json.Unmarshal(body, &settlement); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	fmt.Printf("结算创建成功!\n")
	fmt.Printf("结算ID: %s\n", settlement.ID)
	fmt.Printf("周期: %s\n", settlement.Period)
	fmt.Printf("状态: %s\n", settlement.Status)
	fmt.Printf("DRG组数: %d\n", len(settlement.GroupSummaries))
	fmt.Printf("实际总费用: %d分 (%.2f元)\n", settlement.TotalActualCost, float64(settlement.TotalActualCost)/100)
	fmt.Printf("应支付总额: %d分 (%.2f元)\n", settlement.TotalPayment, float64(settlement.TotalPayment)/100)
	fmt.Printf("差额: %d分 (%.2f元)\n", settlement.Difference, float64(settlement.Difference)/100)

	return nil
}

func calculatePayment(c *cli.Context) error {
	settlementID := c.String("settlement")

	body, status, err := httpPost(baseURL+"/settlements/"+settlementID+"/recalculate", nil)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}

	if status != http.StatusOK {
		return fmt.Errorf("计算失败 (状态码: %d): %s", status, string(body))
	}

	var settlement model.Settlement
	if err := json.Unmarshal(body, &settlement); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	fmt.Printf("重新计算成功!\n")
	fmt.Printf("结算ID: %s\n", settlement.ID)
	fmt.Printf("实际总费用: %d分 (%.2f元)\n", settlement.TotalActualCost, float64(settlement.TotalActualCost)/100)
	fmt.Printf("应支付总额: %d分 (%.2f元)\n", settlement.TotalPayment, float64(settlement.TotalPayment)/100)

	return nil
}

func exportSettlement(c *cli.Context) error {
	settlementID := c.String("settlement")
	outputPath := c.String("output")

	resp, err := http.Get(baseURL + "/settlements/" + settlementID + "/export")
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("导出失败 (状态码: %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	if err := os.WriteFile(outputPath, body, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	fmt.Printf("导出成功! 文件: %s\n", outputPath)
	return nil
}

func listSettlements(c *cli.Context) error {
	body, err := httpGet(baseURL + "/settlements")
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}

	var settlements []model.Settlement
	if err := json.Unmarshal(body, &settlements); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if len(settlements) == 0 {
		fmt.Println("暂无结算数据")
		return nil
	}

	fmt.Printf("%-30s %-15s %-15s %-20s\n", "ID", "医院", "周期", "状态")
	fmt.Println("--------------------------------------------------------------------------------")
	for _, s := range settlements {
		fmt.Printf("%-30s %-15s %-15s %-20s\n", s.ID, s.HospitalID, s.Period, s.Status)
	}

	return nil
}

func getSettlement(c *cli.Context) error {
	settlementID := c.String("settlement")

	body, err := httpGet(baseURL + "/settlements/" + settlementID)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}

	var settlement model.Settlement
	if err := json.Unmarshal(body, &settlement); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	fmt.Printf("结算ID: %s\n", settlement.ID)
	fmt.Printf("医院: %s\n", settlement.HospitalID)
	fmt.Printf("周期: %s\n", settlement.Period)
	fmt.Printf("状态: %s\n", settlement.Status)
	fmt.Printf("实际总费用: %.2f元\n", float64(settlement.TotalActualCost)/100)
	fmt.Printf("应支付总额: %.2f元\n", float64(settlement.TotalPayment)/100)
	fmt.Printf("差额: %.2f元\n", float64(settlement.Difference)/100)
	fmt.Println("\nDRG组汇总:")
	fmt.Printf("%-10s %-25s %-8s %-15s %-12s %-15s\n", "组编号", "组名称", "病案数", "实际费用", "效率指数", "调整后支付")
	fmt.Println("---------------------------------------------------------------------------------------------------")
	for _, s := range settlement.GroupSummaries {
		fmt.Printf("%-10s %-25s %-8d %-15.2f %-12.4f %-15.2f\n",
			s.DRGGroupCode,
			s.DRGGroupName,
			s.CaseCount,
			float64(s.TotalActualCost)/100,
			s.EfficiencyIndex,
			float64(s.AdjustedPayment)/100,
		)
	}

	return nil
}

func generateSampleRecords() []model.MedicalRecord {
	now := time.Now()
	caseNoBase := now.Unix()

	return []model.MedicalRecord{
		{
			CaseNo:        strconv.FormatInt(caseNoBase+1, 10),
			MainDiagnosis: "I21.0",
			MainProcedure: "",
			CCFlag:        model.CCFlagMCC,
			Age:           65,
			Gender:        "男",
			ActualCost:    850000,
		},
		{
			CaseNo:        strconv.FormatInt(caseNoBase+2, 10),
			MainDiagnosis: "I21.0",
			MainProcedure: "",
			CCFlag:        model.CCFlagNone,
			Age:           55,
			Gender:        "男",
			ActualCost:    620000,
		},
		{
			CaseNo:        strconv.FormatInt(caseNoBase+3, 10),
			MainDiagnosis: "J18.9",
			MainProcedure: "",
			CCFlag:        model.CCFlagCC,
			Age:           72,
			Gender:        "女",
			ActualCost:    280000,
		},
		{
			CaseNo:        strconv.FormatInt(caseNoBase+4, 10),
			MainDiagnosis: "K80.2",
			MainProcedure: "51.23",
			CCFlag:        model.CCFlagNone,
			Age:           45,
			Gender:        "女",
			ActualCost:    180000,
		},
		{
			CaseNo:        strconv.FormatInt(caseNoBase+5, 10),
			MainDiagnosis: "I50.9",
			MainProcedure: "",
			CCFlag:        model.CCFlagMCC,
			Age:           78,
			Gender:        "男",
			ActualCost:    420000,
		},
	}
}
