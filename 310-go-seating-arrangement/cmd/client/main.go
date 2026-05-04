package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"seating-arrangement/internal/client"
	"seating-arrangement/pkg/protocol"
)

const (
	serverURL = "http://localhost:8080"
)

var (
	apiClient *client.APIClient
	scanner   *bufio.Scanner
)

func main() {
	apiClient = client.NewAPIClient(serverURL)
	scanner = bufio.NewScanner(os.Stdin)

	fmt.Println("========================================")
	fmt.Println("  场馆座位选座系统客户端")
	fmt.Println("========================================")
	fmt.Println()

	for {
		fmt.Println("\n请选择角色:")
		fmt.Println("1. 管理员")
		fmt.Println("2. 用户")
		fmt.Println("0. 退出")
		fmt.Print("\n请输入选项: ")

		choice := readInt()
		switch choice {
		case 1:
			adminMenu()
		case 2:
			userMenu()
		case 0:
			fmt.Println("再见!")
			return
		default:
			fmt.Println("无效选项，请重新选择。")
		}
	}
}

func readInt() int {
	scanner.Scan()
	input := scanner.Text()
	num, err := strconv.Atoi(input)
	if err != nil {
		return -1
	}
	return num
}

func readString() string {
	scanner.Scan()
	return scanner.Text()
}

func adminMenu() {
	for {
		fmt.Println("\n========== 管理员菜单 ==========")
		fmt.Println("1. 创建场馆")
		fmt.Println("2. 查看场馆列表")
		fmt.Println("3. 创建场次")
		fmt.Println("4. 查看场次列表")
		fmt.Println("5. 查看场次座位统计")
		fmt.Println("6. 导出已售出座位明细")
		fmt.Println("7. 手动释放锁定座位")
		fmt.Println("0. 返回主菜单")
		fmt.Print("\n请输入选项: ")

		choice := readInt()
		switch choice {
		case 1:
			createVenue()
		case 2:
			listVenues()
		case 3:
			createSession()
		case 4:
			listSessions()
		case 5:
			viewSessionStats()
		case 6:
			exportSoldSeats()
		case 7:
			releaseSeat()
		case 0:
			return
		default:
			fmt.Println("无效选项，请重新选择。")
		}
	}
}

func userMenu() {
	for {
		fmt.Println("\n========== 用户菜单 ==========")
		fmt.Println("1. 查看场馆列表")
		fmt.Println("2. 查看场次列表")
		fmt.Println("3. 浏览座位")
		fmt.Println("4. 选座下单")
		fmt.Println("5. 确认支付")
		fmt.Println("6. 取消订单")
		fmt.Println("0. 返回主菜单")
		fmt.Print("\n请输入选项: ")

		choice := readInt()
		switch choice {
		case 1:
			listVenues()
		case 2:
			listSessions()
		case 3:
			browseSeats()
		case 4:
			bookSeat()
		case 5:
			confirmOrder()
		case 6:
			cancelOrder()
		case 0:
			return
		default:
			fmt.Println("无效选项，请重新选择。")
		}
	}
}

func createVenue() {
	fmt.Println("\n--- 创建场馆 ---")

	fmt.Print("请输入场馆名称: ")
	name := readString()
	if name == "" {
		fmt.Println("场馆名称不能为空")
		return
	}

	fmt.Print("请输入区域数量: ")
	sectionCount := readInt()
	if sectionCount <= 0 {
		fmt.Println("区域数量必须大于0")
		return
	}

	var sections []protocol.SectionRequest
	for i := 0; i < sectionCount; i++ {
		fmt.Printf("\n--- 区域 %d ---\n", i+1)

		fmt.Print("区域名称: ")
		sectionName := readString()
		if sectionName == "" {
			fmt.Println("区域名称不能为空")
			return
		}

		fmt.Print("排数: ")
		rows := readInt()
		if rows <= 0 {
			fmt.Println("排数必须大于0")
			return
		}

		fmt.Print("每排座位数: ")
		seatsPerRow := readInt()
		if seatsPerRow <= 0 {
			fmt.Println("每排座位数必须大于0")
			return
		}

		fmt.Print("是否需要设置过道位置? (y/n): ")
		aisleChoice := readString()

		var aisles []string
		if strings.ToLower(aisleChoice) == "y" {
			fmt.Println("过道位置格式: 区域名-排号-座位号 (例如: A区-A-5)")
			fmt.Println("输入多个过道位置用逗号分隔，或直接回车跳过: ")
			aisleInput := readString()
			if aisleInput != "" {
				parts := strings.Split(aisleInput, ",")
				for _, part := range parts {
					aisles = append(aisles, strings.TrimSpace(part))
				}
			}
		}

		sections = append(sections, protocol.SectionRequest{
			Name:        sectionName,
			Rows:        rows,
			SeatsPerRow: seatsPerRow,
			Aisles:      aisles,
		})
	}

	req := &protocol.CreateVenueRequest{
		Name:     name,
		Sections: sections,
	}

	venue, err := apiClient.CreateVenue(req)
	if err != nil {
		fmt.Printf("创建场馆失败: %v\n", err)
		return
	}

	fmt.Printf("\n场馆创建成功!\n")
	fmt.Printf("场馆ID: %s\n", venue.ID)
	fmt.Printf("场馆名称: %s\n", venue.Name)
	fmt.Printf("区域数量: %d\n", len(venue.Sections))
}

func listVenues() {
	venues, err := apiClient.ListVenues()
	if err != nil {
		fmt.Printf("获取场馆列表失败: %v\n", err)
		return
	}

	if len(venues) == 0 {
		fmt.Println("暂无场馆")
		return
	}

	fmt.Println("\n--- 场馆列表 ---")
	for i, venue := range venues {
		fmt.Printf("\n%d. 场馆名称: %s (ID: %s)\n", i+1, venue.Name, venue.ID)
		fmt.Printf("   创建时间: %s\n", venue.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("   区域列表:\n")
		for name, section := range venue.Sections {
			fmt.Printf("     - %s: %d排 x %d座 = %d座位\n",
				name, section.Rows, section.SeatsPerRow, section.TotalSeats)
		}
	}
}

func createSession() {
	fmt.Println("\n--- 创建场次 ---")

	venues, err := apiClient.ListVenues()
	if err != nil {
		fmt.Printf("获取场馆列表失败: %v\n", err)
		return
	}

	if len(venues) == 0 {
		fmt.Println("暂无场馆，请先创建场馆")
		return
	}

	fmt.Println("\n请选择场馆:")
	for i, venue := range venues {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, venue.Name, venue.ID)
	}
	fmt.Print("\n请输入序号: ")

	choice := readInt()
	if choice < 1 || choice > len(venues) {
		fmt.Println("无效选择")
		return
	}

	selectedVenue := venues[choice-1]

	fmt.Print("请输入场次名称 (例如: 2024年演唱会第1场): ")
	sessionName := readString()
	if sessionName == "" {
		fmt.Println("场次名称不能为空")
		return
	}

	req := &protocol.CreateSessionRequest{
		VenueID: selectedVenue.ID,
		Name:    sessionName,
	}

	session, err := apiClient.CreateSession(req)
	if err != nil {
		fmt.Printf("创建场次失败: %v\n", err)
		return
	}

	fmt.Printf("\n场次创建成功!\n")
	fmt.Printf("场次ID: %s\n", session.ID)
	fmt.Printf("场次名称: %s\n", session.Name)
	fmt.Printf("总座位数: %d\n", session.TotalSeats)
}

func listSessions() {
	sessions, err := apiClient.ListSessions()
	if err != nil {
		fmt.Printf("获取场次列表失败: %v\n", err)
		return
	}

	if len(sessions) == 0 {
		fmt.Println("暂无场次")
		return
	}

	fmt.Println("\n--- 场次列表 ---")
	for i, session := range sessions {
		fmt.Printf("\n%d. 场次名称: %s (ID: %s)\n", i+1, session.Name, session.ID)
		fmt.Printf("   关联场馆ID: %s\n", session.VenueID)
		fmt.Printf("   总座位数: %d\n", session.TotalSeats)
		fmt.Printf("   创建时间: %s\n", session.CreatedAt.Format("2006-01-02 15:04:05"))
	}
}

func viewSessionStats() {
	fmt.Println("\n--- 查看场次座位统计 ---")

	sessions, err := apiClient.ListSessions()
	if err != nil {
		fmt.Printf("获取场次列表失败: %v\n", err)
		return
	}

	if len(sessions) == 0 {
		fmt.Println("暂无场次")
		return
	}

	fmt.Println("\n请选择场次:")
	for i, session := range sessions {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, session.Name, session.ID)
	}
	fmt.Print("\n请输入序号: ")

	choice := readInt()
	if choice < 1 || choice > len(sessions) {
		fmt.Println("无效选择")
		return
	}

	selectedSession := sessions[choice-1]

	stats, err := apiClient.GetSessionStats(selectedSession.ID)
	if err != nil {
		fmt.Printf("获取统计信息失败: %v\n", err)
		return
	}

	fmt.Println("\n========== 场次座位统计 ==========")
	fmt.Printf("场次名称: %s\n", stats.SessionName)
	fmt.Printf("总座位数: %d\n", stats.TotalSeats)
	fmt.Println("\n各区域统计:")

	for sectionName, stat := range stats.SectionStats {
		fmt.Printf("\n【%s】\n", sectionName)
		fmt.Printf("  总座位数: %d\n", stat.TotalSeats)
		fmt.Printf("  可用座位: %d\n", stat.AvailableSeats)
		fmt.Printf("  锁定座位: %d\n", stat.LockedSeats)
		fmt.Printf("  已售座位: %d\n", stat.SoldSeats)
	}
}

func exportSoldSeats() {
	fmt.Println("\n--- 导出已售出座位明细 ---")

	sessions, err := apiClient.ListSessions()
	if err != nil {
		fmt.Printf("获取场次列表失败: %v\n", err)
		return
	}

	if len(sessions) == 0 {
		fmt.Println("暂无场次")
		return
	}

	fmt.Println("\n请选择场次:")
	for i, session := range sessions {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, session.Name, session.ID)
	}
	fmt.Print("\n请输入序号: ")

	choice := readInt()
	if choice < 1 || choice > len(sessions) {
		fmt.Println("无效选择")
		return
	}

	selectedSession := sessions[choice-1]

	orders, err := apiClient.GetSoldSeats(selectedSession.ID)
	if err != nil {
		fmt.Printf("获取已售座位明细失败: %v\n", err)
		return
	}

	fmt.Println("\n========== 已售出座位明细 ==========")
	fmt.Printf("场次名称: %s\n", selectedSession.Name)
	fmt.Printf("已售订单数: %d\n\n", len(orders))

	if len(orders) == 0 {
		fmt.Println("暂无已售出座位")
		return
	}

	for i, order := range orders {
		fmt.Printf("%d. 订单ID: %s\n", i+1, order.ID)
		fmt.Printf("   区域: %s\n", order.SectionName)
		fmt.Printf("   座位: %s\n", order.SeatID)
		fmt.Printf("   状态: %s\n", order.Status)
		fmt.Printf("   创建时间: %s\n", order.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}
}

func releaseSeat() {
	fmt.Println("\n--- 手动释放锁定座位 ---")

	sessions, err := apiClient.ListSessions()
	if err != nil {
		fmt.Printf("获取场次列表失败: %v\n", err)
		return
	}

	if len(sessions) == 0 {
		fmt.Println("暂无场次")
		return
	}

	fmt.Println("\n请选择场次:")
	for i, session := range sessions {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, session.Name, session.ID)
	}
	fmt.Print("\n请输入序号: ")

	choice := readInt()
	if choice < 1 || choice > len(sessions) {
		fmt.Println("无效选择")
		return
	}

	selectedSession := sessions[choice-1]

	stats, err := apiClient.GetSessionStats(selectedSession.ID)
	if err != nil {
		fmt.Printf("获取统计信息失败: %v\n", err)
		return
	}

	fmt.Println("\n各区域锁定座位情况:")
	hasLocked := false
	for sectionName, stat := range stats.SectionStats {
		if stat.LockedSeats > 0 {
			fmt.Printf("  %s: %d个锁定座位\n", sectionName, stat.LockedSeats)
			hasLocked = true
		}
	}

	if !hasLocked {
		fmt.Println("当前场次没有锁定的座位")
		return
	}

	fmt.Print("\n请输入要释放的区域名称: ")
	sectionName := readString()

	fmt.Print("请输入要释放的座位ID (格式: 区域-排-号，如 A区-A-5): ")
	seatID := readString()

	req := &protocol.ReleaseSeatRequest{
		SessionID:   selectedSession.ID,
		SectionName: sectionName,
		SeatID:      seatID,
	}

	if err := apiClient.ReleaseSeat(req); err != nil {
		fmt.Printf("释放座位失败: %v\n", err)
		return
	}

	fmt.Println("座位释放成功!")
}

func browseSeats() {
	fmt.Println("\n--- 浏览座位 ---")

	sessions, err := apiClient.ListSessions()
	if err != nil {
		fmt.Printf("获取场次列表失败: %v\n", err)
		return
	}

	if len(sessions) == 0 {
		fmt.Println("暂无场次")
		return
	}

	fmt.Println("\n请选择场次:")
	for i, session := range sessions {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, session.Name, session.ID)
	}
	fmt.Print("\n请输入序号: ")

	choice := readInt()
	if choice < 1 || choice > len(sessions) {
		fmt.Println("无效选择")
		return
	}

	selectedSession := sessions[choice-1]

	stats, err := apiClient.GetSessionStats(selectedSession.ID)
	if err != nil {
		fmt.Printf("获取统计信息失败: %v\n", err)
		return
	}

	fmt.Println("\n区域列表:")
	sectionNames := make([]string, 0, len(stats.SectionStats))
	for sectionName := range stats.SectionStats {
		sectionNames = append(sectionNames, sectionName)
		fmt.Printf("  - %s\n", sectionName)
	}

	fmt.Print("\n请输入要浏览的区域名称: ")
	sectionName := readString()

	seats, err := apiClient.GetAvailableSeats(selectedSession.ID, sectionName)
	if err != nil {
		fmt.Printf("获取座位列表失败: %v\n", err)
		return
	}

	fmt.Printf("\n========== %s 座位列表 ==========\n", sectionName)
	fmt.Printf("说明: [O]可用  [L]锁定  [S]已售\n\n")

	if len(seats) == 0 {
		fmt.Println("该区域没有可用座位")
		return
	}

	seatMap := make(map[string][]*protocol.SeatDetail)
	for _, seat := range seats {
		seatMap[seat.Row] = append(seatMap[seat.Row], seat)
	}

	for row, rowSeats := range seatMap {
		fmt.Printf("%s排: ", row)
		for i := 0; i < len(rowSeats); i++ {
			seat := rowSeats[i]
			var statusSymbol string
			switch seat.Status {
			case protocol.SeatStatusAvailable:
				statusSymbol = "O"
			case protocol.SeatStatusLocked:
				statusSymbol = "L"
			case protocol.SeatStatusSold:
				statusSymbol = "S"
			default:
				statusSymbol = "?"
			}
			fmt.Printf("[%s-%d:%s] ", seat.SectionName, seat.Number, statusSymbol)
		}
		fmt.Println()
	}

	fmt.Println("\n座位ID格式: 区域-排-号 (例如: A区-A-5)")
}

func bookSeat() {
	fmt.Println("\n--- 选座下单 ---")

	sessions, err := apiClient.ListSessions()
	if err != nil {
		fmt.Printf("获取场次列表失败: %v\n", err)
		return
	}

	if len(sessions) == 0 {
		fmt.Println("暂无场次")
		return
	}

	fmt.Println("\n请选择场次:")
	for i, session := range sessions {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, session.Name, session.ID)
	}
	fmt.Print("\n请输入序号: ")

	choice := readInt()
	if choice < 1 || choice > len(sessions) {
		fmt.Println("无效选择")
		return
	}

	selectedSession := sessions[choice-1]

	fmt.Print("请输入区域名称: ")
	sectionName := readString()

	fmt.Print("请输入座位ID (格式: 区域-排-号，如 A区-A-5): ")
	seatID := readString()

	req := &protocol.LockSeatRequest{
		SessionID:   selectedSession.ID,
		SectionName: sectionName,
		SeatID:      seatID,
	}

	order, err := apiClient.LockSeat(req)
	if err != nil {
		fmt.Printf("锁定座位失败: %v\n", err)
		return
	}

	fmt.Println("\n========== 订单信息 ==========")
	fmt.Printf("订单ID: %s\n", order.ID)
	fmt.Printf("场次ID: %s\n", order.SessionID)
	fmt.Printf("区域: %s\n", order.SectionName)
	fmt.Printf("座位: %s\n", order.SeatID)
	fmt.Printf("状态: %s\n", order.Status)
	fmt.Printf("创建时间: %s\n", order.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println("\n重要提示: 座位已锁定10分钟，请在锁定时间内完成支付!")
}

func confirmOrder() {
	fmt.Println("\n--- 确认支付 ---")

	fmt.Print("请输入订单ID: ")
	orderID := readString()

	req := &protocol.ConfirmOrderRequest{
		OrderID: orderID,
	}

	if err := apiClient.ConfirmOrder(req); err != nil {
		fmt.Printf("确认支付失败: %v\n", err)
		return
	}

	fmt.Println("支付确认成功! 座位已售出。")
}

func cancelOrder() {
	fmt.Println("\n--- 取消订单 ---")

	fmt.Print("请输入订单ID: ")
	orderID := readString()

	req := &protocol.CancelOrderRequest{
		OrderID: orderID,
	}

	if err := apiClient.CancelOrder(req); err != nil {
		fmt.Printf("取消订单失败: %v\n", err)
		return
	}

	fmt.Println("订单取消成功! 座位已释放。")
}
