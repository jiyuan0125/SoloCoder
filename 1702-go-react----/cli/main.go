package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const serverURL = "http://localhost:8081/api"

type Doctor struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Department string      `json:"department"`
	WorkShifts []WorkShift `json:"work_shifts"`
}

type WorkShift struct {
	StartHour   int `json:"start_hour"`
	StartMinute int `json:"start_minute"`
	EndHour     int `json:"end_hour"`
	EndMinute   int `json:"end_minute"`
}

type Appointment struct {
	ID         string   `json:"id"`
	PatientID  string   `json:"patient_id"`
	DoctorID   string   `json:"doctor_id"`
	Date       string   `json:"date"`
	StartTime  string   `json:"start_time"`
	EndTime    string   `json:"end_time"`
	Treatments []string `json:"treatments"`
}

type Step struct {
	ID           string `json:"id"`
	PlanID       string `json:"plan_id"`
	Index        int    `json:"index"`
	ExpectedDate string `json:"expected_date"`
	ActualDate   string `json:"actual_date"`
	DoctorID     string `json:"doctor_id"`
	Description  string `json:"description"`
	Fee          int    `json:"fee"`
	Status       string `json:"status"`
	IsOverdue    bool   `json:"is_overdue"`
}

type TreatmentPlan struct {
	ID              string `json:"id"`
	PatientID       string `json:"patient_id"`
	DiscountPercent int    `json:"discount_percent"`
	Steps           []Step `json:"steps"`
	Status          string `json:"status"`
	IsSurgical      bool   `json:"is_surgical"`
}

type APIError struct {
	Error string `json:"error"`
}

func makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, serverURL+endpoint, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

func listDoctors() {
	resp, err := makeRequest("GET", "/doctors", nil)
	if err != nil {
		fmt.Printf("错误: 无法连接服务器 - %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		json.Unmarshal(body, &apiErr)
		fmt.Printf("错误: %s\n", apiErr.Error)
		os.Exit(1)
	}

	var doctors []Doctor
	json.Unmarshal(body, &doctors)

	fmt.Println("=== 医生列表 ===")
	for _, d := range doctors {
		fmt.Printf("\nID: %s\n", d.ID)
		fmt.Printf("姓名: %s\n", d.Name)
		fmt.Printf("科室: %s\n", d.Department)
		fmt.Printf("排班: ")
		for i, shift := range d.WorkShifts {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%02d:%02d-%02d:%02d", shift.StartHour, shift.StartMinute, shift.EndHour, shift.EndMinute)
		}
		fmt.Println()
	}
}

func bookAppointment(args []string) {
	if len(args) < 5 {
		fmt.Println("用法: book-appointment <患者ID> <医生ID> <日期(YYYY-MM-DD)> <开始时间(HH:MM)> <治疗项目1,治疗项目2,...>")
		fmt.Println("示例: book-appointment p1 d1 2026-05-13 09:00 洗牙,补牙")
		os.Exit(1)
	}

	patientID := args[0]
	doctorID := args[1]
	date := args[2]
	startTime := args[3]
	treatments := strings.Split(args[4], ",")

	reqBody := map[string]interface{}{
		"patient_id":  patientID,
		"doctor_id":   doctorID,
		"date":        date,
		"start_time":  startTime,
		"treatments":  treatments,
	}

	resp, err := makeRequest("POST", "/appointments", reqBody)
	if err != nil {
		fmt.Printf("错误: 无法连接服务器 - %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusConflict {
		var apiErr APIError
		json.Unmarshal(body, &apiErr)
		fmt.Printf("错误: %s\n", apiErr.Error)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		json.Unmarshal(body, &apiErr)
		fmt.Printf("错误: %s\n", apiErr.Error)
		os.Exit(1)
	}

	var appt Appointment
	json.Unmarshal(body, &appt)
	fmt.Println("预约成功!")
	fmt.Printf("预约ID: %s\n", appt.ID)
	fmt.Printf("患者ID: %s\n", appt.PatientID)
	fmt.Printf("医生ID: %s\n", appt.DoctorID)
	fmt.Printf("日期: %s\n", date)
	fmt.Printf("时间: %s\n", startTime)
	fmt.Printf("治疗项目: %s\n", strings.Join(appt.Treatments, ", "))
}

func listAppointments() {
	resp, err := makeRequest("GET", "/appointments", nil)
	if err != nil {
		fmt.Printf("错误: 无法连接服务器 - %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		json.Unmarshal(body, &apiErr)
		fmt.Printf("错误: %s\n", apiErr.Error)
		os.Exit(1)
	}

	var appointments []Appointment
	json.Unmarshal(body, &appointments)

	fmt.Println("=== 预约列表 ===")
	if len(appointments) == 0 {
		fmt.Println("暂无预约")
		return
	}

	for _, a := range appointments {
		fmt.Printf("\nID: %s\n", a.ID)
		fmt.Printf("患者ID: %s\n", a.PatientID)
		fmt.Printf("医生ID: %s\n", a.DoctorID)
		fmt.Printf("日期: %s\n", a.Date)
		fmt.Printf("时间: %s\n", a.StartTime)
		fmt.Printf("治疗项目: %s\n", strings.Join(a.Treatments, ", "))
	}
}

func completeStep(args []string) {
	if len(args) < 2 {
		fmt.Println("用法: complete-step <治疗计划ID> <步骤ID>")
		fmt.Println("示例: complete-step tp123 s456")
		os.Exit(1)
	}

	planID := args[0]
	stepID := args[1]

	reqBody := map[string]interface{}{
		"step_id": stepID,
		"status":  "in_progress",
	}

	resp, err := makeRequest("PUT", "/treatment-plans/"+planID+"/steps", reqBody)
	if err != nil {
		fmt.Printf("错误: 无法连接服务器 - %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		json.Unmarshal(body, &apiErr)
		fmt.Printf("错误: %s\n", apiErr.Error)
		os.Exit(1)
	}

	reqBody2 := map[string]interface{}{
		"step_id": stepID,
		"status":  "completed",
	}

	resp2, err := makeRequest("PUT", "/treatment-plans/"+planID+"/steps", reqBody2)
	if err != nil {
		fmt.Printf("错误: 无法连接服务器 - %v\n", err)
		os.Exit(1)
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)
	if resp2.StatusCode != http.StatusOK {
		var apiErr APIError
		json.Unmarshal(body2, &apiErr)
		fmt.Printf("错误: %s\n", apiErr.Error)
		os.Exit(1)
	}

	fmt.Println("步骤已标记为完成")
}

func printUsage() {
	fmt.Println("口腔诊所管理系统 CLI")
	fmt.Println("\n可用命令:")
	fmt.Println("  list-doctors                    查看医生列表和排班")
	fmt.Println("  book-appointment                预约挂号")
	fmt.Println("  list-appointments               查看预约列表")
	fmt.Println("  complete-step                   完成治疗步骤")
	fmt.Println("\n示例:")
	fmt.Println("  ./cli list-doctors")
	fmt.Println("  ./cli book-appointment p1 d1 2026-05-13 09:00 洗牙")
	fmt.Println("  ./cli list-appointments")
	fmt.Println("  ./cli complete-step tp123 s456")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "list-doctors":
		listDoctors()
	case "book-appointment":
		bookAppointment(args)
	case "list-appointments":
		listAppointments()
	case "complete-step":
		completeStep(args)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}
