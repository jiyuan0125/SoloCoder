package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	serverURL string
)

type Device struct {
	ID                 uint   `json:"id"`
	AssetNumber        string `json:"asset_number"`
	Name               string `json:"name"`
	BrandModel         string `json:"brand_model"`
	SerialNumber       string `json:"serial_number"`
	Category           string `json:"category"`
	Department         string `json:"department"`
	Location           string `json:"location"`
	PurchaseDate       string `json:"purchase_date"`
	PurchasePrice      float64 `json:"purchase_price"`
	WarrantyExpiryDate string `json:"warranty_expiry_date"`
	Status             string `json:"status"`
}

type WorkOrder struct {
	ID          uint   `json:"id"`
	DeviceID    uint   `json:"device_id"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
	ReportTime  string `json:"report_time"`
}

var rootCmd = &cobra.Command{
	Use:   "mdm-cli",
	Short: "医疗设备管理系统 CLI",
	Long:  "医疗设备管理系统命令行工具，用于设备注册、报修、工单查询和数据导出。",
}

var registerDeviceCmd = &cobra.Command{
	Use:   "register-device",
	Short: "注册新设备",
	Run:   registerDevice,
}

var reportFaultCmd = &cobra.Command{
	Use:   "report-fault",
	Short: "设备故障报修",
	Run:   reportFault,
}

var listWorkordersCmd = &cobra.Command{
	Use:   "list-workorders",
	Short: "查看工单列表",
	Run:   listWorkorders,
}

var exportDevicesCmd = &cobra.Command{
	Use:   "export-devices",
	Short: "导出设备台账为CSV",
	Run:   exportDevices,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:8080", "服务器地址")

	registerDeviceCmd.Flags().String("asset-number", "", "资产编号 (必需)")
	registerDeviceCmd.Flags().String("name", "", "设备名称 (必需)")
	registerDeviceCmd.Flags().String("brand-model", "", "品牌型号 (必需)")
	registerDeviceCmd.Flags().String("serial-number", "", "序列号 (必需)")
	registerDeviceCmd.Flags().String("category", "", "分类 (diagnostic/therapeutic/auxiliary/monitoring) (必需)")
	registerDeviceCmd.Flags().String("department", "", "所属科室 (必需)")
	registerDeviceCmd.Flags().String("location", "", "安装位置 (必需)")
	registerDeviceCmd.Flags().String("purchase-date", "", "购入日期 (YYYY-MM-DD) (必需)")
	registerDeviceCmd.Flags().Float64("purchase-price", 0, "购入价格 (必需)")
	registerDeviceCmd.Flags().String("warranty-expiry", "", "保修截止日期 (YYYY-MM-DD)")

	registerDeviceCmd.MarkFlagRequired("asset-number")
	registerDeviceCmd.MarkFlagRequired("name")
	registerDeviceCmd.MarkFlagRequired("brand-model")
	registerDeviceCmd.MarkFlagRequired("serial-number")
	registerDeviceCmd.MarkFlagRequired("category")
	registerDeviceCmd.MarkFlagRequired("department")
	registerDeviceCmd.MarkFlagRequired("location")
	registerDeviceCmd.MarkFlagRequired("purchase-date")
	registerDeviceCmd.MarkFlagRequired("purchase-price")

	reportFaultCmd.Flags().Uint("device-id", 0, "设备ID (必需)")
	reportFaultCmd.Flags().String("description", "", "故障描述 (必需)")
	reportFaultCmd.Flags().String("priority", "normal", "紧急程度 (normal/urgent/critical)")
	reportFaultCmd.Flags().String("assigned-to", "", "分配给")

	reportFaultCmd.MarkFlagRequired("device-id")
	reportFaultCmd.MarkFlagRequired("description")

	listWorkordersCmd.Flags().String("status", "", "按状态筛选")
	listWorkordersCmd.Flags().String("priority", "", "按优先级筛选")

	exportDevicesCmd.Flags().String("output", "devices.csv", "输出文件路径")

	rootCmd.AddCommand(registerDeviceCmd)
	rootCmd.AddCommand(reportFaultCmd)
	rootCmd.AddCommand(listWorkordersCmd)
	rootCmd.AddCommand(exportDevicesCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func registerDevice(cmd *cobra.Command, args []string) {
	assetNumber, _ := cmd.Flags().GetString("asset-number")
	name, _ := cmd.Flags().GetString("name")
	brandModel, _ := cmd.Flags().GetString("brand-model")
	serialNumber, _ := cmd.Flags().GetString("serial-number")
	category, _ := cmd.Flags().GetString("category")
	department, _ := cmd.Flags().GetString("department")
	location, _ := cmd.Flags().GetString("location")
	purchaseDate, _ := cmd.Flags().GetString("purchase-date")
	purchasePrice, _ := cmd.Flags().GetFloat64("purchase-price")
	warrantyExpiry, _ := cmd.Flags().GetString("warranty-expiry")

	device := map[string]interface{}{
		"asset_number":         assetNumber,
		"name":                 name,
		"brand_model":          brandModel,
		"serial_number":        serialNumber,
		"category":             category,
		"department":           department,
		"location":             location,
		"purchase_date":        purchaseDate,
		"purchase_price":       purchasePrice,
		"warranty_expiry_date": warrantyExpiry,
	}

	body, _ := json.Marshal(device)
	resp, err := http.Post(serverURL+"/api/devices", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusCreated {
		var createdDevice Device
		json.Unmarshal(respBody, &createdDevice)
		fmt.Printf("设备注册成功！ID: %d, 资产编号: %s\n", createdDevice.ID, createdDevice.AssetNumber)
	} else {
		var errorResp map[string]string
		json.Unmarshal(respBody, &errorResp)
		fmt.Printf("注册失败 (HTTP %d): %s\n", resp.StatusCode, errorResp["error"])
	}
}

func reportFault(cmd *cobra.Command, args []string) {
	deviceID, _ := cmd.Flags().GetUint("device-id")
	description, _ := cmd.Flags().GetString("description")
	priority, _ := cmd.Flags().GetString("priority")
	assignedTo, _ := cmd.Flags().GetString("assigned-to")

	workOrder := map[string]interface{}{
		"device_id":   deviceID,
		"description": description,
		"priority":    priority,
		"assigned_to": assignedTo,
	}

	body, _ := json.Marshal(workOrder)
	resp, err := http.Post(serverURL+"/api/workorders", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusCreated {
		var createdOrder WorkOrder
		json.Unmarshal(respBody, &createdOrder)
		fmt.Printf("报修成功！工单ID: %d, 状态: %s\n", createdOrder.ID, createdOrder.Status)
	} else {
		var errorResp map[string]string
		json.Unmarshal(respBody, &errorResp)
		fmt.Printf("报修失败 (HTTP %d): %s\n", resp.StatusCode, errorResp["error"])
	}
}

func listWorkorders(cmd *cobra.Command, args []string) {
	status, _ := cmd.Flags().GetString("status")
	priority, _ := cmd.Flags().GetString("priority")

	url := serverURL + "/api/workorders?"
	params := []string{}
	if status != "" {
		params = append(params, "status="+status)
	}
	if priority != "" {
		params = append(params, "priority="+priority)
	}
	url += strings.Join(params, "&")

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]string
		json.Unmarshal(respBody, &errorResp)
		fmt.Printf("查询失败 (HTTP %d): %s\n", resp.StatusCode, errorResp["error"])
		return
	}

	var workOrders []WorkOrder
	json.Unmarshal(respBody, &workOrders)

	if len(workOrders) == 0 {
		fmt.Println("暂无工单")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\t设备ID\t优先级\t状态\t描述\t报修时间")
	for _, wo := range workOrders {
		desc := wo.Description
		if len(desc) > 30 {
			desc = desc[:30] + "..."
		}
		fmt.Fprintf(w, "%d\t%d\t%s\t%s\t%s\t%s\n",
			wo.ID, wo.DeviceID, wo.Priority, wo.Status, desc, wo.ReportTime[:19])
	}
	w.Flush()
}

func exportDevices(cmd *cobra.Command, args []string) {
	output, _ := cmd.Flags().GetString("output")

	resp, err := http.Get(serverURL + "/api/devices/export")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		var errorResp map[string]string
		json.Unmarshal(respBody, &errorResp)
		fmt.Printf("导出失败 (HTTP %d): %s\n", resp.StatusCode, errorResp["error"])
		return
	}

	body, _ := io.ReadAll(resp.Body)
	file, err := os.Create(output)
	if err != nil {
		fmt.Printf("创建文件失败: %v\n", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(bytes.NewReader(body))
	records, _ := reader.ReadAll()

	writer := csv.NewWriter(file)
	writer.WriteAll(records)
	writer.Flush()

	fmt.Printf("已导出 %d 台设备到 %s\n", len(records)-1, output)
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
