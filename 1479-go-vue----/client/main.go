package main

import (
	"flag"
	"fmt"
	"os"
)

const defaultServerURL = "http://localhost:8908"

var serverURL string

func main() {
	flag.StringVar(&serverURL, "server", defaultServerURL, "服务端地址")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]
	var err error

	switch command {
	case "student":
		err = handleStudentCommand(args[1:])
	case "coach":
		err = handleCoachCommand(args[1:])
	case "exam":
		err = handleExamCommand(args[1:])
	case "help":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("驾校管理系统客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client [选项] <命令> [子命令] [参数]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -server string   服务端地址 (默认 \"http://localhost:8908\")")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  student   学员管理")
	fmt.Println("  coach     教练管理")
	fmt.Println("  exam      考试管理")
	fmt.Println()
	fmt.Println("学员管理子命令:")
	fmt.Println("  register <姓名> <身份证号> <手机号> <车型>    注册学员")
	fmt.Println("  list                                            列出所有学员")
	fmt.Println("  get <学员ID>                                    获取学员信息")
	fmt.Println("  status <学员ID> <科目> <状态>                   更新科目状态")
	fmt.Println("  hours <学员ID> <科目> <学时>                    添加学时")
	fmt.Println()
	fmt.Println("教练管理子命令:")
	fmt.Println("  add <姓名> <准教车型> <手机号>               添加教练")
	fmt.Println("  list                                           列出所有教练")
	fmt.Println("  get <教练ID>                                   获取教练信息")
	fmt.Println("  practice <学员ID> <教练ID> <开始时间> <结束时间> <科目>  预约练车")
	fmt.Println()
	fmt.Println("考试管理子命令:")
	fmt.Println("  plan <日期> <科目> <考场> <名额>          创建考试计划")
	fmt.Println("  plans                                         列出所有考试计划")
	fmt.Println("  book <学员ID> <考试计划ID>                    预约考试")
	fmt.Println("  confirm <预约ID>                              确认候补预约")
	fmt.Println("  cancel <学员ID> <预约ID>                      取消考试预约")
	fmt.Println("  approve <预约ID> <true/false>              审批取消请求")
	fmt.Println("  bookings [学员ID]                             列出考试预约")
	fmt.Println()
	fmt.Println("车型: C1, C2, A1, B2")
	fmt.Println("科目: 1, 2, 3, 4")
	fmt.Println("状态: not_started, studying, passed")
}
