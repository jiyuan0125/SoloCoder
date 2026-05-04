package client

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"luggage-tracking/common"
)

type CLI struct {
	apiClient *APIClient
	reader    *bufio.Reader
}

func NewCLI(serverURL string) *CLI {
	return &CLI{
		apiClient: NewAPIClient(serverURL),
		reader:    bufio.NewReader(os.Stdin),
	}
}

func (c *CLI) Run() {
	fmt.Println("=======================================")
	fmt.Println("   航空公司行李追踪系统 - 客户端")
	fmt.Println("=======================================")
	fmt.Println()

	for {
		c.showMainMenu()
		choice := c.prompt("请选择功能: ")

		switch choice {
		case "1":
			c.staffMenu()
		case "2":
			c.travelerMenu()
		case "3":
			c.adminMenu()
		case "q", "Q":
			fmt.Println("再见！")
			return
		default:
			fmt.Println("无效的选择，请重新输入。")
		}
	}
}

func (c *CLI) showMainMenu() {
	fmt.Println("\n--- 主菜单 ---")
	fmt.Println("1. 工作人员模式")
	fmt.Println("2. 旅客模式")
	fmt.Println("3. 管理员模式")
	fmt.Println("q. 退出")
	fmt.Println()
}

func (c *CLI) staffMenu() {
	for {
		fmt.Println("\n--- 工作人员模式 ---")
		fmt.Println("1. 扫码上报行李")
		fmt.Println("2. 补录扫码节点")
		fmt.Println("3. 查询航班行李")
		fmt.Println("4. 取消航班")
		fmt.Println("b. 返回主菜单")
		fmt.Println()

		choice := c.prompt("请选择功能: ")

		switch choice {
		case "1":
			c.scanLuggage()
		case "2":
			c.backfillScan()
		case "3":
			c.queryFlight()
		case "4":
			c.cancelFlight()
		case "b", "B":
			return
		default:
			fmt.Println("无效的选择，请重新输入。")
		}
	}
}

func (c *CLI) travelerMenu() {
	for {
		fmt.Println("\n--- 旅客模式 ---")
		fmt.Println("1. 查询行李进度")
		fmt.Println("b. 返回主菜单")
		fmt.Println()

		choice := c.prompt("请选择功能: ")

		switch choice {
		case "1":
			c.queryLuggage()
		case "b", "B":
			return
		default:
			fmt.Println("无效的选择，请重新输入。")
		}
	}
}

func (c *CLI) adminMenu() {
	for {
		fmt.Println("\n--- 管理员模式 ---")
		fmt.Println("1. 查看操作日志")
		fmt.Println("2. 查看审计日志")
		fmt.Println("3. 检查服务状态")
		fmt.Println("b. 返回主菜单")
		fmt.Println()

		choice := c.prompt("请选择功能: ")

		switch choice {
		case "1":
			c.showOperationLogs()
		case "2":
			c.showAuditLogs()
		case "3":
			c.checkServerStatus()
		case "b", "B":
			return
		default:
			fmt.Println("无效的选择，请重新输入。")
		}
	}
}

func (c *CLI) scanLuggage() {
	fmt.Println("\n--- 扫码上报 ---")
	fmt.Println("可用环节: check_in(值机), security(安检), sorting(分拣), loading(装机), arrival(到达), conveyor(传送带)")
	fmt.Println()

	tagNumber := c.prompt("请输入行李牌号 (10位数字字母): ")
	if !common.ValidateLuggageTag(tagNumber) {
		fmt.Println("错误: 行李牌号格式不正确，必须是10位数字或字母")
		return
	}

	stage := c.prompt("请输入环节代码: ")
	if !common.ValidateStage(stage) {
		fmt.Println("错误: 无效的环节代码")
		return
	}

	operator := c.prompt("请输入操作员姓名: ")
	if operator == "" {
		fmt.Println("错误: 操作员姓名不能为空")
		return
	}

	req := common.ScanRequest{
		TagNumber: tagNumber,
		Stage:     stage,
		Operator:  operator,
	}

	if stage == common.StageCheckIn {
		flightNumber := c.prompt("请输入航班号: ")
		if flightNumber == "" {
			fmt.Println("错误: 值机环节必须提供航班号")
			return
		}
		req.FlightNumber = flightNumber
	}

	err := c.apiClient.Scan(req)
	if err != nil {
		fmt.Printf("上报失败: %v\n", err)
		return
	}

	fmt.Println("✓ 上报成功！")
	fmt.Printf("行李 %s 已在 %s 环节扫码\n", tagNumber, common.StageName[stage])
}

func (c *CLI) backfillScan() {
	fmt.Println("\n--- 补录扫码 ---")

	tagNumber := c.prompt("请输入行李牌号 (10位数字字母): ")
	if !common.ValidateLuggageTag(tagNumber) {
		fmt.Println("错误: 行李牌号格式不正确")
		return
	}

	stage := c.prompt("请输入补录环节代码: ")
	if !common.ValidateStage(stage) {
		fmt.Println("错误: 无效的环节代码")
		return
	}

	operator := c.prompt("请输入操作员姓名: ")
	if operator == "" {
		fmt.Println("错误: 操作员姓名不能为空")
		return
	}

	req := common.BackfillRequest{
		TagNumber: tagNumber,
		Stage:     stage,
		Operator:  operator,
	}

	scannedAtStr := c.prompt("请输入扫码时间 (可选, 格式: 2006-01-02 15:04:05, 留空使用当前时间): ")
	if scannedAtStr != "" {
		scannedAt, err := time.ParseInLocation("2006-01-02 15:04:05", scannedAtStr, time.Local)
		if err != nil {
			fmt.Printf("时间格式错误: %v\n", err)
			return
		}
		req.ScannedAt = scannedAt
	}

	err := c.apiClient.Backfill(req)
	if err != nil {
		fmt.Printf("补录失败: %v\n", err)
		return
	}

	fmt.Println("✓ 补录成功！")
	fmt.Printf("行李 %s 已补录 %s 环节\n", tagNumber, common.StageName[stage])
}

func (c *CLI) queryLuggage() {
	fmt.Println("\n--- 查询行李进度 ---")

	tagNumber := c.prompt("请输入行李牌号 (10位数字字母): ")
	if !common.ValidateLuggageTag(tagNumber) {
		fmt.Println("错误: 行李牌号格式不正确")
		return
	}

	luggage, err := c.apiClient.GetLuggage(tagNumber)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	c.printLuggageInfo(luggage)
}

func (c *CLI) queryFlight() {
	fmt.Println("\n--- 查询航班行李 ---")

	flightNumber := c.prompt("请输入航班号: ")
	if flightNumber == "" {
		fmt.Println("错误: 航班号不能为空")
		return
	}

	date := c.prompt("请输入日期 (可选, 格式: 2006-01-02, 留空查询当天): ")

	response, err := c.apiClient.GetFlightLuggages(flightNumber, date)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	fmt.Printf("\n航班 %s (%s) 行李信息:\n", response.FlightNumber, response.Date)
	fmt.Printf("共 %d 件行李\n\n", len(response.Luggages))

	for i, luggage := range response.Luggages {
		fmt.Printf("--- 行李 %d ---\n", i+1)
		c.printLuggageInfo(&luggage)
	}
}

func (c *CLI) cancelFlight() {
	fmt.Println("\n--- 取消航班 ---")

	flightNumber := c.prompt("请输入航班号: ")
	if flightNumber == "" {
		fmt.Println("错误: 航班号不能为空")
		return
	}

	date := c.prompt("请输入日期 (可选, 格式: 2006-01-02, 留空使用当天): ")
	operator := c.prompt("请输入操作员姓名: ")
	if operator == "" {
		fmt.Println("错误: 操作员姓名不能为空")
		return
	}

	confirm := c.prompt(fmt.Sprintf("确认取消航班 %s? (y/n): ", flightNumber))
	if confirm != "y" && confirm != "Y" {
		fmt.Println("操作已取消")
		return
	}

	req := common.FlightCancelRequest{
		FlightNumber: flightNumber,
		Date:         date,
		Operator:     operator,
	}

	err := c.apiClient.CancelFlight(req)
	if err != nil {
		fmt.Printf("取消航班失败: %v\n", err)
		return
	}

	fmt.Println("✓ 航班已取消，相关行李已标记为'航班取消'状态")
}

func (c *CLI) printLuggageInfo(luggage *common.LuggageResponse) {
	fmt.Printf("行李牌号: %s\n", luggage.TagNumber)
	fmt.Printf("航班号: %s\n", luggage.FlightNumber)
	fmt.Printf("当前状态: %s\n", c.getStatusDisplay(luggage.Status))
	fmt.Printf("当前环节: %s\n", luggage.CurrentStage)
	fmt.Printf("创建时间: %s\n", luggage.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Println("轨迹记录:")
	for i, stage := range luggage.Stages {
		backfill := ""
		if stage.IsBackfilled {
			backfill = " [补录]"
		}
		fmt.Printf("  %d. %s - %s (操作员: %s)%s\n",
			i+1,
			common.StageName[stage.Stage],
			stage.ScannedAt.Format("2006-01-02 15:04:05"),
			stage.Operator,
			backfill,
		)
	}
	fmt.Println()
}

func (c *CLI) getStatusDisplay(status string) string {
	switch status {
	case common.StatusNormal:
		return "正常"
	case common.StatusAnomaly:
		return "异常 (停留超时)"
	case common.StatusFlightCancel:
		return "航班取消"
	case common.StatusCompleted:
		return "已完成"
	default:
		return status
	}
}

func (c *CLI) showOperationLogs() {
	logs, err := c.apiClient.GetOperationLogs()
	if err != nil {
		fmt.Printf("获取操作日志失败: %v\n", err)
		return
	}

	fmt.Println("\n--- 操作日志 ---")
	if len(logs) == 0 {
		fmt.Println("暂无操作记录")
		return
	}

	for _, log := range logs {
		fmt.Printf("[%s] %s (%s) - %s: %s\n",
			log.CreatedAt.Format("2006-01-02 15:04:05"),
			log.Operator,
			log.Role,
			log.Action,
			log.Details,
		)
	}
}

func (c *CLI) showAuditLogs() {
	logs, err := c.apiClient.GetAuditLogs()
	if err != nil {
		fmt.Printf("获取审计日志失败: %v\n", err)
		return
	}

	fmt.Println("\n--- 审计日志 ---")
	if len(logs) == 0 {
		fmt.Println("暂无审计记录")
		return
	}

	for _, log := range logs {
		immutable := ""
		if log.Immutable {
			immutable = " [不可修改]"
		}
		fmt.Printf("[%s] %s - %s: %s%s\n",
			log.CreatedAt.Format("2006-01-02 15:04:05"),
			log.Operator,
			log.Action,
			log.Details,
			immutable,
		)
	}
}

func (c *CLI) checkServerStatus() {
	ok, err := c.apiClient.Health()
	if err != nil {
		fmt.Printf("服务连接失败: %v\n", err)
		return
	}

	if ok {
		fmt.Println("✓ 服务运行正常")
	} else {
		fmt.Println("✗ 服务异常")
	}
}

func (c *CLI) prompt(msg string) string {
	fmt.Print(msg)
	input, _ := c.reader.ReadString('\n')
	return strings.TrimSpace(input)
}
