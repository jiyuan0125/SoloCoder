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

	"github.com/spf13/cobra"
)

const baseURL = "http://localhost:8300/api"

func main() {
	rootCmd := &cobra.Command{
		Use:   "hospital-infection",
		Short: "医院感染监测系统CLI",
	}

	rootCmd.AddCommand(reportInfectionCmd())
	rootCmd.AddCommand(addMeasureCmd())
	rootCmd.AddCommand(generateReportCmd())
	rootCmd.AddCommand(showRatesCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func reportInfectionCmd() *cobra.Command {
	var patientID, patientName, gender string
	var age, departmentID int
	var admissionDate, infectionDate, infectionSite, pathogen string
	var drugSensitivity bool

	cmd := &cobra.Command{
		Use:   "report-infection",
		Short: "报告感染病例",
		Run: func(cmd *cobra.Command, args []string) {
			admission, _ := time.Parse("2006-01-02", admissionDate)
			infection, _ := time.Parse("2006-01-02", infectionDate)

			data := map[string]interface{}{
				"patient_id":       patientID,
				"patient_name":     patientName,
				"gender":           gender,
				"age":              age,
				"department_id":    departmentID,
				"admission_date":   admission,
				"infection_date":   infection,
				"infection_site":   infectionSite,
				"pathogen":         pathogen,
				"drug_sensitivity": drugSensitivity,
			}

			resp, err := postRequest(baseURL+"/infections", data)
			if err != nil {
				fmt.Println("错误:", err)
				return
			}
			fmt.Println("感染病例报告成功:")
			fmt.Println(resp)
		},
	}

	cmd.Flags().StringVar(&patientID, "patient-id", "", "患者住院号")
	cmd.Flags().StringVar(&patientName, "name", "", "患者姓名")
	cmd.Flags().StringVar(&gender, "gender", "", "性别")
	cmd.Flags().IntVar(&age, "age", 0, "年龄")
	cmd.Flags().IntVar(&departmentID, "dept", 0, "科室ID")
	cmd.Flags().StringVar(&admissionDate, "admission", "", "入院日期 (YYYY-MM-DD)")
	cmd.Flags().StringVar(&infectionDate, "infection", "", "感染日期 (YYYY-MM-DD)")
	cmd.Flags().StringVar(&infectionSite, "site", "", "感染部位")
	cmd.Flags().StringVar(&pathogen, "pathogen", "", "病原体")
	cmd.Flags().BoolVar(&drugSensitivity, "sensitivity", false, "药敏试验")

	return cmd
}

func addMeasureCmd() *cobra.Command {
	var caseID, departmentID int
	var measureType, executor, executeDate string

	cmd := &cobra.Command{
		Use:   "add-measure",
		Short: "添加防控措施",
		Run: func(cmd *cobra.Command, args []string) {
			data := map[string]interface{}{
				"infection_case_id": caseID,
				"measure_type":      measureType,
				"department_id":     departmentID,
				"executor":          executor,
			}

			if executeDate != "" {
				date, _ := time.Parse("2006-01-02", executeDate)
				data["execute_date"] = date
			}

			resp, err := postRequest(baseURL+"/measures", data)
			if err != nil {
				fmt.Println("错误:", err)
				return
			}
			fmt.Println("防控措施添加成功:")
			fmt.Println(resp)
		},
	}

	cmd.Flags().IntVar(&caseID, "case-id", 0, "感染病例ID")
	cmd.Flags().StringVar(&measureType, "type", "", "措施类型")
	cmd.Flags().IntVar(&departmentID, "dept", 0, "执行科室ID")
	cmd.Flags().StringVar(&executor, "executor", "", "执行人")
	cmd.Flags().StringVar(&executeDate, "date", "", "执行日期 (YYYY-MM-DD)")

	return cmd
}

func generateReportCmd() *cobra.Command {
	var reportType, month string
	var year int

	cmd := &cobra.Command{
		Use:   "generate-report",
		Short: "生成报告",
		Run: func(cmd *cobra.Command, args []string) {
			data := map[string]interface{}{
				"report_type": reportType,
			}

			if month != "" {
				data["month"] = month
			}
			if year != 0 {
				data["year"] = year
			}

			resp, err := postRequest(baseURL+"/reports/generate", data)
			if err != nil {
				fmt.Println("错误:", err)
				return
			}
			fmt.Println("报告生成成功:")
			fmt.Println(resp)
		},
	}

	cmd.Flags().StringVar(&reportType, "type", "monthly", "报告类型: monthly 或 yearly")
	cmd.Flags().StringVar(&month, "month", "", "月份 (YYYY-MM)")
	cmd.Flags().IntVar(&year, "year", 0, "年份")

	return cmd
}

func showRatesCmd() *cobra.Command {
	var month string

	cmd := &cobra.Command{
		Use:   "show-rates",
		Short: "查看感染率",
		Run: func(cmd *cobra.Command, args []string) {
			url := baseURL + "/statistics/rates"
			if month != "" {
				url += "?month=" + month
			}

			resp, err := getRequest(url)
			if err != nil {
				fmt.Println("错误:", err)
				return
			}

			var result map[string]interface{}
			json.Unmarshal([]byte(resp), &result)

			fmt.Printf("全院感染率: %.2f%%\n", result["hospital_rate"])
			fmt.Println("\n各科室感染率排名:")

			depts := result["department_rates"].([]interface{})
			for i, d := range depts {
				dept := d.(map[string]interface{})
				exceeded := ""
				if dept["exceeded"].(bool) {
					exceeded = " [超标]"
				}
				fmt.Printf("%d. %s: %.2f%% (%d例/%d人)%s\n",
					i+1,
					dept["department_name"],
					dept["infection_rate"],
					int(dept["infection_count"].(float64)),
					int(dept["discharge_count"].(float64)),
					exceeded,
				)
			}
		},
	}

	cmd.Flags().StringVar(&month, "month", "", "月份 (YYYY-MM)")

	return cmd
}

func postRequest(url string, data map[string]interface{}) (string, error) {
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, body, "", "  ")
	return prettyJSON.String(), nil
}

func getRequest(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}

func _unused() {
	_ = strconv.Itoa(0)
}
