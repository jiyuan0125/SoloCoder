package main

import (
	"archivesystem/common"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
)

func printUsage() {
	fmt.Println("档案管理系统客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  archive-client [命令] [参数]")
	fmt.Println()
	fmt.Println("命令列表:")
	fmt.Println("  归档:")
	fmt.Println("    archive-create <title> <category> <secrecy> <archiver> [archive-date]")
	fmt.Println("    archive-list [--role=管理员/部门经理/普通员工] [--keyword=xxx] [--category=xxx]")
	fmt.Println("    archive-get <archive-id> [--role=xxx]")
	fmt.Println("    archive-export <filename> [--role=xxx]")
	fmt.Println()
	fmt.Println("  借阅:")
	fmt.Println("    borrow-apply <archive-id> <applicant> <reason> <expected-return>")
	fmt.Println("    borrow-approve <borrow-id> <approver> <role> <approve(1/0)>")
	fmt.Println("    borrow-confirm <borrow-id>")
	fmt.Println("    borrow-return <borrow-id> <returner>")
	fmt.Println("    borrow-list")
	fmt.Println("    borrow-overdue [current-date]")
	fmt.Println("    borrow-export <filename>")
	fmt.Println()
	fmt.Println("  销毁:")
	fmt.Println("    destroy-pending [current-date]")
	fmt.Println("    destroy-execute <archive-id> <destroyer>")
	fmt.Println("    destroy-records")
	fmt.Println()
	fmt.Println("全局参数:")
	fmt.Println("  --server=<url>  服务端地址，默认 http://localhost:8080")
}

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "服务端地址")
	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(*serverURL)
	cmd := args[0]
	params := args[1:]

	switch cmd {
	case "archive-create":
		cmdArchiveCreate(client, params)
	case "archive-list":
		cmdArchiveList(client, os.Args)
	case "archive-get":
		cmdArchiveGet(client, params, os.Args)
	case "archive-export":
		cmdArchiveExport(client, params, os.Args)
	case "borrow-apply":
		cmdBorrowApply(client, params)
	case "borrow-approve":
		cmdBorrowApprove(client, params)
	case "borrow-confirm":
		cmdBorrowConfirm(client, params)
	case "borrow-return":
		cmdBorrowReturn(client, params)
	case "borrow-list":
		cmdBorrowList(client)
	case "borrow-overdue":
		cmdBorrowOverdue(client, params)
	case "borrow-export":
		cmdBorrowExport(client, params)
	case "destroy-pending":
		cmdDestroyPending(client, params)
	case "destroy-execute":
		cmdDestroyExecute(client, params)
	case "destroy-records":
		cmdDestroyRecords(client)
	default:
		fmt.Printf("未知命令: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func cmdArchiveCreate(client *Client, params []string) {
	if len(params) < 4 {
		fmt.Println("参数不足")
		os.Exit(1)
	}

	req := common.CreateArchiveRequest{
		Title:        params[0],
		Category:     common.ArchiveCategory(params[1]),
		SecrecyLevel: common.ArchiveSecrecyLevel(params[2]),
		Archiver:     params[3],
	}
	if len(params) >= 5 {
		req.ArchiveDate = params[4]
	}

	var resp common.CreateArchiveResponse
	if err := client.postJSON("/api/archives/create", req, &resp); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("归档成功，档案编号: %s\n", resp.ArchiveID)
}

func parseFlagRole(args []string) string {
	for _, arg := range args {
		if len(arg) > 7 && arg[:7] == "--role=" {
			return arg[7:]
		}
	}
	return "管理员"
}

func parseFlagKeyword(args []string) string {
	for _, arg := range args {
		if len(arg) > 10 && arg[:10] == "--keyword=" {
			return arg[10:]
		}
	}
	return ""
}

func parseFlagCategory(args []string) string {
	for _, arg := range args {
		if len(arg) > 11 && arg[:11] == "--category=" {
			return arg[11:]
		}
	}
	return ""
}

func cmdArchiveList(client *Client, args []string) {
	req := common.ListArchivesRequest{
		UserRole: common.UserRole(parseFlagRole(args)),
		Keyword:  parseFlagKeyword(args),
		Category: common.ArchiveCategory(parseFlagCategory(args)),
	}

	var archives []common.Archive
	if err := client.postJSON("/api/archives/list", req, &archives); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	printJSON(archives)
}

func cmdArchiveGet(client *Client, params []string, args []string) {
	if len(params) < 1 {
		fmt.Println("请指定档案编号")
		os.Exit(1)
	}

	req := common.GetArchiveRequest{
		ArchiveID: params[0],
		UserRole:  common.UserRole(parseFlagRole(args)),
	}

	var arch common.Archive
	if err := client.postJSON("/api/archives/get", req, &arch); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	printJSON(arch)
}

func cmdArchiveExport(client *Client, params []string, args []string) {
	if len(params) < 1 {
		fmt.Println("请指定导出文件名")
		os.Exit(1)
	}

	req := common.ListArchivesRequest{
		UserRole: common.UserRole(parseFlagRole(args)),
	}

	if err := client.downloadCSV("/api/archives/export", req, params[0]); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("导出成功: %s\n", params[0])
}

func cmdBorrowApply(client *Client, params []string) {
	if len(params) < 4 {
		fmt.Println("参数不足")
		os.Exit(1)
	}

	req := common.ApplyBorrowRequest{
		ArchiveID:      params[0],
		Applicant:      params[1],
		Reason:         params[2],
		ExpectedReturn: params[3],
	}

	var resp map[string]int64
	if err := client.postJSON("/api/borrow/apply", req, &resp); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("申请成功，借阅记录ID: %d\n", resp["borrow_id"])
}

func cmdBorrowApprove(client *Client, params []string) {
	if len(params) < 4 {
		fmt.Println("参数不足")
		os.Exit(1)
	}

	borrowID, _ := strconv.ParseInt(params[0], 10, 64)
	isApproved := params[3] == "1"

	req := common.ApproveBorrowRequest{
		BorrowID:   borrowID,
		Approver:   params[1],
		UserRole:   common.UserRole(params[2]),
		IsApproved: isApproved,
	}

	if err := client.postJSON("/api/borrow/approve", req, nil); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("审批成功")
}

func cmdBorrowConfirm(client *Client, params []string) {
	if len(params) < 1 {
		fmt.Println("请指定借阅记录ID")
		os.Exit(1)
	}

	borrowID, _ := strconv.ParseInt(params[0], 10, 64)
	req := map[string]int64{"borrow_id": borrowID}

	if err := client.postJSON("/api/borrow/confirm", req, nil); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("确认借阅成功")
}

func cmdBorrowReturn(client *Client, params []string) {
	if len(params) < 2 {
		fmt.Println("参数不足")
		os.Exit(1)
	}

	borrowID, _ := strconv.ParseInt(params[0], 10, 64)
	req := common.ReturnArchiveRequest{
		BorrowID: borrowID,
		Returner: params[1],
	}

	if err := client.postJSON("/api/borrow/return", req, nil); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("归还成功")
}

func cmdBorrowList(client *Client) {
	var records []common.BorrowRecord
	if err := client.getJSON("/api/borrow/list", &records); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	printJSON(records)
}

func cmdBorrowOverdue(client *Client, params []string) {
	req := common.ListPendingDestroyRequest{}
	if len(params) >= 1 {
		req.CurrentDate = params[0]
	}

	var reminders []common.OverdueReminder
	if err := client.postJSON("/api/borrow/overdue", req, &reminders); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	printJSON(reminders)
}

func cmdBorrowExport(client *Client, params []string) {
	if len(params) < 1 {
		fmt.Println("请指定导出文件名")
		os.Exit(1)
	}

	if err := client.downloadCSVGet("/api/borrow/export", params[0]); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("导出成功: %s\n", params[0])
}

func cmdDestroyPending(client *Client, params []string) {
	req := common.ListPendingDestroyRequest{}
	if len(params) >= 1 {
		req.CurrentDate = params[0]
	}

	var archives []common.Archive
	if err := client.postJSON("/api/destroy/pending", req, &archives); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	printJSON(archives)
}

func cmdDestroyExecute(client *Client, params []string) {
	if len(params) < 2 {
		fmt.Println("参数不足")
		os.Exit(1)
	}

	req := common.DestroyArchiveRequest{
		ArchiveID: params[0],
		Destroyer: params[1],
	}

	if err := client.postJSON("/api/destroy/execute", req, nil); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("销毁成功")
}

func cmdDestroyRecords(client *Client) {
	var records []common.DestroyRecord
	if err := client.getJSON("/api/destroy/records", &records); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	printJSON(records)
}

func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}
