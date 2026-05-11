package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"piperepair/api"
)

const defaultServerURL = "http://localhost:9010"

var serverURL string

func main() {
	flag.StringVar(&serverURL, "server", defaultServerURL, "服务端地址")
	flag.Parse()

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	client := NewAPIClient(serverURL)
	scanner := bufio.NewScanner(os.Stdin)

	command := os.Args[1]

	switch command {
	case "create":
		handleCreateOrder(client, scanner)
	case "get":
		handleGetOrder(client)
	case "list":
		handleListOrders(client)
	case "start":
		handleStartProcessing(client)
	case "submit":
		handleSubmitRepairRecord(client, scanner)
	case "accept":
		handleAcceptOrder(client, scanner)
	case "master":
		handleGetMasterInfo(client)
	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("管道疏通报修管理系统 - 客户端命令行工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client -server=<服务端地址> <命令> [参数]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  create        创建报修单")
	fmt.Println("  get <工单ID>   查询工单详情")
	fmt.Println("  list          列出所有工单")
	fmt.Println("  start <工单ID> <师傅ID>  开始处理工单")
	fmt.Println("  submit        提交维修记录")
	fmt.Println("  accept        验收工单")
	fmt.Println("  master <师傅ID>  查询师傅信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  client create")
	fmt.Println("  client get ORD20240101120000000001")
	fmt.Println("  client list")
	fmt.Println("  client start ORD20240101120000000001 m001")
	fmt.Println("  client submit")
	fmt.Println("  client accept")
	fmt.Println("  client master m001")
	fmt.Println()
	fmt.Println("可用区域代码: 01001-01009")
	fmt.Println("  朝阳区: 01001, 01002, 01003")
	fmt.Println("  海淀区: 01004, 01005, 01006")
	fmt.Println("  丰台区: 01007, 01008, 01009")
	fmt.Println()
	fmt.Println("可用师傅ID:")
	fmt.Println("  m001 - 张师傅 (朝阳区)")
	fmt.Println("  m002 - 李师傅 (朝阳区)")
	fmt.Println("  m003 - 王师傅 (海淀区)")
	fmt.Println("  m004 - 赵师傅 (丰台区)")
	fmt.Println("  m005 - 刘师傅 (丰台区)")
}

func readInput(scanner *bufio.Scanner, prompt string) string {
	fmt.Print(prompt)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func handleCreateOrder(client *APIClient, scanner *bufio.Scanner) {
	fmt.Println("=== 创建报修单 ===")

	address := readInput(scanner, "请输入地址: ")
	contact := readInput(scanner, "请输入联系人姓名: ")
	contactPhone := readInput(scanner, "请输入联系电话: ")

	fmt.Println("\n请选择堵塞类型:")
	fmt.Println("  1. 厨房下水道")
	fmt.Println("  2. 卫生间下水道")
	fmt.Println("  3. 马桶")
	fmt.Println("  4. 地漏")
	fmt.Println("  5. 主管道")
	blockageTypeChoice := readInput(scanner, "请选择 (1-5): ")

	var blockageType api.BlockageType
	switch blockageTypeChoice {
	case "1":
		blockageType = api.BlockageTypeKitchenSink
	case "2":
		blockageType = api.BlockageTypeBathroomSink
	case "3":
		blockageType = api.BlockageTypeToilet
	case "4":
		blockageType = api.BlockageTypeFloorDrain
	case "5":
		blockageType = api.BlockageTypeMainPipe
	default:
		fmt.Println("无效选择")
		return
	}

	fmt.Println("\n请选择堵塞程度:")
	fmt.Println("  1. 轻微流水慢")
	fmt.Println("  2. 严重完全堵塞")
	severityChoice := readInput(scanner, "请选择 (1-2): ")

	var severity api.BlockageSeverity
	switch severityChoice {
	case "1":
		severity = api.SeverityMild
	case "2":
		severity = api.SeveritySevere
	default:
		fmt.Println("无效选择")
		return
	}

	isRecurringStr := readInput(scanner, "是否反复堵塞? (y/n): ")
	isRecurring := strings.ToLower(isRecurringStr) == "y"

	areaCode := readInput(scanner, "请输入区域代码 (01001-01009): ")

	req := &api.CreateRepairOrderRequest{
		Address:          address,
		Contact:          contact,
		ContactPhone:     contactPhone,
		BlockageType:     blockageType,
		BlockageSeverity: severity,
		IsRecurring:      isRecurring,
		AreaCode:         areaCode,
	}

	resp, err := client.CreateOrder(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	printResponse(resp)
}

func handleGetOrder(client *APIClient) {
	if len(os.Args) < 3 {
		fmt.Println("请提供工单ID")
		return
	}
	orderID := os.Args[2]

	resp, err := client.GetOrder(orderID)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	printResponse(resp)
}

func handleListOrders(client *APIClient) {
	resp, err := client.ListOrders()
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	printResponse(resp)
}

func handleStartProcessing(client *APIClient) {
	if len(os.Args) < 4 {
		fmt.Println("请提供工单ID和师傅ID")
		return
	}
	orderID := os.Args[2]
	masterID := os.Args[3]

	resp, err := client.StartProcessing(orderID, masterID)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	printResponse(resp)
}

func handleSubmitRepairRecord(client *APIClient, scanner *bufio.Scanner) {
	fmt.Println("=== 提交维修记录 ===")

	orderID := readInput(scanner, "请输入工单ID: ")
	masterID := readInput(scanner, "请输入师傅ID: ")

	fmt.Println("\n请选择疏通方式:")
	fmt.Println("  1. 手工疏通 (80元)")
	fmt.Println("  2. 管道疏通机 (120元)")
	fmt.Println("  3. 高压清洗 (200元)")
	methodChoice := readInput(scanner, "请选择 (1-3): ")

	var method api.UncloggingMethod
	switch methodChoice {
	case "1":
		method = api.MethodManual
	case "2":
		method = api.MethodMachine
	case "3":
		method = api.MethodHighPressure
	default:
		fmt.Println("无效选择")
		return
	}

	durationStr := readInput(scanner, "请输入疏通时长(分钟): ")
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		fmt.Println("无效的时长")
		return
	}

	replacedPartsStr := readInput(scanner, "是否更换管道配件? (y/n): ")
	replacedParts := strings.ToLower(replacedPartsStr) == "y"

	var partName string
	var partCost float64
	if replacedParts {
		partName = readInput(scanner, "请输入配件名称: ")
		partCostStr := readInput(scanner, "请输入配件费用: ")
		partCost, err = strconv.ParseFloat(partCostStr, 64)
		if err != nil {
			fmt.Println("无效的费用")
			return
		}
	}

	req := &api.SubmitRepairRecordRequest{
		OrderID:          orderID,
		MasterID:         masterID,
		UncloggingMethod: method,
		DurationMinutes:  duration,
		ReplacedParts:    replacedParts,
		PartName:         partName,
		PartCost:         partCost,
	}

	resp, err := client.SubmitRepairRecord(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	printResponse(resp)
}

func handleAcceptOrder(client *APIClient, scanner *bufio.Scanner) {
	fmt.Println("=== 验收工单 ===")

	orderID := readInput(scanner, "请输入工单ID: ")

	resp, err := client.GetOrder(orderID)
	if err != nil {
		fmt.Printf("查询工单失败: %v\n", err)
		return
	}

	fmt.Println("工单详情:")
	printResponse(resp)

	acceptedStr := readInput(scanner, "\n是否确认验收? (y/n): ")
	accepted := strings.ToLower(acceptedStr) == "y"

	var comment string
	if !accepted {
		comment = readInput(scanner, "请输入驳回原因: ")
	}

	req := &api.AcceptOrderRequest{
		OrderID:  orderID,
		Accepted: accepted,
		Comment:  comment,
	}

	resp, err = client.AcceptOrder(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	printResponse(resp)
}

func handleGetMasterInfo(client *APIClient) {
	if len(os.Args) < 3 {
		fmt.Println("请提供师傅ID")
		return
	}
	masterID := os.Args[2]

	resp, err := client.GetMasterInfo(masterID)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	printResponse(resp)
}

func printResponse(resp *api.Response) {
	fmt.Printf("状态码: %d\n", resp.Code)
	fmt.Printf("消息: %s\n", resp.Message)
	if resp.Data != nil {
		jsonData, err := json.MarshalIndent(resp.Data, "", "  ")
		if err == nil {
			fmt.Println("数据:")
			fmt.Println(string(jsonData))
		}
	}
}
