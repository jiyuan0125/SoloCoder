package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"laboratory/common"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var serverURL = "http://localhost:8080"

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "服务端地址")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		return
	}

	cmd := args[0]
	switch cmd {
	case "sample-create":
		cmdSampleCreate(args[1:])
	case "sample-list":
		cmdSampleList(args[1:])
	case "sample-get":
		cmdSampleGet(args[1:])
	case "sample-assign":
		cmdSampleAssign(args[1:])
	case "sample-status":
		cmdSampleStatus(args[1:])
	case "sample-claim":
		cmdSampleClaim(args[1:])
	case "sample-overdue":
		cmdSampleOverdue(args[1:])

	case "cabinet-create":
		cmdCabinetCreate(args[1:])
	case "cabinet-list":
		cmdCabinetList(args[1:])

	case "item-create":
		cmdItemCreate(args[1:])
	case "item-list":
		cmdItemList(args[1:])
	case "item-assign":
		cmdItemAssign(args[1:])

	case "result-record":
		cmdResultRecord(args[1:])
	case "result-list":
		cmdResultList(args[1:])

	case "retest-apply":
		cmdRetestApply(args[1:])
	case "retest-list":
		cmdRetestList(args[1:])
	case "retest-approve":
		cmdRetestApprove(args[1:])

	case "report-gen":
		cmdReportGen(args[1:])
	case "report-list":
		cmdReportList(args[1:])
	case "report-issue":
		cmdReportIssue(args[1:])
	case "report-void":
		cmdReportVoid(args[1:])

	default:
		fmt.Printf("未知命令: %s\n", cmd)
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`实验室管理系统客户端

样品管理:
  sample-create -name <名称> -qty <数量> -customer <客户> -date <送样日期> -req <要求> -vol <体积>
  sample-list [-status <状态>]
  sample-get <样品ID>
  sample-assign <样品ID> <样品柜ID>
  sample-status <样品ID> <新状态> <操作人> [-remark <备注>]
  sample-claim <样品ID> <检测员>
  sample-overdue

样品柜管理:
  cabinet-create <ID> <容量> [--special]
  cabinet-list

检测项目:
  item-create -method <方法编号> -name <项目名> -steps <步骤> -std <判定标准> -type <数值型|判定型> [-unit <单位>] [-threshold <合格阈值>] [-retest]
  item-list
  item-assign <样品ID> <项目ID1,项目ID2,...>

检测结果:
  result-record <样品ID> <项目ID> <操作人> -type <数值型|判定型> [-value <数值>] [-unit <单位>] [-result <合格|不合格>] [-retest]
  result-list <样品ID>

复检管理:
  retest-apply <样品ID> <项目ID> <申请人> <原因>
  retest-list
  retest-approve <申请ID> <审批人> (--approve|--reject) [-remark <备注>]

报告管理:
  report-gen <样品ID>
  report-list
  report-issue <报告ID> <签发人>
  report-void <报告ID> <操作人> <原因>

全局选项:
  -server <地址>    服务端地址 (默认 http://localhost:8080)`)
}

func postJSON(path string, reqBody interface{}, respBody interface{}) error {
	data, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(serverURL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var respData common.Response
	if err := json.Unmarshal(body, &respData); err != nil {
		return err
	}

	if respData.Code != 0 {
		return fmt.Errorf("请求失败: %s", respData.Message)
	}

	if respBody != nil {
		respBytes, _ := json.Marshal(respData.Data)
		json.Unmarshal(respBytes, respBody)
	} else {
		if respData.Message != "" {
			fmt.Println(respData.Message)
		}
	}

	return nil
}

func getJSON(path string, respBody interface{}) error {
	resp, err := http.Get(serverURL + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var respData common.Response
	if err := json.Unmarshal(body, &respData); err != nil {
		return err
	}

	if respData.Code != 0 {
		return fmt.Errorf("请求失败: %s", respData.Message)
	}

	if respBody != nil {
		respBytes, _ := json.Marshal(respData.Data)
		json.Unmarshal(respBytes, respBody)
	}

	return nil
}

func cmdSampleCreate(args []string) {
	fs := flag.NewFlagSet("sample-create", flag.ExitOnError)
	name := fs.String("name", "", "样品名称")
	qty := fs.Int("qty", 1, "样品数量")
	customer := fs.String("customer", "", "送样客户")
	dateStr := fs.String("date", time.Now().Format("2006-01-02"), "送样日期 (YYYY-MM-DD)")
	req := fs.String("req", "", "检测要求")
	vol := fs.Float64("vol", 0, "单样体积")
	fs.Parse(args)

	if *name == "" || *customer == "" {
		fmt.Println("错误: 样品名称和客户不能为空")
		fs.Usage()
		return
	}

	date, err := time.Parse("2006-01-02", *dateStr)
	if err != nil {
		fmt.Printf("日期格式错误: %v\n", err)
		return
	}

	var result common.CreateSampleResponse
	err = postJSON("/samples", &common.CreateSampleRequest{
		Name:         *name,
		Quantity:     *qty,
		Customer:     *customer,
		DeliveryDate: date,
		Requirements: *req,
		Volume:       *vol,
	}, &result)

	if err != nil {
		fmt.Printf("创建失败: %v\n", err)
		return
	}

	fmt.Printf("样品创建成功\n样品ID: %s\n是否大样: %v\n", result.SampleID, result.IsLargeSample)
}

func cmdSampleList(args []string) {
	fs := flag.NewFlagSet("sample-list", flag.ExitOnError)
	status := fs.String("status", "", "状态筛选: 待检测|检测中|已检毕|已退回|已销毁")
	fs.Parse(args)

	path := "/samples"
	if *status != "" {
		path = "/samples?status=" + *status
	}

	var result []*common.SampleInfo
	if err := getJSON(path, &result); err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	fmt.Printf("共 %d 个样品:\n", len(result))
	for _, s := range result {
		fmt.Printf("  ID: %-20s 名称: %-15s 状态: %s\n", s.ID, s.Name, s.Status)
	}
}

func cmdSampleGet(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: sample-get <样品ID>")
		return
	}

	var result common.SampleInfo
	if err := getJSON("/samples/"+args[0], &result); err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	fmt.Printf("样品ID: %s\n名称: %s\n数量: %d\n客户: %s\n送样日期: %s\n检测要求: %s\n体积: %.2f\n是否大样: %v\n状态: %s\n存储柜: %s\n",
		result.ID, result.Name, result.Quantity, result.Customer,
		result.DeliveryDate.Format("2006-01-02"), result.Requirements,
		result.Volume, result.IsLargeSample, result.Status, result.CabinetID)

	fmt.Println("\n状态日志:")
	for _, log := range result.StatusLogs {
		fmt.Printf("  %s [%s] %s - %s\n", log.Time.Format("2006-01-02 15:04:05"), log.Status, log.Operator, log.Remark)
	}
}

func cmdSampleAssign(args []string) {
	if len(args) < 2 {
		fmt.Println("用法: sample-assign <样品ID> <样品柜ID>")
		return
	}

	err := postJSON("/samples/assign", &common.AssignCabinetRequest{
		SampleID:  args[0],
		CabinetID: args[1],
	}, nil)

	if err != nil {
		fmt.Printf("分配失败: %v\n", err)
		return
	}
}

func cmdSampleStatus(args []string) {
	fs := flag.NewFlagSet("sample-status", flag.ExitOnError)
	remark := fs.String("remark", "", "备注")
	fs.Parse(args)

	if len(fs.Args()) < 3 {
		fmt.Println("用法: sample-status <样品ID> <新状态> <操作人> [-remark <备注>]")
		return
	}

	err := postJSON("/samples/status", &common.UpdateSampleStatusRequest{
		SampleID:  fs.Args()[0],
		NewStatus: common.SampleStatus(fs.Args()[1]),
		Operator:  fs.Args()[2],
		Remark:    *remark,
	}, nil)

	if err != nil {
		fmt.Printf("更新失败: %v\n", err)
		return
	}
}

func cmdSampleClaim(args []string) {
	if len(args) < 2 {
		fmt.Println("用法: sample-claim <样品ID> <检测员>")
		return
	}

	err := postJSON("/samples/claim", &common.ClaimSampleRequest{
		SampleID:   args[0],
		TesterName: args[1],
	}, nil)

	if err != nil {
		fmt.Printf("领取失败: %v\n", err)
		return
	}
}

func cmdSampleOverdue(args []string) {
	var result []*common.SampleInfo
	if err := getJSON("/samples/overdue", &result); err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	if len(result) == 0 {
		fmt.Println("无超期未检测样品")
		return
	}

	fmt.Printf("超期未检测样品 (%d 个):\n", len(result))
	for _, s := range result {
		fmt.Printf("  ID: %s  名称: %s  送样日期: %s\n", s.ID, s.Name, s.DeliveryDate.Format("2006-01-02"))
	}
}

func cmdCabinetCreate(args []string) {
	fs := flag.NewFlagSet("cabinet-create", flag.ExitOnError)
	special := fs.Bool("special", false, "特殊存储柜")
	fs.Parse(args)

	if len(fs.Args()) < 2 {
		fmt.Println("用法: cabinet-create <ID> <容量> [--special]")
		return
	}

	capacity, err := strconv.Atoi(fs.Args()[1])
	if err != nil {
		fmt.Println("容量必须是数字")
		return
	}

	err = postJSON("/cabinets", &common.CreateCabinetRequest{
		ID:        fs.Args()[0],
		Capacity:  capacity,
		IsSpecial: *special,
	}, nil)

	if err != nil {
		fmt.Printf("创建失败: %v\n", err)
		return
	}
}

func cmdCabinetList(args []string) {
	var result []*common.CabinetInfo
	if err := getJSON("/cabinets", &result); err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	fmt.Printf("共 %d 个样品柜:\n", len(result))
	for _, c := range result {
		special := "普通"
		if c.IsSpecial {
			special = "特殊"
		}
		fmt.Printf("  ID: %-10s 容量: %d/%d  类型: %s\n", c.ID, c.Used, c.Capacity, special)
	}
}

func cmdItemCreate(args []string) {
	fs := flag.NewFlagSet("item-create", flag.ExitOnError)
	method := fs.String("method", "", "标准方法编号")
	name := fs.String("name", "", "项目名称")
	steps := fs.String("steps", "", "检测步骤")
	std := fs.String("std", "", "判定标准")
	itemType := fs.String("type", "", "类型: 数值型|判定型")
	unit := fs.String("unit", "", "单位 (数值型)")
	threshold := fs.Float64("threshold", 0, "合格阈值 (数值型)")
	retest := fs.Bool("retest", false, "允许复检")
	fs.Parse(args)

	if *method == "" || *name == "" || *itemType == "" {
		fmt.Println("错误: 方法编号、项目名称和类型不能为空")
		fs.Usage()
		return
	}

	var result common.TestItemInfo
	err := postJSON("/test-items", &common.CreateTestItemRequest{
		MethodID:      *method,
		Name:          *name,
		Steps:         *steps,
		JudgmentStd:   *std,
		Type:          common.TestType(*itemType),
		Unit:          *unit,
		PassThreshold: *threshold,
		AllowRetest:   *retest,
	}, &result)

	if err != nil {
		fmt.Printf("创建失败: %v\n", err)
		return
	}

	fmt.Printf("检测项目创建成功\n项目ID: %s\n", result.ID)
}

func cmdItemList(args []string) {
	var result []*common.TestItemInfo
	if err := getJSON("/test-items", &result); err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	fmt.Printf("共 %d 个检测项目:\n", len(result))
	for _, i := range result {
		retest := "否"
		if i.AllowRetest {
			retest = "是"
		}
		fmt.Printf("  ID: %-10s 方法: %-10s 名称: %-15s 类型: %s  允许复检: %s\n",
			i.ID, i.MethodID, i.Name, i.Type, retest)
	}
}

func cmdItemAssign(args []string) {
	if len(args) < 2 {
		fmt.Println("用法: item-assign <样品ID> <项目ID1,项目ID2,...>")
		return
	}

	itemIDs := strings.Split(args[1], ",")

	err := postJSON("/test-items/assign", &common.AssignTestItemsRequest{
		SampleID: args[0],
		ItemIDs:  itemIDs,
	}, nil)

	if err != nil {
		fmt.Printf("分配失败: %v\n", err)
		return
	}
}

func cmdResultRecord(args []string) {
	fs := flag.NewFlagSet("result-record", flag.ExitOnError)
	resultType := fs.String("type", "", "类型: 数值型|判定型")
	value := fs.Float64("value", 0, "检测数值 (数值型)")
	unit := fs.String("unit", "", "单位 (数值型)")
	judgment := fs.String("result", "", "检测结果: 合格|不合格 (判定型)")
	retest := fs.Bool("retest", false, "是否复检")
	fs.Parse(args)

	if len(fs.Args()) < 3 {
		fmt.Println("用法: result-record <样品ID> <项目ID> <操作人> -type <数值型|判定型> [-value <数值>] [-unit <单位>] [-result <合格|不合格>] [-retest]")
		return
	}

	req := &common.RecordTestResultRequest{
		SampleID:    fs.Args()[0],
		ItemID:      fs.Args()[1],
		Operator:    fs.Args()[2],
		IsRetest:    *retest,
		NumericValue: *value,
		Unit:        *unit,
	}

	if *resultType == "判定型" {
		req.JudgmentResult = common.TestResultStatus(*judgment)
	}

	err := postJSON("/test-results", req, nil)
	if err != nil {
		fmt.Printf("录入失败: %v\n", err)
		return
	}
}

func cmdResultList(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: result-list <样品ID>")
		return
	}

	var result []*common.TestRecordInfo
	if err := getJSON("/test-results/"+args[0], &result); err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	fmt.Printf("样品 %s 的检测记录:\n", args[0])
	for _, r := range result {
		retest := ""
		if r.IsRetest {
			retest = " (复检)"
		}
		fmt.Printf("  项目: %-15s 状态: %s", r.ItemName, r.Status)
		if r.Status != common.TestResultPending {
			if r.NumericValue != 0 || r.Unit != "" {
				fmt.Printf("  数值: %.2f %s", r.NumericValue, r.Unit)
			}
			fmt.Printf("  判定: %s%s", r.JudgmentResult, retest)
			fmt.Printf("  操作人: %s  时间: %s", r.Operator, r.TestTime.Format("2006-01-02 15:04:05"))
		}
		fmt.Println()
	}
}

func cmdRetestApply(args []string) {
	if len(args) < 4 {
		fmt.Println("用法: retest-apply <样品ID> <项目ID> <申请人> <原因>")
		return
	}

	var result common.RetestRequestInfo
	err := postJSON("/retest", &common.ApplyRetestRequest{
		SampleID:  args[0],
		ItemID:    args[1],
		Applicant: args[2],
		Reason:    args[3],
	}, &result)

	if err != nil {
		fmt.Printf("申请失败: %v\n", err)
		return
	}

	fmt.Printf("复检申请已提交\n申请ID: %s\n状态: %s\n", result.ID, result.Status)
}

func cmdRetestList(args []string) {
	var result []*common.RetestRequestInfo
	if err := getJSON("/retest", &result); err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	fmt.Printf("共 %d 个复检申请:\n", len(result))
	for _, r := range result {
		fmt.Printf("  ID: %-10s 样品: %-15s 项目: %-10s 状态: %s  申请人: %s\n",
			r.ID, r.SampleID, r.ItemID, r.Status, r.Applicant)
	}
}

func cmdRetestApprove(args []string) {
	fs := flag.NewFlagSet("retest-approve", flag.ExitOnError)
	approve := fs.Bool("approve", false, "批准")
	reject := fs.Bool("reject", false, "拒绝")
	remark := fs.String("remark", "", "备注")
	fs.Parse(args)

	if len(fs.Args()) < 2 {
		fmt.Println("用法: retest-approve <申请ID> <审批人> (--approve|--reject) [-remark <备注>]")
		return
	}

	if !*approve && !*reject {
		fmt.Println("必须指定 --approve 或 --reject")
		return
	}

	err := postJSON("/retest/approve", &common.ApproveRetestRequest{
		RequestID: fs.Args()[0],
		Approved:  *approve,
		Approver:  fs.Args()[1],
		Remark:    *remark,
	}, nil)

	if err != nil {
		fmt.Printf("审批失败: %v\n", err)
		return
	}
}

func cmdReportGen(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: report-gen <样品ID>")
		return
	}

	var result common.ReportInfo
	err := postJSON("/reports", &common.GenerateReportRequest{SampleID: args[0]}, &result)

	if err != nil {
		fmt.Printf("生成失败: %v\n", err)
		return
	}

	fmt.Printf("报告已生成\n报告ID: %s\n状态: %s\n结论: %s\n", result.ID, result.Status, result.Conclusion)
}

func cmdReportList(args []string) {
	var result []*common.ReportInfo
	if err := getJSON("/reports", &result); err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	fmt.Printf("共 %d 份报告:\n", len(result))
	for _, r := range result {
		fmt.Printf("  ID: %-15s 样品: %-15s 状态: %-8s 结论: %s\n",
			r.ID, r.SampleID, r.Status, r.Conclusion)
	}
}

func cmdReportIssue(args []string) {
	if len(args) < 2 {
		fmt.Println("用法: report-issue <报告ID> <签发人>")
		return
	}

	err := postJSON("/reports/issue", &common.IssueReportRequest{
		ReportID: args[0],
		Issuer:   args[1],
	}, nil)

	if err != nil {
		fmt.Printf("签发失败: %v\n", err)
		return
	}
}

func cmdReportVoid(args []string) {
	if len(args) < 3 {
		fmt.Println("用法: report-void <报告ID> <操作人> <原因>")
		return
	}

	var result common.ReportInfo
	err := postJSON("/reports/void", &common.VoidReportRequest{
		ReportID: args[0],
		Operator: args[1],
		Reason:   args[2],
	}, &result)

	if err != nil {
		fmt.Printf("作废失败: %v\n", err)
		return
	}

	fmt.Printf("已作废原报告，新报告草稿ID: %s\n", result.ID)
	os.Exit(0)
}
