package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

const BaseURL = "http://localhost:8080/api"

var rootCmd = &cobra.Command{
	Use:   "epidemic-cli",
	Short: "疫情报告管理系统命令行工具",
}

var reportOutbreakCmd = &cobra.Command{
	Use:   "report-outbreak",
	Short: "上报疫情",
	Run: func(cmd *cobra.Command, args []string) {
		patientName, _ := cmd.Flags().GetString("patient-name")
		patientID, _ := cmd.Flags().GetString("patient-id")
		diseaseName, _ := cmd.Flags().GetString("disease")
		region, _ := cmd.Flags().GetString("region")
		district, _ := cmd.Flags().GetString("district")
		onsetTime, _ := cmd.Flags().GetString("onset-time")

		if onsetTime == "" {
			onsetTime = time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		}

		data := map[string]interface{}{
			"patientName": patientName,
			"patientId":   patientID,
			"diseaseName": diseaseName,
			"region":      region,
			"district":    district,
			"onsetTime":   onsetTime,
			"reportTime":  time.Now().Format(time.RFC3339),
		}

		body, _ := json.Marshal(data)
		resp, err := http.Post(BaseURL+"/outbreak/reports", "application/json", bytes.NewReader(body))
		if err != nil {
			fmt.Println("请求失败:", err)
			return
		}
		defer resp.Body.Close()

		respBody, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(respBody))
	},
}

var createContactCmd = &cobra.Command{
	Use:   "create-contact",
	Short: "添加密接者",
	Run: func(cmd *cobra.Command, args []string) {
		investigationID, _ := cmd.Flags().GetUint("investigation-id")
		name, _ := cmd.Flags().GetString("name")
		phone, _ := cmd.Flags().GetString("phone")
		contactType, _ := cmd.Flags().GetString("contact-type")
		firstContactDate, _ := cmd.Flags().GetString("first-contact")
		lastContactDate, _ := cmd.Flags().GetString("last-contact")
		diseaseName, _ := cmd.Flags().GetString("disease")

		data := map[string]interface{}{
			"investigationId":  investigationID,
			"name":             name,
			"phone":            phone,
			"contactType":      contactType,
			"firstContactDate": firstContactDate,
			"lastContactDate":  lastContactDate,
			"diseaseName":      diseaseName,
		}

		body, _ := json.Marshal(data)
		resp, err := http.Post(BaseURL+"/contacts", "application/json", bytes.NewReader(body))
		if err != nil {
			fmt.Println("请求失败:", err)
			return
		}
		defer resp.Body.Close()

		respBody, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(respBody))
	},
}

var recordVaccineCmd = &cobra.Command{
	Use:   "record-vaccine",
	Short: "记录接种",
	Run: func(cmd *cobra.Command, args []string) {
		recipientName, _ := cmd.Flags().GetString("recipient-name")
		recipientID, _ := cmd.Flags().GetString("recipient-id")
		vaccineName, _ := cmd.Flags().GetString("vaccine")
		doseNumber, _ := cmd.Flags().GetInt("dose")
		vaccinationDate, _ := cmd.Flags().GetString("date")
		unit, _ := cmd.Flags().GetString("unit")
		doctor, _ := cmd.Flags().GetString("doctor")

		if vaccinationDate == "" {
			vaccinationDate = time.Now().Format("2006-01-02")
		}

		data := map[string]interface{}{
			"recipientName":   recipientName,
			"recipientId":     recipientID,
			"vaccineName":     vaccineName,
			"doseNumber":      doseNumber,
			"vaccinationDate": vaccinationDate,
			"unit":            unit,
			"doctor":          doctor,
		}

		body, _ := json.Marshal(data)
		resp, err := http.Post(BaseURL+"/vaccinations", "application/json", bytes.NewReader(body))
		if err != nil {
			fmt.Println("请求失败:", err)
			return
		}
		defer resp.Body.Close()

		respBody, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(respBody))
	},
}

var generateWeeklyCmd = &cobra.Command{
	Use:   "generate-weekly",
	Short: "生成周报",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Post(BaseURL+"/weekly-reports/generate", "application/json", nil)
		if err != nil {
			fmt.Println("请求失败:", err)
			return
		}
		defer resp.Body.Close()

		respBody, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(respBody))
	},
}

func init() {
	reportOutbreakCmd.Flags().String("patient-name", "", "患者姓名")
	reportOutbreakCmd.Flags().String("patient-id", "", "患者ID")
	reportOutbreakCmd.Flags().String("disease", "", "传染病名称")
	reportOutbreakCmd.Flags().String("region", "", "地区")
	reportOutbreakCmd.Flags().String("district", "", "区县")
	reportOutbreakCmd.Flags().String("onset-time", "", "发病时间")

	createContactCmd.Flags().Uint("investigation-id", 0, "调查ID")
	createContactCmd.Flags().String("name", "", "密接者姓名")
	createContactCmd.Flags().String("phone", "", "联系方式")
	createContactCmd.Flags().String("contact-type", "同住", "接触方式")
	createContactCmd.Flags().String("first-contact", "", "首次接触日期")
	createContactCmd.Flags().String("last-contact", "", "最后接触日期")
	createContactCmd.Flags().String("disease", "", "传染病名称")

	recordVaccineCmd.Flags().String("recipient-name", "", "接种者姓名")
	recordVaccineCmd.Flags().String("recipient-id", "", "接种者ID")
	recordVaccineCmd.Flags().String("vaccine", "", "疫苗名称")
	recordVaccineCmd.Flags().Int("dose", 1, "剂次")
	recordVaccineCmd.Flags().String("date", "", "接种日期")
	recordVaccineCmd.Flags().String("unit", "", "接种单位")
	recordVaccineCmd.Flags().String("doctor", "", "接种医生")

	rootCmd.AddCommand(reportOutbreakCmd)
	rootCmd.AddCommand(createContactCmd)
	rootCmd.AddCommand(recordVaccineCmd)
	rootCmd.AddCommand(generateWeeklyCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
