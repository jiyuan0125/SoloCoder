package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create-account":
		cmdCreateAccount(os.Args[2:])
	case "get-account":
		cmdGetAccount(os.Args[2:])
	case "recharge":
		cmdRecharge(os.Args[2:])
	case "pass":
		cmdPass(os.Args[2:])
	case "export":
		cmdExport(os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("ETC 管理系统客户端")
	fmt.Println("")
	fmt.Println("用法:")
	fmt.Println("  etc-client <命令> [参数]")
	fmt.Println("")
	fmt.Println("命令:")
	fmt.Println("  create-account  创建 ETC 账户")
	fmt.Println("  get-account     查询账户信息")
	fmt.Println("  recharge        账户充值")
	fmt.Println("  pass            记录通行")
	fmt.Println("  export          导出通行记录")
	fmt.Println("  help            显示帮助信息")
	fmt.Println("")
	fmt.Println("使用 'etc-client <命令> -h' 查看命令详情")
}

func cmdCreateAccount(args []string) {
	fs := flag.NewFlagSet("create-account", flag.ExitOnError)
	licensePlate := fs.String("plate", "", "车牌号 (必填)")
	bankCard := fs.String("card", "", "银行卡号 (必填)")
	balance := fs.Float64("balance", 0, "预存余额")
	vehicleType := fs.Int("type", 1, "车辆类型: 1-客车, 2-货车")
	seats := fs.Int("seats", 0, "客车座位数 (客车必填)")
	loadWeight := fs.Float64("weight", 0, "货车载重量(吨) (货车必填)")
	server := fs.String("server", "http://localhost:8080", "服务端地址")

	fs.Parse(args)

	if *licensePlate == "" || *bankCard == "" {
		fmt.Println("错误: 车牌号和银行卡号为必填项")
		fs.Usage()
		return
	}

	err := CreateAccount(*server, *licensePlate, *bankCard, *balance, *vehicleType, *seats, *loadWeight)
	if err != nil {
		fmt.Printf("创建账户失败: %v\n", err)
		return
	}

	fmt.Println("账户创建成功!")
}

func cmdGetAccount(args []string) {
	fs := flag.NewFlagSet("get-account", flag.ExitOnError)
	licensePlate := fs.String("plate", "", "车牌号 (必填)")
	server := fs.String("server", "http://localhost:8080", "服务端地址")

	fs.Parse(args)

	if *licensePlate == "" {
		fmt.Println("错误: 车牌号为必填项")
		fs.Usage()
		return
	}

	account, err := GetAccount(*server, *licensePlate)
	if err != nil {
		fmt.Printf("查询账户失败: %v\n", err)
		return
	}

	PrintAccount(account)
}

func cmdRecharge(args []string) {
	fs := flag.NewFlagSet("recharge", flag.ExitOnError)
	licensePlate := fs.String("plate", "", "车牌号 (必填)")
	amount := fs.Float64("amount", 0, "充值金额 (必填)")
	server := fs.String("server", "http://localhost:8080", "服务端地址")

	fs.Parse(args)

	if *licensePlate == "" || *amount <= 0 {
		fmt.Println("错误: 车牌号和充值金额(>0)为必填项")
		fs.Usage()
		return
	}

	balance, err := Recharge(*server, *licensePlate, *amount)
	if err != nil {
		fmt.Printf("充值失败: %v\n", err)
		return
	}

	fmt.Printf("充值成功! 当前余额: %.2f 元\n", balance)
}

func cmdPass(args []string) {
	fs := flag.NewFlagSet("pass", flag.ExitOnError)
	licensePlate := fs.String("plate", "", "车牌号 (必填)")
	entryStation := fs.String("entry", "", "入口站 (必填)")
	exitStation := fs.String("exit", "", "出口站 (必填)")
	mileage := fs.Float64("mileage", 0, "行驶里程(公里) (必填)")
	isFree := fs.Bool("free", false, "是否免费通行")
	server := fs.String("server", "http://localhost:8080", "服务端地址")

	fs.Parse(args)

	if *licensePlate == "" || *entryStation == "" || *exitStation == "" || *mileage <= 0 {
		fmt.Println("错误: 车牌号、入口站、出口站、行驶里程为必填项")
		fs.Usage()
		return
	}

	record, err := Pass(*server, *licensePlate, *entryStation, *exitStation, *mileage, *isFree)
	if err != nil {
		fmt.Printf("通行处理失败: %v\n", err)
		return
	}

	PrintPassRecord(record)
}

func cmdExport(args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	startDate := fs.String("start", "", "开始日期 (YYYY-MM-DD) (必填)")
	endDate := fs.String("end", "", "结束日期 (YYYY-MM-DD) (必填)")
	outputFile := fs.String("output", "pass_records.csv", "输出文件名")
	server := fs.String("server", "http://localhost:8080", "服务端地址")

	fs.Parse(args)

	if *startDate == "" || *endDate == "" {
		fmt.Println("错误: 开始日期和结束日期为必填项")
		fs.Usage()
		return
	}

	err := ExportRecords(*server, *startDate, *endDate, *outputFile)
	if err != nil {
		fmt.Printf("导出失败: %v\n", err)
		return
	}

	fmt.Printf("导出成功! 文件: %s\n", *outputFile)
}
