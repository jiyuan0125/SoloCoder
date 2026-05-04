package main

import (
	"flag"
	"fmt"
	"hospital-bed/internal/client"
	"hospital-bed/shared/protocol"
	"os"
	"strconv"
	"strings"
)

const usage = `医院床位管理系统 - 命令行客户端

使用方式:
  bed-client [全局选项] <命令> [命令选项]

全局选项:
  -server string    服务端地址 (默认 "localhost:8080")

工作人员命令:
  admit           办理入院
  discharge       办理出院
  status          查看科室床位状态
  patient         查看患者床位信息
  queue           查看候床队列
  transfer        患者转科

管理员命令:
  configure       配置科室床位布局

使用 'bed-client <命令> -help' 查看命令详细帮助
`

func printUsage() {
	fmt.Println(usage)
	os.Exit(1)
}

func bedTypeLabel(t protocol.BedType) string {
	switch t {
	case protocol.BedTypeSingle:
		return "单人间"
	case protocol.BedTypeDouble:
		return "双人间"
	case protocol.BedTypeTriple:
		return "三人间"
	case protocol.BedTypeExtra:
		return "加床"
	default:
		return string(t)
	}
}

func parseBedType(s string) (protocol.BedType, error) {
	switch strings.ToLower(s) {
	case "single", "单人间", "单":
		return protocol.BedTypeSingle, nil
	case "double", "双人间", "双":
		return protocol.BedTypeDouble, nil
	case "triple", "三人间", "三":
		return protocol.BedTypeTriple, nil
	case "extra", "加床", "加":
		return protocol.BedTypeExtra, nil
	default:
		return "", fmt.Errorf("无效的床位类型: %s，可选值: single/double/triple/extra 或 单人间/双人间/三人间/加床", s)
	}
}

func cmdAdmit(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("admit", flag.ExitOnError)
	patientID := fs.String("id", "", "患者ID (必需)")
	patientName := fs.String("name", "", "患者姓名 (必需)")
	dept := fs.String("dept", "", "目标科室 (必需)")
	ward := fs.String("ward", "", "目标病区 (可选，不指定则在科室所有病区中分配)")
	preferred := fs.String("prefer", "single", "偏好床位类型: single/double/triple/extra")

	fs.Usage = func() {
		fmt.Println(`办理入院

使用方式:
  bed-client admit -id <患者ID> -name <姓名> -dept <科室> [选项]

选项:
  -id string      患者ID (必需)
  -name string    患者姓名 (必需)
  -dept string    目标科室 (必需)
  -ward string    目标病区 (可选)
  -prefer string  偏好床位类型 (默认 "single")
                  可选: single(单人间), double(双人间), triple(三人间), extra(加床)

示例:
  bed-client admit -id P001 -name 张三 -dept 内科 -prefer single
  bed-client admit -id P002 -name 李四 -dept 内科 -ward 一区 -prefer double
`)
	}

	fs.Parse(args)

	if *patientID == "" || *patientName == "" || *dept == "" {
		fs.Usage()
		os.Exit(1)
	}

	bedType, err := parseBedType(*preferred)
	if err != nil {
		fmt.Println("错误:", err)
		os.Exit(1)
	}

	req := protocol.AdmissionRequest{
		PatientID:      *patientID,
		PatientName:    *patientName,
		DepartmentName: *dept,
		PreferredType:  bedType,
	}
	if *ward != "" {
		req.WardName = ward
	}

	resp, err := c.AdmitPatient(req)
	if err != nil {
		fmt.Println("错误:", err)
		os.Exit(1)
	}

	switch resp.Result {
	case protocol.AdmissionResultAllocated:
		fmt.Println("入院办理成功！")
		fmt.Printf("  床位编号: %s\n", resp.BedID)
		fmt.Printf("  床位类型: %s\n", bedTypeLabel(resp.BedType))
		if resp.DowngradeReason != nil {
			fmt.Println()
			fmt.Println("注意: 床位已降级分配")
			fmt.Printf("  偏好类型: %s\n", bedTypeLabel(resp.DowngradeReason.PreferredType))
			fmt.Printf("  实际分配: %s\n", bedTypeLabel(resp.DowngradeReason.ActualType))
			fmt.Printf("  原因: %s\n", resp.DowngradeReason.Reason)
		}
	case protocol.AdmissionResultQueued:
		fmt.Println("当前无可用床位，患者已加入候床队列")
		fmt.Printf("  队列位置: 第 %d 位\n", resp.QueuePosition)
	}
}

func cmdDischarge(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("discharge", flag.ExitOnError)
	patientID := fs.String("id", "", "患者ID (必需)")

	fs.Usage = func() {
		fmt.Println(`办理出院

使用方式:
  bed-client discharge -id <患者ID>

选项:
  -id string    患者ID (必需)

示例:
  bed-client discharge -id P001
`)
	}

	fs.Parse(args)

	if *patientID == "" {
		fs.Usage()
		os.Exit(1)
	}

	req := protocol.DischargeRequest{PatientID: *patientID}
	resp, err := c.DischargePatient(req)
	if err != nil {
		fmt.Println("错误:", err)
		os.Exit(1)
	}

	if resp.Message != "" {
		fmt.Println(resp.Message)
	} else {
		fmt.Println("出院办理成功！")
		if resp.BedID != "" {
			fmt.Printf("  已释放床位: %s\n", resp.BedID)
		}
	}
}

func cmdStatus(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	dept := fs.String("dept", "", "科室名称 (必需)")

	fs.Usage = func() {
		fmt.Println(`查看科室床位状态

使用方式:
  bed-client status -dept <科室名称>

选项:
  -dept string    科室名称 (必需)

示例:
  bed-client status -dept 内科
`)
	}

	fs.Parse(args)

	if *dept == "" {
		fs.Usage()
		os.Exit(1)
	}

	resp, err := c.GetDepartmentStatus(*dept)
	if err != nil {
		fmt.Println("错误:", err)
		os.Exit(1)
	}

	fmt.Printf("科室: %s\n", resp.DepartmentName)
	fmt.Println()
	fmt.Println("床位状态:")
	fmt.Printf("  %-10s  %-10s  %-6s  %-6s\n", "病区", "类型", "总数", "可用")
	fmt.Println("  ------------------------------------")
	for _, s := range resp.BedStatuses {
		fmt.Printf("  %-10s  %-10s  %-6d  %-6d\n",
			s.WardName, bedTypeLabel(s.BedType), s.Total, s.Available)
	}
	fmt.Println()
	fmt.Printf("候床队列: %d 人\n", resp.QueueLength)
}

func cmdPatient(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("patient", flag.ExitOnError)
	patientID := fs.String("id", "", "患者ID (必需)")

	fs.Usage = func() {
		fmt.Println(`查看患者床位信息

使用方式:
  bed-client patient -id <患者ID>

选项:
  -id string    患者ID (必需)

示例:
  bed-client patient -id P001
`)
	}

	fs.Parse(args)

	if *patientID == "" {
		fs.Usage()
		os.Exit(1)
	}

	resp, err := c.GetPatientBed(*patientID)
	if err != nil {
		fmt.Println("错误:", err)
		os.Exit(1)
	}

	fmt.Printf("患者ID: %s\n", resp.PatientID)
	fmt.Printf("患者姓名: %s\n", resp.PatientName)

	if resp.InQueue {
		fmt.Println()
		fmt.Println("状态: 候床队列中")
		fmt.Printf("  科室: %s\n", resp.DepartmentName)
		fmt.Printf("  队列位置: 第 %d 位\n", resp.QueuePosition)
	} else {
		fmt.Println()
		fmt.Println("状态: 已入院")
		fmt.Printf("  科室: %s\n", resp.DepartmentName)
		fmt.Printf("  病区: %s\n", resp.WardName)
		fmt.Printf("  床位: %s\n", resp.BedID)
		fmt.Printf("  类型: %s\n", bedTypeLabel(resp.BedType))
	}
}

func cmdQueue(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("queue", flag.ExitOnError)
	dept := fs.String("dept", "", "科室名称 (必需)")

	fs.Usage = func() {
		fmt.Println(`查看候床队列

使用方式:
  bed-client queue -dept <科室名称>

选项:
  -dept string    科室名称 (必需)

示例:
  bed-client queue -dept 内科
`)
	}

	fs.Parse(args)

	if *dept == "" {
		fs.Usage()
		os.Exit(1)
	}

	resp, err := c.GetQueueStatus(*dept)
	if err != nil {
		fmt.Println("错误:", err)
		os.Exit(1)
	}

	fmt.Printf("科室: %s\n", resp.DepartmentName)
	fmt.Printf("候床队列: %d 人\n", resp.QueueLength)

	if len(resp.QueueItems) > 0 {
		fmt.Println()
		fmt.Printf("  %-6s  %-10s  %-15s  %-10s\n", "位置", "患者ID", "姓名", "偏好类型")
		fmt.Println("  --------------------------------------------------")
		for _, item := range resp.QueueItems {
			fmt.Printf("  %-6d  %-10s  %-15s  %-10s\n",
				item.Position, item.PatientID, item.PatientName, bedTypeLabel(item.PreferredType))
		}
	}
}

func cmdTransfer(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("transfer", flag.ExitOnError)
	patientID := fs.String("id", "", "患者ID (必需)")
	targetDept := fs.String("dept", "", "目标科室 (必需)")
	targetWard := fs.String("ward", "", "目标病区 (可选)")
	preferred := fs.String("prefer", "single", "偏好床位类型")

	fs.Usage = func() {
		fmt.Println(`患者转科

说明: 转科会先在原科室办理出院释放床位，再到新科室办理入院重新分配。

使用方式:
  bed-client transfer -id <患者ID> -dept <目标科室> [选项]

选项:
  -id string      患者ID (必需)
  -dept string    目标科室 (必需)
  -ward string    目标病区 (可选)
  -prefer string  偏好床位类型 (默认 "single")

示例:
  bed-client transfer -id P001 -dept 外科 -prefer single
`)
	}

	fs.Parse(args)

	if *patientID == "" || *targetDept == "" {
		fs.Usage()
		os.Exit(1)
	}

	bedType, err := parseBedType(*preferred)
	if err != nil {
		fmt.Println("错误:", err)
		os.Exit(1)
	}

	req := client.TransferRequest{
		PatientID:     *patientID,
		TargetDept:    *targetDept,
		PreferredType: bedType,
	}
	if *targetWard != "" {
		req.TargetWard = targetWard
	}

	if err := c.TransferPatient(req); err != nil {
		fmt.Println("错误:", err)
		os.Exit(1)
	}

	fmt.Println("转科办理成功！")
	fmt.Printf("  目标科室: %s\n", *targetDept)

	patientResp, err := c.GetPatientBed(*patientID)
	if err == nil {
		if patientResp.InQueue {
			fmt.Printf("  状态: 加入目标科室候床队列，第 %d 位\n", patientResp.QueuePosition)
		} else {
			fmt.Printf("  新床位: %s\n", patientResp.BedID)
			fmt.Printf("  类型: %s\n", bedTypeLabel(patientResp.BedType))
		}
	}
}

func cmdConfigure(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("configure", flag.ExitOnError)
	dept := fs.String("dept", "", "科室名称 (必需)")
	wardsSpec := fs.String("wards", "", "病区配置列表 (必需)，格式: 病区名:单人间数:双人间数:三人间数:加床上限,...")

	fs.Usage = func() {
		fmt.Println(`配置科室床位布局 (管理员功能)

说明:
  配置指定科室的病区及各类型床位数量。可以随时调整配置。
  床位编号规则:
  - 普通床位: 科室-病区-类型+序号 (如 "内科-一区-单1")
  - 加床: 科室-病区-加-序号 (如 "内科-一区-加-1")

使用方式:
  bed-client configure -dept <科室名称> -wards <病区配置>

选项:
  -dept string      科室名称 (必需)
  -wards string     病区配置列表 (必需)
                    格式: 病区名:单人间:双人间:三人间:加床上限,...

示例:
  # 配置内科，有两个病区
  bed-client configure -dept 内科 -wards "一区:2:4:6:5,二区:3:6:9:5"
  
  # 配置说明:
  #   一区: 单人间2张, 双人间4张, 三人间6张, 加床最多5张
  #   二区: 单人间3张, 双人间6张, 三人间9张, 加床最多5张
`)
	}

	fs.Parse(args)

	if *dept == "" || *wardsSpec == "" {
		fs.Usage()
		os.Exit(1)
	}

	wardParts := strings.Split(*wardsSpec, ",")
	var wards []protocol.WardConfig

	for i, wp := range wardParts {
		wp = strings.TrimSpace(wp)
		if wp == "" {
			continue
		}
		parts := strings.Split(wp, ":")
		if len(parts) != 5 {
			fmt.Printf("错误: 病区配置格式错误 (第 %d 个病区)\n", i+1)
			os.Exit(1)
		}

		wardName := strings.TrimSpace(parts[0])
		single, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			fmt.Printf("错误: 单人间数量格式错误 (病区: %s)\n", wardName)
			os.Exit(1)
		}
		double, err := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err != nil {
			fmt.Printf("错误: 双人间数量格式错误 (病区: %s)\n", wardName)
			os.Exit(1)
		}
		triple, err := strconv.Atoi(strings.TrimSpace(parts[3]))
		if err != nil {
			fmt.Printf("错误: 三人间数量格式错误 (病区: %s)\n", wardName)
			os.Exit(1)
		}
		extraLimit, err := strconv.Atoi(strings.TrimSpace(parts[4]))
		if err != nil {
			fmt.Printf("错误: 加床上限格式错误 (病区: %s)\n", wardName)
			os.Exit(1)
		}

		wards = append(wards, protocol.WardConfig{
			WardName:      wardName,
			SingleBeds:    single,
			DoubleBeds:    double,
			TripleBeds:    triple,
			ExtraBedLimit: extraLimit,
		})
	}

	req := protocol.DepartmentConfigRequest{
		DepartmentName: *dept,
		Wards:          wards,
	}

	if err := c.ConfigureDepartment(req); err != nil {
		fmt.Println("错误:", err)
		os.Exit(1)
	}

	fmt.Printf("科室 [%s] 配置成功！\n", *dept)
	fmt.Println()
	fmt.Println("病区配置:")
	for _, w := range wards {
		fmt.Printf("  %s:\n", w.WardName)
		fmt.Printf("    单人间: %d 张\n", w.SingleBeds)
		fmt.Printf("    双人间: %d 张\n", w.DoubleBeds)
		fmt.Printf("    三人间: %d 张\n", w.TripleBeds)
		fmt.Printf("    加床上限: %d 张\n", w.ExtraBedLimit)
		fmt.Println()
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
	}

	globalFS := flag.NewFlagSet("global", flag.ContinueOnError)
	server := globalFS.String("server", "localhost:8080", "服务端地址")

	args := os.Args[1:]
	for i, arg := range args {
		if arg == "-server" || arg == "--server" {
			if i+1 < len(args) {
				*server = args[i+1]
			}
		}
	}

	cmdStart := 0
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			cmdStart = i
			break
		}
		if arg == "-server" || arg == "--server" {
			i++
			cmdStart = i + 1
		}
	}

	if cmdStart >= len(args) {
		printUsage()
	}

	cmd := args[cmdStart]
	cmdArgs := args[cmdStart+1:]

	c := client.NewAPIClient(*server)

	switch cmd {
	case "admit":
		cmdAdmit(c, cmdArgs)
	case "discharge":
		cmdDischarge(c, cmdArgs)
	case "status":
		cmdStatus(c, cmdArgs)
	case "patient":
		cmdPatient(c, cmdArgs)
	case "queue":
		cmdQueue(c, cmdArgs)
	case "transfer":
		cmdTransfer(c, cmdArgs)
	case "configure":
		cmdConfigure(c, cmdArgs)
	case "help", "-help", "--help":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n\n", cmd)
		printUsage()
	}
}
