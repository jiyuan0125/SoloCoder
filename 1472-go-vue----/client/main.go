package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"taxisystem/common"
)

func printHelp() {
	fmt.Println(`出租车调度管理系统 - 命令行客户端

用法: taxi-client [全局选项] <命令> [命令选项]

全局选项:
  -server <地址>    服务端地址 (默认: http://localhost:8901)

可用命令:
  vehicle           车辆管理
    add              添加车辆
    list             列出所有车辆
    get              获取单个车辆
    status           更新车辆状态
    location         更新车辆位置

  order             订单管理
    estimate         预估费用
    create           创建订单
    list             列出所有订单
    get              获取订单详情
    accept           司机接单
    start            开始行程
    complete         完成行程
    rating           评价订单

  complaint         投诉管理
    list             列出待处理投诉
    handle           处理投诉

  metrics           指标查询
    show             显示系统指标

示例:
  taxi-client vehicle add --plate 京A12345 --driver 张三 --phone 13800138000 --lat 39.9 --lng 116.4
  taxi-client order estimate --pickup-lat 39.9 --pickup-lng 116.4 --dest-lat 40.0 --dest-lng 116.5
  taxi-client order create --passenger-id P001 --phone 13900139000 --pickup-lat 39.9 --pickup-lng 116.4 --dest-lat 40.0 --dest-lng 116.5
  taxi-client order accept --order-id <订单ID> --plate 京A12345
  taxi-client order start --order-id <订单ID> --distance 10.5 --duration 1800 --low-speed 5
  taxi-client order complete --order-id <订单ID>
  taxi-client order rating --order-id <订单ID> --stars 5 --content "服务很好"
  taxi-client metrics show
`)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	var serverURL string
	flag.StringVar(&serverURL, "server", "http://localhost:8901", "服务端地址")
	flag.Parse()

	client := NewAPIClient(serverURL)

	args := flag.Args()
	if len(args) == 0 {
		printHelp()
		os.Exit(0)
	}

	cmd := args[0]

	switch cmd {
	case "vehicle":
		if len(args) < 2 {
			fmt.Println("请指定车辆子命令: add|list|get|status|location")
			os.Exit(1)
		}
		handleVehicleCommand(client, args[1], args[2:])
	case "order":
		if len(args) < 2 {
			fmt.Println("请指定订单子命令: estimate|create|list|get|accept|start|complete|rating")
			os.Exit(1)
		}
		handleOrderCommand(client, args[1], args[2:])
	case "complaint":
		if len(args) < 2 {
			fmt.Println("请指定投诉子命令: list|handle")
			os.Exit(1)
		}
		handleComplaintCommand(client, args[1], args[2:])
	case "metrics":
		if len(args) < 2 || args[1] != "show" {
			fmt.Println("请使用: taxi-client metrics show")
			os.Exit(1)
		}
		handleMetricsShow(client)
	case "-h", "--help", "help":
		printHelp()
	default:
		fmt.Printf("未知命令: %s\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func handleVehicleCommand(client *APIClient, subcmd string, args []string) {
	fs := flag.NewFlagSet("vehicle "+subcmd, flag.ExitOnError)

	switch subcmd {
	case "add":
		var plate, driver, phone string
		var lat, lng float64
		fs.StringVar(&plate, "plate", "", "车牌号")
		fs.StringVar(&driver, "driver", "", "司机姓名")
		fs.StringVar(&phone, "phone", "", "司机手机号")
		fs.Float64Var(&lat, "lat", 0, "纬度")
		fs.Float64Var(&lng, "lng", 0, "经度")
		fs.Parse(args)

		if plate == "" || driver == "" || phone == "" {
			fmt.Println("缺少必要参数: --plate, --driver, --phone")
			os.Exit(1)
		}

		resp, err := client.CreateVehicle(plate, driver, phone, lat, lng)
		printResult(resp, err)

	case "list":
		resp, err := client.ListVehicles()
		printResult(resp, err)

	case "get":
		var plate string
		fs.StringVar(&plate, "plate", "", "车牌号")
		fs.Parse(args)

		if plate == "" {
			fmt.Println("缺少必要参数: --plate")
			os.Exit(1)
		}

		resp, err := client.GetVehicle(plate)
		printResult(resp, err)

	case "status":
		var plate string
		var status string
		fs.StringVar(&plate, "plate", "", "车牌号")
		fs.StringVar(&status, "status", "", "状态: idle|occupied|reserved|rest|maintenance")
		fs.Parse(args)

		if plate == "" || status == "" {
			fmt.Println("缺少必要参数: --plate, --status")
			os.Exit(1)
		}

		resp, err := client.UpdateVehicleStatus(plate, common.VehicleStatus(status))
		printResult(resp, err)

	case "location":
		var plate string
		var lat, lng float64
		fs.StringVar(&plate, "plate", "", "车牌号")
		fs.Float64Var(&lat, "lat", 0, "纬度")
		fs.Float64Var(&lng, "lng", 0, "经度")
		fs.Parse(args)

		if plate == "" {
			fmt.Println("缺少必要参数: --plate")
			os.Exit(1)
		}

		resp, err := client.UpdateVehicleLocation(plate, lat, lng)
		printResult(resp, err)

	default:
		fmt.Printf("未知车辆子命令: %s\n", subcmd)
		os.Exit(1)
	}
}

func handleOrderCommand(client *APIClient, subcmd string, args []string) {
	fs := flag.NewFlagSet("order "+subcmd, flag.ExitOnError)

	switch subcmd {
	case "estimate":
		var pickupLat, pickupLng, destLat, destLng float64
		fs.Float64Var(&pickupLat, "pickup-lat", 0, "上车点纬度")
		fs.Float64Var(&pickupLng, "pickup-lng", 0, "上车点经度")
		fs.Float64Var(&destLat, "dest-lat", 0, "目的地纬度")
		fs.Float64Var(&destLng, "dest-lng", 0, "目的地经度")
		fs.Parse(args)

		pickup := common.Location{Latitude: pickupLat, Longitude: pickupLng}
		dest := common.Location{Latitude: destLat, Longitude: destLng}

		resp, err := client.EstimateFare(pickup, dest)
		printResult(resp, err)

	case "create":
		var passengerID, phone string
		var pickupLat, pickupLng, destLat, destLng float64
		fs.StringVar(&passengerID, "passenger-id", "", "乘客ID")
		fs.StringVar(&phone, "phone", "", "乘客手机号")
		fs.Float64Var(&pickupLat, "pickup-lat", 0, "上车点纬度")
		fs.Float64Var(&pickupLng, "pickup-lng", 0, "上车点经度")
		fs.Float64Var(&destLat, "dest-lat", 0, "目的地纬度")
		fs.Float64Var(&destLng, "dest-lng", 0, "目的地经度")
		fs.Parse(args)

		if passengerID == "" || phone == "" {
			fmt.Println("缺少必要参数: --passenger-id, --phone")
			os.Exit(1)
		}

		pickup := common.Location{Latitude: pickupLat, Longitude: pickupLng}
		dest := common.Location{Latitude: destLat, Longitude: destLng}

		resp, err := client.CreateOrder(passengerID, phone, pickup, dest)
		printResult(resp, err)

	case "list":
		resp, err := client.ListOrders()
		printResult(resp, err)

	case "get":
		var orderID string
		fs.StringVar(&orderID, "order-id", "", "订单ID")
		fs.Parse(args)

		if orderID == "" {
			fmt.Println("缺少必要参数: --order-id")
			os.Exit(1)
		}

		resp, err := client.GetOrder(orderID)
		printResult(resp, err)

	case "accept":
		var orderID, plate string
		fs.StringVar(&orderID, "order-id", "", "订单ID")
		fs.StringVar(&plate, "plate", "", "车牌号")
		fs.Parse(args)

		if orderID == "" || plate == "" {
			fmt.Println("缺少必要参数: --order-id, --plate")
			os.Exit(1)
		}

		resp, err := client.AcceptOrder(orderID, plate)
		printResult(resp, err)

	case "start":
		var orderID string
		var distance float64
		var duration int64
		var lowSpeed int
		fs.StringVar(&orderID, "order-id", "", "订单ID")
		fs.Float64Var(&distance, "distance", 0, "实际距离(公里)")
		fs.Int64Var(&duration, "duration", 0, "预计时长(秒)")
		fs.IntVar(&lowSpeed, "low-speed", 0, "低速行驶分钟数")
		fs.Parse(args)

		if orderID == "" {
			fmt.Println("缺少必要参数: --order-id")
			os.Exit(1)
		}

		resp, err := client.StartTrip(orderID, distance, duration, lowSpeed)
		printResult(resp, err)

	case "complete":
		var orderID string
		fs.StringVar(&orderID, "order-id", "", "订单ID")
		fs.Parse(args)

		if orderID == "" {
			fmt.Println("缺少必要参数: --order-id")
			os.Exit(1)
		}

		resp, err := client.CompleteTrip(orderID)
		printResult(resp, err)

	case "rating":
		var orderID string
		var stars int
		var content string
		fs.StringVar(&orderID, "order-id", "", "订单ID")
		fs.IntVar(&stars, "stars", 0, "评分(1-5星)")
		fs.StringVar(&content, "content", "", "评价内容")
		fs.Parse(args)

		if orderID == "" || stars == 0 {
			fmt.Println("缺少必要参数: --order-id, --stars")
			os.Exit(1)
		}

		resp, err := client.SubmitRating(orderID, stars, content)
		printResult(resp, err)

	default:
		fmt.Printf("未知订单子命令: %s\n", subcmd)
		os.Exit(1)
	}
}

func handleComplaintCommand(client *APIClient, subcmd string, args []string) {
	fs := flag.NewFlagSet("complaint "+subcmd, flag.ExitOnError)

	switch subcmd {
	case "list":
		resp, err := client.ListComplaints()
		printResult(resp, err)

	case "handle":
		var complaintID string
		fs.StringVar(&complaintID, "id", "", "投诉ID")
		fs.Parse(args)

		if complaintID == "" {
			fmt.Println("缺少必要参数: --id")
			os.Exit(1)
		}

		resp, err := client.HandleComplaint(complaintID)
		printResult(resp, err)

	default:
		fmt.Printf("未知投诉子命令: %s\n", subcmd)
		os.Exit(1)
	}
}

func handleMetricsShow(client *APIClient) {
	resp, err := client.GetMetrics()
	printResult(resp, err)
}

func printResult(resp *common.APIResponse, err error) {
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	output, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))

	if resp.Code != 0 {
		os.Exit(1)
	}
}
