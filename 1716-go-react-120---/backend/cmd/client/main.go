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
)

const baseURL = "http://localhost:8080/api"

type EmergencyCall struct {
	ID             string        `json:"id"`
	CallTime       time.Time     `json:"call_time"`
	CallerName     string        `json:"caller_name"`
	CallerPhone    string        `json:"caller_phone"`
	Location       string        `json:"location"`
	PatientCount   int           `json:"patient_count"`
	PatientGender  string        `json:"patient_gender"`
	PatientAgeGroup string       `json:"patient_age_group"`
	ChiefComplaint string        `json:"chief_complaint"`
	SeverityLevel  string        `json:"severity_level"`
	Status         string        `json:"status"`
}

type Ambulance struct {
	ID              string `json:"id"`
	VehicleNumber   string `json:"vehicle_number"`
	VehicleType     string `json:"vehicle_type"`
	CurrentStatus   string `json:"current_status"`
	CurrentLocation string `json:"current_location"`
	DoctorCount     int    `json:"doctor_count"`
	NurseCount      int    `json:"nurse_count"`
	PatientCount    int    `json:"patient_count"`
}

type Statistics struct {
	TodayCallsCount        int     `json:"today_calls_count"`
	TotalCallsCount        int     `json:"total_calls_count"`
	AverageResponseTime    float64 `json:"average_response_time_minutes"`
	VehicleUtilizationRate float64 `json:"vehicle_utilization_rate"`
	IdleVehiclesCount      int     `json:"idle_vehicles_count"`
	InTransitVehiclesCount int     `json:"in_transit_vehicles_count"`
	WaitingInQueueCount    int     `json:"waiting_in_queue_count"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "receive-call":
		receiveCall(args)
	case "dispatch-vehicle":
		dispatchVehicle(args)
	case "update-status":
		updateStatus(args)
	case "show-stats":
		showStats()
	case "list-vehicles":
		listVehicles()
	case "list-calls":
		listCalls()
	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("急救调度系统 CLI")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client <command> [options]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  receive-call    接入求救电话")
	fmt.Println("  dispatch-vehicle 派车")
	fmt.Println("  update-status   更新车辆状态")
	fmt.Println("  show-stats      查看统计")
	fmt.Println("  list-vehicles   列出所有车辆")
	fmt.Println("  list-calls      列出所有求救")
	fmt.Println()
	fmt.Println("receive-call 参数:")
	fmt.Println("  -name         求救人姓名")
	fmt.Println("  -phone        求救人电话")
	fmt.Println("  -location     求救地点")
	fmt.Println("  -count        患者人数")
	fmt.Println("  -gender       患者性别 (男/女)")
	fmt.Println("  -agegroup     年龄段 (儿童/成人/老人)")
	fmt.Println("  -complaint    主诉症状")
	fmt.Println("  -severity     严重程度 (一级濒危/二级危重/三级急症/四级非急症)")
	fmt.Println()
	fmt.Println("dispatch-vehicle 参数:")
	fmt.Println("  -call-id      求救记录ID")
	fmt.Println("  -vehicle-id   车辆ID (可选)")
	fmt.Println()
	fmt.Println("update-status 参数:")
	fmt.Println("  -vehicle-id   车辆ID")
	fmt.Println("  -status       新状态 (出车中/返回途中/空闲)")
}

func receiveCall(args []string) {
	fs := flag.NewFlagSet("receive-call", flag.ExitOnError)
	name := fs.String("name", "", "求救人姓名")
	phone := fs.String("phone", "", "求救人电话")
	location := fs.String("location", "", "求救地点")
	count := fs.Int("count", 1, "患者人数")
	gender := fs.String("gender", "男", "患者性别")
	ageGroup := fs.String("agegroup", "成人", "年龄段")
	complaint := fs.String("complaint", "", "主诉症状")
	severity := fs.String("severity", "三级急症", "严重程度")

	_ = fs.Parse(args)

	if *name == "" || *phone == "" || *location == "" || *complaint == "" {
		fmt.Println("错误: 姓名、电话、地点、主诉症状为必填项")
		os.Exit(1)
	}

	call := map[string]interface{}{
		"call_time":          time.Now().Format(time.RFC3339),
		"caller_name":        *name,
		"caller_phone":       *phone,
		"location":           *location,
		"patient_count":      *count,
		"patient_gender":     *gender,
		"patient_age_group":  *ageGroup,
		"chief_complaint":    *complaint,
		"severity_level":     *severity,
	}

	body, _ := json.Marshal(call)
	resp, err := http.Post(baseURL+"/calls", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		fmt.Printf("错误 (%d): %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var result EmergencyCall
	_ = json.Unmarshal(respBody, &result)

	fmt.Println("求救记录已创建:")
	fmt.Printf("  ID: %s\n", result.ID)
	fmt.Printf("  求救人: %s\n", result.CallerName)
	fmt.Printf("  电话: %s\n", result.CallerPhone)
	fmt.Printf("  地点: %s\n", result.Location)
	fmt.Printf("  症状: %s\n", result.ChiefComplaint)
	fmt.Printf("  严重程度: %s\n", result.SeverityLevel)
}

func dispatchVehicle(args []string) {
	fs := flag.NewFlagSet("dispatch-vehicle", flag.ExitOnError)
	callID := fs.String("call-id", "", "求救记录ID")
	vehicleID := fs.String("vehicle-id", "", "车辆ID (可选)")

	_ = fs.Parse(args)

	if *callID == "" {
		fmt.Println("错误: 必须指定求救记录ID")
		os.Exit(1)
	}

	dispatch := map[string]string{
		"call_id": *callID,
	}
	if *vehicleID != "" {
		dispatch["vehicle_id"] = *vehicleID
	}

	body, _ := json.Marshal(dispatch)
	resp, err := http.Post(baseURL+"/dispatch", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		fmt.Printf("错误 (%d): %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var result map[string]interface{}
	_ = json.Unmarshal(respBody, &result)

	fmt.Println("派车成功:")
	fmt.Printf("  派车记录ID: %v\n", result["id"])
	fmt.Printf("  目标地点: %v\n", result["target_location"])
	fmt.Printf("  预计到达时间: %.0f分钟\n", result["estimated_arrival_time_minutes"])
	fmt.Printf("  预计距离: %.1f公里\n", result["estimated_distance_km"])
}

func updateStatus(args []string) {
	fs := flag.NewFlagSet("update-status", flag.ExitOnError)
	vehicleID := fs.String("vehicle-id", "", "车辆ID")
	status := fs.String("status", "", "新状态")

	_ = fs.Parse(args)

	if *vehicleID == "" || *status == "" {
		fmt.Println("错误: 必须指定车辆ID和新状态")
		os.Exit(1)
	}

	body, _ := json.Marshal(map[string]string{"status": *status})
	req, _ := http.NewRequest(http.MethodPut, baseURL+"/vehicles/"+*vehicleID+"/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		fmt.Printf("错误 (%d): %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	fmt.Println("车辆状态更新成功")
}

func showStats() {
	resp, err := http.Get(baseURL + "/stats")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var stats Statistics
	_ = json.Unmarshal(respBody, &stats)

	fmt.Println("=== 急救调度系统统计 ===")
	fmt.Printf("今日接警数: %d\n", stats.TodayCallsCount)
	fmt.Printf("总接警数: %d\n", stats.TotalCallsCount)
	fmt.Printf("平均响应时间: %.1f分钟\n", stats.AverageResponseTime)
	fmt.Printf("车辆利用率: %.1f%%\n", stats.VehicleUtilizationRate)
	fmt.Println()
	fmt.Println("车辆状态:")
	fmt.Printf("  空闲: %d\n", stats.IdleVehiclesCount)
	fmt.Printf("  出车中: %d\n", stats.InTransitVehiclesCount)
	fmt.Printf("  等待队列: %d\n", stats.WaitingInQueueCount)
}

func listVehicles() {
	resp, err := http.Get(baseURL + "/vehicles")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var vehicles []Ambulance
	_ = json.Unmarshal(respBody, &vehicles)

	fmt.Println("=== 救护车列表 ===")
	for _, v := range vehicles {
		fmt.Printf("%s (%s) - %s\n", v.VehicleNumber, v.VehicleType, v.CurrentStatus)
		fmt.Printf("  ID: %s\n", v.ID)
		fmt.Printf("  位置: %s\n", v.CurrentLocation)
		fmt.Printf("  医护: %d医 %d护\n", v.DoctorCount, v.NurseCount)
		fmt.Printf("  乘客: %d/3\n", v.PatientCount)
		fmt.Println()
	}
}

func listCalls() {
	resp, err := http.Get(baseURL + "/calls")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(respBody)
	var calls []EmergencyCall
	_ = json.Unmarshal(respBody, &calls)

	fmt.Println("=== 求救记录列表 ===")
	for _, c := range calls {
		fmt.Printf("%s - %s\n", c.ID, c.Status)
		fmt.Printf("  求救人: %s\n", c.CallerName)
		fmt.Printf("  地点: %s\n", c.Location)
		fmt.Printf("  症状: %s\n", c.ChiefComplaint)
		fmt.Printf("  严重程度: %s\n", c.SeverityLevel)
		fmt.Println()
	}
}
