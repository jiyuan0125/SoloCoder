package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"usedcar/api"
)

func printUsage() {
	fmt.Println("二手车评估系统客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client [--server <url>] <command> [arguments]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  create-vehicle      创建车辆")
	fmt.Println("  list-vehicles       列出车辆")
	fmt.Println("  get-vehicle         查看车辆详情")
	fmt.Println("  evaluate            评估车辆")
	fmt.Println("  list-evaluations    查看车辆评估历史")
	fmt.Println("  pay-deposit         支付定金")
	fmt.Println("  pay-full            支付全款")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  client create-vehicle --brand=大众 --model=帕萨特 --year=2020 --color=黑色 --mileage=8.5 --displacement=2.0 --transmission=automatic")
	fmt.Println("  client list-vehicles")
	fmt.Println("  client list-vehicles --status=in_stock")
	fmt.Println("  client get-vehicle --id=VH...")
	fmt.Println("  client evaluate --vehicle-id=VH... --evaluator=张评估师 --exterior=5 --interior=4 --engine=4 --chassis=4 --description=车况良好")
	fmt.Println("  client list-evaluations --vehicle-id=VH...")
	fmt.Println("  client pay-deposit --vehicle-id=VH... --report-id=PG... --customer=张三")
	fmt.Println("  client pay-full --vehicle-id=VH... --customer=张三")
	os.Exit(1)
}

func getServerURL() string {
	server := os.Getenv("SERVER_URL")
	if server == "" {
		server = "http://localhost:8080"
	}
	return server
}

func cmdCreateVehicle(args []string, client *APIClient) {
	fs := flag.NewFlagSet("create-vehicle", flag.ExitOnError)
	brand := fs.String("brand", "", "品牌")
	model := fs.String("model", "", "型号")
	year := fs.Int("year", 0, "年份")
	color := fs.String("color", "", "颜色")
	mileage := fs.Float64("mileage", 0, "行驶里程（万公里）")
	displacement := fs.Float64("displacement", 0, "排量")
	transmission := fs.String("transmission", "automatic", "变速箱类型 (manual/automatic)")
	fs.Parse(args)

	if *brand == "" || *model == "" || *year == 0 || *color == "" {
		fmt.Println("错误: brand, model, year, color 为必填项")
		fs.Usage()
		os.Exit(1)
	}

	req := api.CreateVehicleRequest{
		Brand:        *brand,
		Model:        *model,
		Year:         *year,
		Color:        *color,
		Mileage:      *mileage,
		Displacement: *displacement,
		Transmission: api.TransmissionType(*transmission),
	}

	vehicle, err := client.CreateVehicle(req)
	if err != nil {
		fmt.Printf("创建失败: %v\n", err)
		os.Exit(1)
	}

	printVehicle(vehicle)
}

func cmdListVehicles(args []string, client *APIClient) {
	fs := flag.NewFlagSet("list-vehicles", flag.ExitOnError)
	status := fs.String("status", "", "状态筛选 (in_stock/reserved/sold)")
	fs.Parse(args)

	vehicles, err := client.ListVehicles(api.VehicleStatus(*status))
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		os.Exit(1)
	}

	if len(vehicles) == 0 {
		fmt.Println("暂无车辆")
		return
	}

	fmt.Printf("共 %d 辆车:\n", len(vehicles))
	fmt.Println(strings.Repeat("-", 80))
	for _, v := range vehicles {
		printVehicleSummary(&v)
	}
}

func cmdGetVehicle(args []string, client *APIClient) {
	fs := flag.NewFlagSet("get-vehicle", flag.ExitOnError)
	id := fs.String("id", "", "车辆ID")
	fs.Parse(args)

	if *id == "" {
		fmt.Println("错误: id 为必填项")
		fs.Usage()
		os.Exit(1)
	}

	vehicle, err := client.GetVehicle(*id)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		os.Exit(1)
	}

	printVehicle(vehicle)
}

func cmdEvaluate(args []string, client *APIClient) {
	fs := flag.NewFlagSet("evaluate", flag.ExitOnError)
	vehicleID := fs.String("vehicle-id", "", "车辆ID")
	evaluator := fs.String("evaluator", "", "评估师")
	exterior := fs.Int("exterior", 0, "外观评分 (1-5)")
	interior := fs.Int("interior", 0, "内饰评分 (1-5)")
	engine := fs.Int("engine", 0, "发动机评分 (1-5)")
	chassis := fs.Int("chassis", 0, "底盘评分 (1-5)")
	description := fs.String("description", "", "车况描述")
	fs.Parse(args)

	if *vehicleID == "" || *evaluator == "" {
		fmt.Println("错误: vehicle-id, evaluator 为必填项")
		fs.Usage()
		os.Exit(1)
	}

	req := api.EvaluateVehicleRequest{
		VehicleID:   *vehicleID,
		Evaluator:   *evaluator,
		Condition: api.Condition{
			Exterior: *exterior,
			Interior: *interior,
			Engine:   *engine,
			Chassis:  *chassis,
		},
		Description: *description,
	}

	evaluation, err := client.EvaluateVehicle(req)
	if err != nil {
		fmt.Printf("评估失败: %v\n", err)
		os.Exit(1)
	}

	printEvaluation(evaluation)
}

func cmdListEvaluations(args []string, client *APIClient) {
	fs := flag.NewFlagSet("list-evaluations", flag.ExitOnError)
	vehicleID := fs.String("vehicle-id", "", "车辆ID")
	fs.Parse(args)

	if *vehicleID == "" {
		fmt.Println("错误: vehicle-id 为必填项")
		fs.Usage()
		os.Exit(1)
	}

	evaluations, err := client.ListEvaluations(*vehicleID)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		os.Exit(1)
	}

	if len(evaluations) == 0 {
		fmt.Println("暂无评估记录")
		return
	}

	fmt.Printf("共 %d 条评估记录:\n", len(evaluations))
	fmt.Println(strings.Repeat("-", 80))
	for _, e := range evaluations {
		printEvaluationSummary(&e)
	}
}

func cmdPayDeposit(args []string, client *APIClient) {
	fs := flag.NewFlagSet("pay-deposit", flag.ExitOnError)
	vehicleID := fs.String("vehicle-id", "", "车辆ID")
	reportID := fs.String("report-id", "", "评估报告ID")
	customer := fs.String("customer", "", "客户姓名")
	amount := fs.String("amount", "", "定金金额（元，可选）")
	fs.Parse(args)

	if *vehicleID == "" || *customer == "" {
		fmt.Println("错误: vehicle-id, customer 为必填项")
		fs.Usage()
		os.Exit(1)
	}

	req := api.PayDepositRequest{
		VehicleID:    *vehicleID,
		ReportID:     *reportID,
		CustomerName: *customer,
	}

	if *amount != "" {
		yuan, err := strconv.ParseFloat(*amount, 64)
		if err != nil {
			fmt.Printf("金额格式错误: %v\n", err)
			os.Exit(1)
		}
		fen := int64(yuan * 100)
		req.AmountFen = &fen
	}

	deposit, err := client.PayDeposit(req)
	if err != nil {
		fmt.Printf("支付失败: %v\n", err)
		os.Exit(1)
	}

	printDeposit(deposit)
}

func cmdPayFull(args []string, client *APIClient) {
	fs := flag.NewFlagSet("pay-full", flag.ExitOnError)
	vehicleID := fs.String("vehicle-id", "", "车辆ID")
	customer := fs.String("customer", "", "客户姓名")
	fs.Parse(args)

	if *vehicleID == "" || *customer == "" {
		fmt.Println("错误: vehicle-id, customer 为必填项")
		fs.Usage()
		os.Exit(1)
	}

	req := api.PayFullRequest{
		VehicleID:    *vehicleID,
		CustomerName: *customer,
	}

	transfer, err := client.PayFull(req)
	if err != nil {
		fmt.Printf("支付失败: %v\n", err)
		os.Exit(1)
	}

	printTransfer(transfer)
}

func printVehicle(v *api.Vehicle) {
	statusMap := map[api.VehicleStatus]string{
		api.StatusInStock:  "在库",
		api.StatusReserved: "已预订",
		api.StatusSold:     "已售出",
	}
	transMap := map[api.TransmissionType]string{
		api.TransmissionManual:    "手动",
		api.TransmissionAutomatic: "自动",
	}

	fmt.Println("=== 车辆信息 ===")
	fmt.Printf("ID:           %s\n", v.ID)
	fmt.Printf("品牌:         %s\n", v.Brand)
	fmt.Printf("型号:         %s\n", v.Model)
	fmt.Printf("年份:         %d\n", v.Year)
	fmt.Printf("颜色:         %s\n", v.Color)
	fmt.Printf("行驶里程:     %s 万公里\n", formatMileage(v.Mileage))
	fmt.Printf("排量:         %.1f L\n", v.Displacement)
	fmt.Printf("变速箱:       %s\n", transMap[v.Transmission])
	fmt.Printf("状态:         %s\n", statusMap[v.Status])
}

func printVehicleSummary(v *api.Vehicle) {
	statusMap := map[api.VehicleStatus]string{
		api.StatusInStock:  "在库",
		api.StatusReserved: "已预订",
		api.StatusSold:     "已售出",
	}
	fmt.Printf("%s | %s %s (%d年) | %s万公里 | %s\n",
		v.ID, v.Brand, v.Model, v.Year, formatMileage(v.Mileage), statusMap[v.Status])
}

func printEvaluation(e *api.Evaluation) {
	fmt.Println("=== 评估报告 ===")
	fmt.Printf("报告编号:     %s\n", e.ReportID)
	fmt.Printf("车辆ID:       %s\n", e.VehicleID)
	fmt.Printf("评估师:       %s\n", e.Evaluator)
	fmt.Printf("外观:         %d 星\n", e.Condition.Exterior)
	fmt.Printf("内饰:         %d 星\n", e.Condition.Interior)
	fmt.Printf("发动机:       %d 星\n", e.Condition.Engine)
	fmt.Printf("底盘:         %d 星\n", e.Condition.Chassis)
	fmt.Printf("参考价:       %s 元\n", fenToYuan(e.ReferencePriceFen))
	fmt.Printf("评估价:       %s 元\n", fenToYuan(e.FinalPriceFen))
	fmt.Printf("车况描述:     %s\n", e.Description)
	if e.IsDuplicate != nil && *e.IsDuplicate {
		fmt.Println("⚠  警告: 可能存在重复评估")
	}
}

func printEvaluationSummary(e *api.Evaluation) {
	dup := ""
	if e.IsDuplicate != nil && *e.IsDuplicate {
		dup = " [重复]"
	}
	fmt.Printf("%s | 评估师:%s | 外观:%d 内饰:%d 发动机:%d 底盘:%d | 评估价:%s元%s\n",
		e.ReportID, e.Evaluator,
		e.Condition.Exterior, e.Condition.Interior, e.Condition.Engine, e.Condition.Chassis,
		fenToYuan(e.FinalPriceFen), dup)
}

func printDeposit(d *api.Deposit) {
	fmt.Println("=== 定金支付 ===")
	fmt.Printf("订单ID:       %s\n", d.ID)
	fmt.Printf("车辆ID:       %s\n", d.VehicleID)
	fmt.Printf("客户:         %s\n", d.CustomerName)
	fmt.Printf("定金金额:     %s 元\n", fenToYuan(d.AmountFen))
	fmt.Println("✅ 车辆状态已变更为：已预订")
}

func printTransfer(t *api.TransferRecord) {
	fmt.Println("=== 过户记录 ===")
	fmt.Printf("过户ID:       %s\n", t.ID)
	fmt.Printf("车辆ID:       %s\n", t.VehicleID)
	fmt.Printf("客户:         %s\n", t.CustomerName)
	fmt.Printf("全款金额:     %s 元\n", fenToYuan(t.FullPriceFen))
	fmt.Println("✅ 车辆状态已变更为：已售出")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
	}

	serverURL := getServerURL()
	if len(os.Args) >= 2 && strings.HasPrefix(os.Args[1], "--server=") {
		parts := strings.SplitN(os.Args[1], "=", 2)
		if len(parts) == 2 {
			serverURL = parts[1]
		}
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	if len(os.Args) < 2 {
		printUsage()
	}

	client := NewAPIClient(serverURL)
	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "create-vehicle":
		cmdCreateVehicle(args, client)
	case "list-vehicles":
		cmdListVehicles(args, client)
	case "get-vehicle":
		cmdGetVehicle(args, client)
	case "evaluate":
		cmdEvaluate(args, client)
	case "list-evaluations":
		cmdListEvaluations(args, client)
	case "pay-deposit":
		cmdPayDeposit(args, client)
	case "pay-full":
		cmdPayFull(args, client)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n", cmd)
		printUsage()
	}
}
