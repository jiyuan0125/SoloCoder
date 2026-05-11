package main

import (
	"bytes"
	"charging-station/common"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultServerURL = "http://localhost:8906"
)

type APIClient struct {
	baseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	if baseURL == "" {
		baseURL = defaultServerURL
	}
	return &APIClient{baseURL: baseURL}
}

func (c *APIClient) doRequest(method, path string, body interface{}) (*common.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp common.Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return &apiResp, nil
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func formatAmount(amount int64) string {
	return fmt.Sprintf("¥%.2f", float64(amount)/100.0)
}

func (c *APIClient) ListStations() error {
	resp, err := c.doRequest("GET", "/api/stations", nil)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return err
	}

	var stationsResp common.ListStationsResponse
	if err := json.Unmarshal(dataBytes, &stationsResp); err != nil {
		return err
	}

	fmt.Println("\n=== 充电站列表 ===")
	for _, station := range stationsResp.Stations {
		fmt.Printf("\n站点: %s (%s)\n", station.Name, station.ID)
		fmt.Printf("位置: %s\n", station.Location)
		fmt.Println("充电桩:")
		
		for _, charger := range station.Chargers {
			typeStr := "慢充"
			if charger.Type == "fast" {
				typeStr = "快充"
			}
			
			statusStr := map[string]string{
				"idle":      "空闲",
				"charging":  "充电中",
				"paused":    "暂停中",
				"fault":     "故障",
				"offline":   "离线",
				"reserved":  "已预约",
			}[charger.Status]
			
			fmt.Printf("  - %s [%s] - %s - 电表: %.2f kWh\n", 
				charger.Code, typeStr, statusStr, charger.MeterReading)
		}
	}
	
	return nil
}

func (c *APIClient) CreateReservation(userID, stationID, chargerID string, startTime, endTime time.Time) error {
	req := common.CreateReservationRequest{
		UserID:    userID,
		StationID: stationID,
		ChargerID: chargerID,
		StartTime: startTime,
		EndTime:   endTime,
	}

	resp, err := c.doRequest("POST", "/api/reservations", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return err
	}

	var reservationResp common.CreateReservationResponse
	if err := json.Unmarshal(dataBytes, &reservationResp); err != nil {
		return err
	}

	fmt.Println("\n=== 预约成功 ===")
	fmt.Printf("预约ID: %s\n", reservationResp.Reservation.ID)
	fmt.Printf("用户ID: %s\n", reservationResp.Reservation.UserID)
	fmt.Printf("站点ID: %s\n", reservationResp.Reservation.StationID)
	fmt.Printf("充电桩ID: %s\n", reservationResp.Reservation.ChargerID)
	fmt.Printf("开始时间: %s\n", reservationResp.Reservation.StartTime.Format("2006-01-02 15:04"))
	fmt.Printf("结束时间: %s\n", reservationResp.Reservation.EndTime.Format("2006-01-02 15:04"))
	fmt.Printf("状态: %s\n", reservationResp.Reservation.Status)

	return nil
}

func (c *APIClient) StartCharging(reservationID string) error {
	req := common.StartChargingRequest{
		ReservationID: reservationID,
	}

	resp, err := c.doRequest("POST", "/api/charging/start", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return err
	}

	var sessionResp common.StartChargingResponse
	if err := json.Unmarshal(dataBytes, &sessionResp); err != nil {
		return err
	}

	fmt.Println("\n=== 开始充电 ===")
	fmt.Printf("会话ID: %s\n", sessionResp.Session.ID)
	fmt.Printf("预约ID: %s\n", sessionResp.Session.ReservationID)
	fmt.Printf("开始时间: %s\n", sessionResp.Session.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("起始电表: %.2f kWh\n", sessionResp.Session.StartMeter)

	return nil
}

func (c *APIClient) PauseCharging(sessionID string) error {
	req := common.PauseChargingRequest{
		SessionID: sessionID,
	}

	resp, err := c.doRequest("POST", "/api/charging/pause", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	fmt.Println("\n=== 暂停充电成功 ===")
	return nil
}

func (c *APIClient) ResumeCharging(sessionID string) error {
	req := common.ResumeChargingRequest{
		SessionID: sessionID,
	}

	resp, err := c.doRequest("POST", "/api/charging/resume", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	fmt.Println("\n=== 恢复充电成功 ===")
	return nil
}

func (c *APIClient) EndCharging(sessionID string, endMeter float64) error {
	req := common.EndChargingRequest{
		SessionID: sessionID,
		EndMeter:  endMeter,
	}

	resp, err := c.doRequest("POST", "/api/charging/end", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return err
	}

	var sessionResp common.EndChargingResponse
	if err := json.Unmarshal(dataBytes, &sessionResp); err != nil {
		return err
	}

	fmt.Println("\n=== 结束充电 ===")
	fmt.Printf("会话ID: %s\n", sessionResp.Session.ID)
	fmt.Printf("结束时间: %s\n", sessionResp.Session.EndTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("用电量: %.2f kWh\n", sessionResp.Session.EnergyUsed)
	fmt.Printf("费用: %s\n", formatAmount(sessionResp.Session.Amount))

	return nil
}

func (c *APIClient) ListChargingRecords(userID string) error {
	resp, err := c.doRequest("GET", "/api/records?user_id="+userID, nil)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return err
	}

	var recordsResp common.ListChargingRecordsResponse
	if err := json.Unmarshal(dataBytes, &recordsResp); err != nil {
		return err
	}

	if len(recordsResp.Records) == 0 {
		fmt.Println("\n暂无充电记录")
		return nil
	}

	fmt.Println("\n=== 充电记录 ===")
	for i, record := range recordsResp.Records {
		typeStr := "慢充"
		if record.ChargingType == "fast" {
			typeStr = "快充"
		}

		fmt.Printf("\n记录 #%d\n", i+1)
		fmt.Printf("  月份: %s\n", record.Month)
		fmt.Printf("  开始: %s\n", record.StartTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("  结束: %s\n", record.EndTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("  时长: %s\n", formatDuration(record.Duration))
		fmt.Printf("  类型: %s\n", typeStr)
		fmt.Printf("  电量: %.2f kWh\n", record.EnergyUsed)
		fmt.Printf("  费用: %s\n", formatAmount(record.Amount))
	}

	return nil
}

func (c *APIClient) GetMonthlySummary(userID string) error {
	resp, err := c.doRequest("GET", "/api/summary?user_id="+userID, nil)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return err
	}

	var summaryResp common.GetMonthlySummaryResponse
	if err := json.Unmarshal(dataBytes, &summaryResp); err != nil {
		return err
	}

	if len(summaryResp.Summaries) == 0 {
		fmt.Println("\n暂无月度统计")
		return nil
	}

	fmt.Println("\n=== 月度统计 ===")
	for _, summary := range summaryResp.Summaries {
		fmt.Printf("\n月份: %s\n", summary.Month)
		fmt.Printf("  充电次数: %d\n", summary.TotalSessions)
		fmt.Printf("  总电量: %.2f kWh\n", summary.TotalEnergy)
		fmt.Printf("  总费用: %s\n", formatAmount(summary.TotalAmount))
	}

	return nil
}

func printUsage() {
	fmt.Println("充电站预约管理系统 - 命令行客户端")
	fmt.Println("\n用法:")
	fmt.Println("  client [options] <command> [args]")
	fmt.Println("\n命令:")
	fmt.Println("  stations                           列出所有充电站和充电桩")
	fmt.Println("  reserve <user> <station> <charger> <start> <end>  预约充电桩")
	fmt.Println("  start <reservation_id>             开始充电")
	fmt.Println("  pause <session_id>                 暂停充电")
	fmt.Println("  resume <session_id>                恢复充电")
	fmt.Println("  end <session_id> <end_meter>       结束充电")
	fmt.Println("  records <user_id>                  查看充电记录")
	fmt.Println("  summary <user_id>                  查看月度统计")
	fmt.Println("\n选项:")
	fmt.Println("  --server URL                       服务端URL (默认: http://localhost:8906)")
	fmt.Println("\n示例:")
	fmt.Println("  client stations")
	fmt.Println("  client reserve user-001 station-001 charger-001 14:00 15:00")
	fmt.Println("  client reserve user-001 station-001 charger-001 2026-05-12T14:00:00 2026-05-12T15:00:00")
}

func parseTime(timeStr string, baseDate time.Time) (time.Time, error) {
	if strings.Contains(timeStr, "T") {
		return time.Parse(time.RFC3339, timeStr)
	}

	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("无效的时间格式: %s", timeStr)
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("无效的小时: %s", parts[0])
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("无效的分钟: %s", parts[1])
	}

	now := time.Now()
	result := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.Local)
	
	if result.Before(now) {
		result = result.Add(24 * time.Hour)
	}

	return result, nil
}

func main() {
	serverURL := flag.String("server", "", "服务端URL")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(0)
	}

	client := NewAPIClient(*serverURL)
	command := args[0]

	var err error
	switch command {
	case "stations":
		err = client.ListStations()

	case "reserve":
		if len(args) != 6 {
			fmt.Println("用法: client reserve <user> <station> <charger> <start> <end>")
			os.Exit(1)
		}

		userID := args[1]
		stationID := args[2]
		chargerID := args[3]
		startTimeStr := args[4]
		endTimeStr := args[5]

		startTime, err2 := parseTime(startTimeStr, time.Now())
		if err2 != nil {
			fmt.Printf("无效的开始时间: %v\n", err2)
			os.Exit(1)
		}

		endTime, err2 := parseTime(endTimeStr, startTime)
		if err2 != nil {
			fmt.Printf("无效的结束时间: %v\n", err2)
			os.Exit(1)
		}

		err = client.CreateReservation(userID, stationID, chargerID, startTime, endTime)

	case "start":
		if len(args) != 2 {
			fmt.Println("用法: client start <reservation_id>")
			os.Exit(1)
		}
		err = client.StartCharging(args[1])

	case "pause":
		if len(args) != 2 {
			fmt.Println("用法: client pause <session_id>")
			os.Exit(1)
		}
		err = client.PauseCharging(args[1])

	case "resume":
		if len(args) != 2 {
			fmt.Println("用法: client resume <session_id>")
			os.Exit(1)
		}
		err = client.ResumeCharging(args[1])

	case "end":
		if len(args) != 3 {
			fmt.Println("用法: client end <session_id> <end_meter>")
			os.Exit(1)
		}
		endMeter, err2 := strconv.ParseFloat(args[2], 64)
		if err2 != nil {
			fmt.Printf("无效的电表读数: %v\n", err2)
			os.Exit(1)
		}
		err = client.EndCharging(args[1], endMeter)

	case "records":
		if len(args) != 2 {
			fmt.Println("用法: client records <user_id>")
			os.Exit(1)
		}
		err = client.ListChargingRecords(args[1])

	case "summary":
		if len(args) != 2 {
			fmt.Println("用法: client summary <user_id>")
			os.Exit(1)
		}
		err = client.GetMonthlySummary(args[1])

	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("\n错误: %v\n", err)
		os.Exit(1)
	}
}
