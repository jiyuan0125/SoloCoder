package main

import (
	"bus-station/internal/api"
	"bus-station/pkg/client"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8900", "服务端地址")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	c := client.NewClient(*serverURL)

	command := args[0]

	switch command {
	case "search":
		if len(args) < 4 {
			fmt.Println("用法: client search <出发站> <到达站> <日期(YYYY-MM-DD)")
			os.Exit(1)
		}
		searchSchedules(c, args[1], args[2], args[3])

	case "list":
		date := time.Now().Format("2006-01-02")
		if len(args) >= 2 {
			date = args[1]
		}
		listSchedules(c, date)

	case "seats":
		if len(args) < 3 {
			fmt.Println("用法: client seats <班次号> <日期(YYYY-MM-DD)")
			os.Exit(1)
		}
		getSeats(c, args[1], args[2])

	case "buy":
		if len(args) < 5 {
			fmt.Println("用法: client buy <班次号> <日期> <乘客姓名> <座位号1,座位号2...>")
			os.Exit(1)
		}
		seatNos, err := parseSeatNos(args[4])
		if err != nil {
			fmt.Printf("座位号解析错误: %v\n", err)
			os.Exit(1)
		}
		buyTickets(c, args[1], args[2], args[3], seatNos)

	case "ticket":
		if len(args) < 2 {
			fmt.Println("用法: client ticket <票号>")
			os.Exit(1)
		}
		getTicket(c, args[1])

	case "checkin":
		if len(args) < 2 {
			fmt.Println("用法: client checkin <票号>")
			os.Exit(1)
		}
		checkIn(c, args[1])

	case "refund":
		if len(args) < 2 {
			fmt.Println("用法: client refund <票号>")
			os.Exit(1)
		}
		requestRefund(c, args[1])

	case "pending-refunds":
		listPendingRefunds(c)

	case "process-refund":
		if len(args) < 3 {
			fmt.Println("用法: client process-refund <退款请求ID> <approve|reject> [拒绝原因]")
			os.Exit(1)
		}
		approved := args[2] == "approve"
		reason := ""
		if len(args) >= 4 {
			reason = args[3]
		}
		processRefund(c, args[1], approved, reason)

	case "create-schedule":
		if len(args) < 10 {
			fmt.Println("用法: client create-schedule <班次号> <日期> <出发站> <到达站> <出发时间(HH:MM)> <到达时间(HH:MM)> <车型> <票价> <总座位数>")
			os.Exit(1)
		}
		price, _ := strconv.Atoi(args[8])
		totalSeats, _ := strconv.Atoi(args[9])
		createSchedule(c, args[1], args[2], args[3], args[4], args[5], args[6], args[7], price, totalSeats)

	case "cancel-schedule":
		if len(args) < 3 {
			fmt.Println("用法: client cancel-schedule <班次号> <日期>")
			os.Exit(1)
		}
		cancelSchedule(c, args[1], args[2])

	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("客运站管理系统 - 命令行客户端")
	fmt.Println("")
	fmt.Println("用法: client [命令] [参数]")
	fmt.Println("")
	fmt.Println("命令列表:")
	fmt.Println("  search <出发站> <到达站> <日期>      - 搜索班次")
	fmt.Println("  list [日期]                      - 列出某天所有班次（默认今天）")
	fmt.Println("  seats <班次号> <日期>               - 查看班次座位情况")
	fmt.Println("  buy <班次号> <日期> <姓名> <座位> - 购票（座位用逗号分隔，最多5个）")
	fmt.Println("  ticket <票号>                       - 查询车票信息")
	fmt.Println("  checkin <票号>                    - 检票")
	fmt.Println("  refund <票号>                      - 申请退票")
	fmt.Println("  pending-refunds                    - 列出待处理退款")
	fmt.Println("  process-refund <ID> <approve|reject> [原因] - 处理退款")
	fmt.Println("  create-schedule ...                  - 创建班次")
	fmt.Println("  cancel-schedule <班次号> <日期>       - 取消班次")
}

func parseSeatNos(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, nil
}

func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

func searchSchedules(c *client.Client, departure, arrival, date string) {
	resp, err := c.SearchSchedules(departure, arrival, date)
	if err := client.HandleResponse(resp, err); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("搜索结果:")
	printJSON(resp.Data)
}

func listSchedules(c *client.Client, date string) {
	resp, err := c.ListSchedules(date)
	if err := client.HandleResponse(resp, err); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("日期 %s 的班次列表:\n", date)
	printJSON(resp.Data)
}

func getSeats(c *client.Client, scheduleNo, date string) {
	resp, err := c.GetSeats(scheduleNo, date)
	if err := client.HandleResponse(resp, err); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("座位情况:")
	printJSON(resp.Data)
}

func buyTickets(c *client.Client, scheduleNo, date, passengerName string, seatNos []int) {
	resp, err := c.PurchaseTickets(scheduleNo, date, passengerName, seatNos)
	if err := client.HandleResponse(resp, err); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("购票成功!")
	printJSON(resp.Data)
}

func getTicket(c *client.Client, ticketNo string) {
	resp, err := c.GetTicket(ticketNo)
	if err := client.HandleResponse(resp, err); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("车票信息:")
	printJSON(resp.Data)
}

func checkIn(c *client.Client, ticketNo string) {
	resp, err := c.CheckIn(ticketNo)
	if err := client.HandleResponse(resp, err); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("检票成功!")
	printJSON(resp.Data)
}

func requestRefund(c *client.Client, ticketNo string) {
	resp, err := c.RequestRefund(ticketNo)
	if err := client.HandleResponse(resp, err); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("退票申请成功!")
	printJSON(resp.Data)
}

func listPendingRefunds(c *client.Client) {
	resp, err := c.ListPendingRefunds()
	if err := client.HandleResponse(resp, err); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("待处理退款列表:")
	printJSON(resp.Data)
}

func processRefund(c *client.Client, requestID string, approved bool, reason string) {
	resp, err := c.ProcessRefund(requestID, approved, reason)
	if err := client.HandleResponse(resp, err); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("退款处理完成!")
	printJSON(resp.Data)
}

func createSchedule(c *client.Client, scheduleNo, date, departure, arrival, depTime, arrTime, busType string, price, totalSeats int) {
	req := &api.CreateScheduleRequest{
		ScheduleNo:       scheduleNo,
		Date:             date,
		DepartureStation: departure,
		ArrivalStation:   arrival,
		DepartureTime:    depTime,
		ArrivalTime:      arrTime,
		BusType:          busType,
		Price:            price,
		TotalSeats:       totalSeats,
	}

	resp, err := c.CreateSchedule(req)
	if err := client.HandleResponse(resp, err); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("班次创建成功!")
	printJSON(resp.Data)
}

func cancelSchedule(c *client.Client, scheduleNo, date string) {
	resp, err := c.CancelSchedule(scheduleNo, date)
	if err := client.HandleResponse(resp, err); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("班次已取消!")
	printJSON(resp.Data)
}
