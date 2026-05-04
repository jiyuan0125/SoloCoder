package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"go-delivery-tracking/common"
)

const (
	defaultServerAddr = "localhost:8080"
)

type Command string

const (
	CmdReport Command = "report"
	CmdQuery  Command = "query"
)

type ReportOptions struct {
	ServerAddr  string
	OrderID     string
	StatusName  string
	Longitude   float64
	Latitude    float64
	Timestamp   int64
	StatusesFile string
}

type QueryOptions struct {
	ServerAddr string
	OrderID    string
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := Command(os.Args[1])
	switch cmd {
	case CmdReport:
		handleReportCommand(os.Args[2:])
	case CmdQuery:
		handleQueryCommand(os.Args[2:])
	default:
		fmt.Printf("未知命令: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("用法:")
	fmt.Println("  骑手上报: client report [选项]")
	fmt.Println("  顾客查询: client query [选项]")
	fmt.Println()
	fmt.Println("骑手上报选项:")
	fmt.Println("  -s, --server      服务器地址 (默认: localhost:8080)")
	fmt.Println("  -o, --order-id    订单号 (必需)")
	fmt.Println("  -n, --status-name 状态名称: 取餐确认, 商家出发, 到达小区, 已送达 (必需)")
	fmt.Println("  -L, --longitude   经度 (必需)")
	fmt.Println("  -l, --latitude    纬度 (必需)")
	fmt.Println("  -t, --timestamp   时间戳 (默认: 当前时间)")
	fmt.Println("  -f, --file        从JSON文件读取多个状态 (可选，会忽略其他参数)")
	fmt.Println()
	fmt.Println("顾客查询选项:")
	fmt.Println("  -s, --server      服务器地址 (默认: localhost:8080)")
	fmt.Println("  -o, --order-id    订单号 (必需)")
}

func handleReportCommand(args []string) {
	opts := &ReportOptions{}
	flagSet := flag.NewFlagSet("report", flag.ExitOnError)
	flagSet.StringVar(&opts.ServerAddr, "server", defaultServerAddr, "服务器地址")
	flagSet.StringVar(&opts.ServerAddr, "s", defaultServerAddr, "服务器地址 (短选项)")
	flagSet.StringVar(&opts.OrderID, "order-id", "", "订单号")
	flagSet.StringVar(&opts.OrderID, "o", "", "订单号 (短选项)")
	flagSet.StringVar(&opts.StatusName, "status-name", "", "状态名称")
	flagSet.StringVar(&opts.StatusName, "n", "", "状态名称 (短选项)")
	flagSet.Float64Var(&opts.Longitude, "longitude", 0, "经度")
	flagSet.Float64Var(&opts.Longitude, "L", 0, "经度 (短选项)")
	flagSet.Float64Var(&opts.Latitude, "latitude", 0, "纬度")
	flagSet.Float64Var(&opts.Latitude, "l", 0, "纬度 (短选项)")
	flagSet.Int64Var(&opts.Timestamp, "timestamp", 0, "时间戳")
	flagSet.Int64Var(&opts.Timestamp, "t", 0, "时间戳 (短选项)")
	flagSet.StringVar(&opts.StatusesFile, "file", "", "状态文件")
	flagSet.StringVar(&opts.StatusesFile, "f", "", "状态文件 (短选项)")

	flagSet.Parse(args)

	if opts.StatusesFile != "" {
		handleReportFromFile(opts)
		return
	}

	if opts.OrderID == "" || opts.StatusName == "" {
		fmt.Println("错误: 订单号和状态名称是必需的")
		flagSet.Usage()
		os.Exit(1)
	}

	if opts.Timestamp == 0 {
		opts.Timestamp = time.Now().Unix()
	}

	status := common.DeliveryStatus{
		OrderID:    opts.OrderID,
		StatusName: common.StatusName(opts.StatusName),
		Longitude:  opts.Longitude,
		Latitude:   opts.Latitude,
		Timestamp:  opts.Timestamp,
	}

	conn, err := connectToServer(opts.ServerAddr)
	if err != nil {
		fmt.Printf("连接服务器失败: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	statuses := []common.DeliveryStatus{status}
	msg, err := common.BuildReportMessage(statuses)
	if err != nil {
		fmt.Printf("构建上报消息失败: %v\n", err)
		os.Exit(1)
	}

	if err := common.WriteMessage(conn, msg); err != nil {
		fmt.Printf("发送上报消息失败: %v\n", err)
		os.Exit(1)
	}

	respMsg, err := common.ReadMessage(conn)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		os.Exit(1)
	}

	var resp common.Response
	if err := json.Unmarshal(respMsg.Payload, &resp); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Println("上报成功")
	} else {
		fmt.Printf("上报失败: %s\n", resp.Error)
		os.Exit(1)
	}
}

func handleReportFromFile(opts *ReportOptions) {
	fileContent, err := os.ReadFile(opts.StatusesFile)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		os.Exit(1)
	}

	var statuses []common.DeliveryStatus
	if err := json.Unmarshal(fileContent, &statuses); err != nil {
		fmt.Printf("解析JSON失败: %v\n", err)
		os.Exit(1)
	}

	if len(statuses) == 0 {
		fmt.Println("错误: 文件中没有状态数据")
		os.Exit(1)
	}

	conn, err := connectToServer(opts.ServerAddr)
	if err != nil {
		fmt.Printf("连接服务器失败: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	msg, err := common.BuildReportMessage(statuses)
	if err != nil {
		fmt.Printf("构建上报消息失败: %v\n", err)
		os.Exit(1)
	}

	if err := common.WriteMessage(conn, msg); err != nil {
		fmt.Printf("发送上报消息失败: %v\n", err)
		os.Exit(1)
	}

	respMsg, err := common.ReadMessage(conn)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		os.Exit(1)
	}

	var resp common.Response
	if err := json.Unmarshal(respMsg.Payload, &resp); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("上报成功，共 %d 条状态\n", len(statuses))
	} else {
		fmt.Printf("上报失败: %s\n", resp.Error)
		os.Exit(1)
	}
}

func handleQueryCommand(args []string) {
	opts := &QueryOptions{}
	flagSet := flag.NewFlagSet("query", flag.ExitOnError)
	flagSet.StringVar(&opts.ServerAddr, "server", defaultServerAddr, "服务器地址")
	flagSet.StringVar(&opts.ServerAddr, "s", defaultServerAddr, "服务器地址 (短选项)")
	flagSet.StringVar(&opts.OrderID, "order-id", "", "订单号")
	flagSet.StringVar(&opts.OrderID, "o", "", "订单号 (短选项)")

	flagSet.Parse(args)

	if opts.OrderID == "" {
		fmt.Println("错误: 订单号是必需的")
		flagSet.Usage()
		os.Exit(1)
	}

	conn, err := connectToServer(opts.ServerAddr)
	if err != nil {
		fmt.Printf("连接服务器失败: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	msg, err := common.BuildQueryMessage(opts.OrderID)
	if err != nil {
		fmt.Printf("构建查询消息失败: %v\n", err)
		os.Exit(1)
	}

	if err := common.WriteMessage(conn, msg); err != nil {
		fmt.Printf("发送查询消息失败: %v\n", err)
		os.Exit(1)
	}

	respMsg, err := common.ReadMessage(conn)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		os.Exit(1)
	}

	var resp common.Response
	if err := json.Unmarshal(respMsg.Payload, &resp); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
		os.Exit(1)
	}

	if resp.Success && resp.Data != nil {
		fmt.Printf("订单 %s 的配送轨迹:\n", opts.OrderID)
		for i, status := range resp.Data.Statuses {
			fmt.Printf("  %d. 状态: %s, 时间: %s, 坐标: (%.6f, %.6f)\n",
				i+1,
				status.StatusName,
				time.Unix(status.Timestamp, 0).Format("2006-01-02 15:04:05"),
				status.Longitude,
				status.Latitude)
		}
	} else {
		fmt.Printf("查询失败: %s\n", resp.Error)
		os.Exit(1)
	}
}
