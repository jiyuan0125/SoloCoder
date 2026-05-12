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
		Use:   "hicli",
		Short: "医院感染监测系统CLI",
	}

	rootCmd.AddCommand(reportInfectionCmd())
	rootCmd.AddCommand(addMeasureCmd())
	rootCmd.AddCommand(generateReportCmd())
	rootCmd.AddCommand(showRatesCmd())
	rootCmd.AddCommand(createDeptCmd())
	rootCmd.AddCommand(listDeptsCmd())

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
				"PatientID":       patientID,
				"PatientName":     patientName,
				"Gender":          gender,
				"Age":             age,
				"DepartmentID":    departmentID,
				"AdmissionDate":   admission,
				"InfectionDate":   infection,
				"InfectionSite":   infectionSite,
				"Pathogen":        pathogen,
				"DrugSensitivity": drugSensitivity,
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

	cmd.Flags().StringVar(&patientID, "patient-id", "", "患者住院号 (必填)")
	cmd.Flags().StringVar(&patientName, "name", "", "患者姓名 (必填)")
	cmd.Flags().StringVar(&gender, "gender", "", "性别")
	cmd.Flags().IntVar(&age, "age", 0, "年龄")
	cmd.Flags().IntVar(&departmentID, "dept", 0, "科室ID (必填)")
	cmd.Flags().StringVar(&admissionDate, "admission", "", "入院日期 (YYYY-MM-DD, 必填)")
	cmd.Flags().StringVar(&infectionDate, "infection", "", "感染日期 (YYYY-MM-DD, 必填)")
	cmd.Flags().StringVar(&infectionSite, "site", "", "感染部位: respiratory/surgical_incision/urinary_tract/bloodstream/digestive/skin_soft_tissue (必填)")
	cmd.Flags().StringVar(&pathogen, "pathogen", "", "病原体 (必填)")
	cmd.Flags().BoolVar(&drugSensitivity, "sensitivity", false, "已做药敏试验")

	cmd.MarkFlagRequired("patient-id")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("dept")
	cmd.MarkFlagRequired("admission")
	cmd.MarkFlagRequired("infection")
	cmd.MarkFlagRequired("site")
	cmd.MarkFlagRequired("pathogen")

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
				"InfectionCaseID": caseID,
				"MeasureType":     measureType,
				"DepartmentID":    departmentID,
				"Executor":        executor,
			}

			if executeDate != "" {
				date, _ := time.Parse("2006-01-02", executeDate)
				data["ExecuteDate"] = date
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

	cmd.Flags().IntVar(&caseID, "case-id", 0, "感染病例ID (必填)")
	cmd.Flags().StringVar(&measureType, "type", "", "措施类型: isolation/hand_hygiene/environment_clean/antibiotic_adjust/equipment_sterile (必填)")
	cmd.Flags().IntVar(&departmentID, "dept", 0, "执行科室ID (必填)")
	cmd.Flags().StringVar(&executor, "executor", "", "执行人 (必填)")
	cmd.Flags().StringVar(&executeDate, "date", "", "执行日期 (YYYY-MM-DD)")

	cmd.MarkFlagRequired("case-id")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("dept")
	cmd.MarkFlagRequired("executor")

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
				"ReportType": reportType,
			}

			if month != "" {
				data["Month"] = month
			}
			if year != 0 {
				data["Year"] = year
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
	cmd.Flags().StringVar(&month, "month", "", "月份 (YYYY-MM, 月报必填)")
	cmd.Flags().IntVar(&year, "year", 0, "年份 (年报必填)")

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
				url += "?Month=" + month
			}

			resp, err := getRequest(url)
			if err != nil {
				fmt.Println("错误:", err)
				return
			}

			var result map[string]interface{}
			json.Unmarshal([]byte(resp), &result)

			hospitalRate, _ := result["HospitalRate"].(float64)
			fmt.Printf("全院感染率: %.2f%%\n", hospitalRate)
			fmt.Println("\n各科室感染率排名:")

			depts := result["DepartmentRates"].([]interface{})
			for i, d := range depts {
				dept := d.(map[string]interface{})
				exceeded := ""
				if dept["Exceeded"].(bool) {
					exceeded = " [超标]"
				}
				fmt.Printf("%d. %s: %.2f%% (%d例/%d人)%s\n",
					i+1,
					dept["DepartmentName"],
					dept["InfectionRate"],
					int(dept["InfectionCount"].(float64)),
					int(dept["DischargeCount"].(float64)),
					exceeded,
				)
			}
		},
	}

	cmd.Flags().StringVar(&month, "month", "", "月份 (YYYY-MM)")

	return cmd
}

func createDeptCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "create-dept",
		Short: "创建科室",
		Run: func(cmd *cobra.Command, args []string) {
			data := map[string]interface{}{
				"Name": name,
			}

			resp, err := postRequest(baseURL+"/departments", data)
			if err != nil {
				fmt.Println("错误:", err)
				return
			}
			fmt.Println("科室创建成功:")
			fmt.Println(resp)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "科室名称 (必填)")
	cmd.MarkFlagRequired("name")

	return cmd
}

func listDeptsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-depts",
		Short: "查看所有科室",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := getRequest(baseURL + "/departments")
			if err != nil {
				fmt.Println("错误:", err)
				return
			}

			var depts []map[string]interface{}
			json.Unmarshal([]byte(resp), &depts)

			fmt.Println("科室列表:")
			fmt.Println("ID\t名称")
			fmt.Println("--\t----")
			for _, d := range depts {
				fmt.Printf("%.0f\t%s\n", d["ID"], d["Name"])
			}
		},
	}

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
