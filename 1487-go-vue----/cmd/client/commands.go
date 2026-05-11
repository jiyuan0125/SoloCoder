package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"moving-platform/pkg/common"
	"moving-platform/pkg/core"
)

func cmdEstimate(c *Client, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== 估价查询 ===")

	distance, err := readFloat(reader, "请输入距离（公里）: ")
	if err != nil {
		return err
	}

	items, err := readItems(reader)
	if err != nil {
		return err
	}

	estimate, err := c.Estimate(items, distance)
	if err != nil {
		return err
	}

	fmt.Printf("\n估价: %s\n", formatCents(estimate))
	return nil
}

func cmdBook(c *Client, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== 创建预约 ===")

	dateStr, err := readString(reader, "请输入搬家日期 (YYYY-MM-DD): ")
	if err != nil {
		return err
	}
	moveDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("无效的日期格式: %w", err)
	}

	slots, err := c.CheckAvailability(moveDate)
	if err != nil {
		return err
	}

	if len(slots) == 0 {
		return fmt.Errorf("该日期没有可用时段")
	}

	fmt.Printf("\n可用时段: %v\n", slots)
	slotStr, err := readString(reader, "请选择时段 (morning/afternoon/evening): ")
	if err != nil {
		return err
	}
	timeSlot := core.TimeSlot(slotStr)

	customerName, err := readString(reader, "请输入客户姓名: ")
	if err != nil {
		return err
	}

	phone, err := readString(reader, "请输入联系电话: ")
	if err != nil {
		return err
	}

	fromAddr, err := readString(reader, "请输入起始地址: ")
	if err != nil {
		return err
	}

	toAddr, err := readString(reader, "请输入目标地址: ")
	if err != nil {
		return err
	}

	distance, err := readFloat(reader, "请输入距离（公里）: ")
	if err != nil {
		return err
	}

	items, err := readItems(reader)
	if err != nil {
		return err
	}

	req := common.CreateBookingRequest{
		CustomerName: customerName,
		Phone:        phone,
		MoveDate:     moveDate,
		TimeSlot:     timeSlot,
		FromAddress:  fromAddr,
		ToAddress:    toAddr,
		DistanceKM:   distance,
		Items:        items,
	}

	bookingID, estimate, err := c.CreateBooking(req)
	if err != nil {
		return err
	}

	fmt.Printf("\n预约创建成功！\n")
	fmt.Printf("预约号: %s\n", bookingID)
	fmt.Printf("估价: %s\n", formatCents(estimate))
	return nil
}

func cmdList(c *Client) error {
	bookings, err := c.ListBookings()
	if err != nil {
		return err
	}

	if len(bookings) == 0 {
		fmt.Println("没有预约记录")
		return nil
	}

	fmt.Printf("%-12s %-10s %-15s %-12s %-10s\n", "ID", "日期", "时段", "状态", "估价")
	fmt.Println(strings.Repeat("-", 60))
	for _, b := range bookings {
		fmt.Printf("%-12s %-10s %-15s %-12s %s\n",
			b.ID[len(b.ID)-10:],
			b.MoveDate.Format("2006-01-02"),
			b.TimeSlot,
			b.Status,
			formatCents(b.Estimate),
		)
	}

	return nil
}

func cmdGet(c *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("请提供预约号")
	}

	booking, err := c.GetBooking(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("预约号: %s\n", booking.ID)
	fmt.Printf("客户: %s\n", booking.CustomerName)
	fmt.Printf("电话: %s\n", booking.Phone)
	fmt.Printf("搬家日期: %s\n", booking.MoveDate.Format("2006-01-02"))
	fmt.Printf("时段: %s\n", booking.TimeSlot)
	fmt.Printf("起始地址: %s\n", booking.FromAddress)
	fmt.Printf("目标地址: %s\n", booking.ToAddress)
	fmt.Printf("距离: %.1f公里\n", booking.DistanceKM)
	fmt.Printf("状态: %s\n", booking.Status)
	fmt.Printf("估价: %s\n", formatCents(booking.Estimate))
	if booking.FinalAmount > 0 {
		fmt.Printf("结算金额: %s\n", formatCents(booking.FinalAmount))
	}
	if booking.Review != nil {
		fmt.Printf("评分: %d星\n", booking.Review.Rating)
		if booking.Review.Comment != "" {
			fmt.Printf("评价: %s\n", booking.Review.Comment)
		}
	}

	return nil
}

func cmdCancel(c *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("请提供预约号")
	}

	fee, err := c.CancelBooking(args[0])
	if err != nil {
		return err
	}

	if fee == 0 {
		fmt.Println("预约已取消，无需支付违约金")
	} else {
		fmt.Printf("预约已取消，需支付违约金: %s\n", formatCents(fee))
	}

	return nil
}

func cmdComplete(c *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("请提供预约号")
	}

	if err := c.CompleteBooking(args[0]); err != nil {
		return err
	}

	fmt.Println("预约已标记为完成")
	return nil
}

func cmdSettle(c *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("请提供预约号")
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("=== 结算 ===")

	fmt.Println("是否需要更新物品清单？(y/n，默认不更新)")
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	var finalItems *core.ItemList
	if answer == "y" || answer == "yes" {
		items, err := readItems(reader)
		if err != nil {
			return err
		}
		finalItems = &items
	}

	fmt.Println("是否需要更新距离？(y/n，默认不更新)")
	answer, _ = reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	var finalDistanceKM float64
	if answer == "y" || answer == "yes" {
		distance, err := readFloat(reader, "请输入实际距离（公里）: ")
		if err != nil {
			return err
		}
		finalDistanceKM = distance
	}

	amount, err := c.SettleBooking(args[0], finalItems, finalDistanceKM)
	if err != nil {
		return err
	}

	fmt.Printf("结算金额: %s\n", formatCents(amount))
	return nil
}

func cmdReview(c *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("请提供预约号")
	}

	reader := bufio.NewReader(os.Stdin)

	rating, err := readInt(reader, "请输入评分 (1-5): ")
	if err != nil {
		return err
	}

	if rating < 1 || rating > 5 {
		return fmt.Errorf("评分必须在1-5之间")
	}

	comment, _ := readString(reader, "请输入评价（回车跳过）: ")

	if err := c.AddReview(args[0], rating, comment); err != nil {
		return err
	}

	fmt.Println("评价已提交")
	return nil
}

func cmdAvailability(c *Client, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	dateStr, err := readString(reader, "请输入查询日期 (YYYY-MM-DD): ")
	if err != nil {
		return err
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("无效的日期格式: %w", err)
	}

	slots, err := c.CheckAvailability(date)
	if err != nil {
		return err
	}

	if len(slots) == 0 {
		fmt.Println("该日期没有可用时段")
	} else {
		fmt.Printf("可用时段: %v\n", slots)
	}

	return nil
}

func readString(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func readInt(reader *bufio.Reader, prompt string) (int, error) {
	s, err := readString(reader, prompt)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(s)
}

func readFloat(reader *bufio.Reader, prompt string) (float64, error) {
	s, err := readString(reader, prompt)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(s, 64)
}

func readItems(reader *bufio.Reader) (core.ItemList, error) {
	items := core.ItemList{}

	fmt.Println("\n=== 物品清单 ===")

	count, err := readInt(reader, "请输入家具数量 (回车为0): ")
	if err != nil {
		count = 0
	}

	for i := 0; i < count; i++ {
		fmt.Printf("\n家具 %d:\n", i+1)
		name, _ := readString(reader, "  名称: ")
		size, _ := readString(reader, "  尺寸 (small/medium/large): ")
		items.Furniture = append(items.Furniture, core.Furniture{
			Name: name,
			Size: core.ItemSize(size),
		})
	}

	count, err = readInt(reader, "请输入家电数量 (回车为0): ")
	if err != nil {
		count = 0
	}

	for i := 0; i < count; i++ {
		fmt.Printf("\n家电 %d:\n", i+1)
		name, _ := readString(reader, "  名称: ")
		needDis, _ := readString(reader, "  需要拆装？(y/n): ")
		items.Appliances = append(items.Appliances, core.Appliance{
			Name:             name,
			NeedDisassemble:  strings.ToLower(needDis) == "y",
		})
	}

	items.BoxCount, err = readInt(reader, "请输入纸箱数量 (回车为0): ")
	if err != nil {
		items.BoxCount = 0
	}

	count, err = readInt(reader, "请输入特殊物品数量 (回车为0): ")
	if err != nil {
		count = 0
	}

	for i := 0; i < count; i++ {
		name, _ := readString(reader, fmt.Sprintf("  特殊物品 %d 名称: ", i+1))
		items.Specials = append(items.Specials, name)
	}

	return items, nil
}
