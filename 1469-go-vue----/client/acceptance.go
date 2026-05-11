package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"supervision-log-system/common"
)

type AcceptanceCommand struct{}

func (c *AcceptanceCommand) Name() string {
	return "acceptance"
}

func (c *AcceptanceCommand) Description() string {
	return "验收管理"
}

func (c *AcceptanceCommand) Execute(args []string) error {
	if len(args) < 1 {
		c.printHelp()
		return nil
	}

	subCommand := args[0]
	args = args[1:]

	switch subCommand {
	case "create":
		return c.create(args)
	case "list":
		return c.list()
	case "rectify":
		return c.rectify(args)
	case "complete-rectify":
		return c.completeRectify(args)
	case "recheck":
		return c.recheck(args)
	case "--help":
		c.printHelp()
		return nil
	default:
		return fmt.Errorf("未知验收命令: %s", subCommand)
	}
}

func (c *AcceptanceCommand) printHelp() {
	fmt.Println("验收管理命令:")
	fmt.Println("  create          创建验收记录")
	fmt.Println("  list            列出所有验收记录")
	fmt.Println("  rectify         创建整改通知")
	fmt.Println("  complete-rectify 完成整改")
	fmt.Println("  recheck         复查验收")
	fmt.Println("  --help          显示帮助")
}

func (c *AcceptanceCommand) create(args []string) error {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	projectPart := fs.String("project-part", "", "验收部位")
	itemsJSON := fs.String("items", "[]", "验收项目清单 (JSON 格式)")
	conclusion := fs.String("conclusion", "", "验收结论")
	status := fs.String("status", "pending", "验收状态 (pending/passed/failed)")
	supervisorsJSON := fs.String("supervisors", "[]", "验收人员列表 (JSON 数组)")
	acceptanceDate := fs.String("date", time.Now().Format(time.RFC3339), "验收日期 (RFC3339 格式)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *projectPart == "" || *conclusion == "" {
		fmt.Println("必须提供所有必需参数")
		fs.PrintDefaults()
		return fmt.Errorf("缺少必需参数")
	}

	var items []common.AcceptanceItem
	if err := json.Unmarshal([]byte(*itemsJSON), &items); err != nil {
		return fmt.Errorf("验收项目清单格式错误: %v", err)
	}

	var supervisors []string
	if err := json.Unmarshal([]byte(*supervisorsJSON), &supervisors); err != nil {
		return fmt.Errorf("验收人员列表格式错误: %v", err)
	}

	date, err := time.Parse(time.RFC3339, *acceptanceDate)
	if err != nil {
		return fmt.Errorf("验收日期格式错误: %v", err)
	}

	req := common.CreateAcceptanceRecordRequest{
		ProjectPart:    *projectPart,
		Items:          items,
		Conclusion:     *conclusion,
		Status:         common.AcceptanceStatus(*status),
		Supervisors:    supervisors,
		AcceptanceDate: date,
	}

	respBytes, err := SendRequest("POST", "/api/acceptance", req)
	if err != nil {
		return err
	}

	var resp common.SuccessResponse
	if err := ParseResponse(respBytes, &resp); err != nil {
		return err
	}

	fmt.Println("验收记录创建成功:")
	PrintJSON(resp.Data)
	return nil
}

func (c *AcceptanceCommand) list() error {
	respBytes, err := SendRequest("GET", "/api/acceptance", nil)
	if err != nil {
		return err
	}

	var resp common.SuccessResponse
	if err := ParseResponse(respBytes, &resp); err != nil {
		return err
	}

	PrintJSON(resp.Data)
	return nil
}

func (c *AcceptanceCommand) rectify(args []string) error {
	fs := flag.NewFlagSet("rectify", flag.ExitOnError)
	acceptanceID := fs.String("acceptance-id", "", "验收记录ID")
	contractor := fs.String("contractor", "", "施工单位")
	requirements := fs.String("requirements", "", "整改要求")
	deadline := fs.String("deadline", "", "整改期限 (RFC3339 格式)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *acceptanceID == "" || *contractor == "" || *requirements == "" || *deadline == "" {
		fs.PrintDefaults()
		return fmt.Errorf("缺少必需参数")
	}

	dl, err := time.Parse(time.RFC3339, *deadline)
	if err != nil {
		return fmt.Errorf("整改期限格式错误: %v", err)
	}

	req := common.CreateRectificationNoticeRequest{
		AcceptanceID: *acceptanceID,
		Contractor:   *contractor,
		Requirements: *requirements,
		Deadline:     dl,
	}

	respBytes, err := SendRequest("POST", "/api/acceptance/rectification", req)
	if err != nil {
		return err
	}

	var resp common.SuccessResponse
	if err := ParseResponse(respBytes, &resp); err != nil {
		return err
	}

	fmt.Println("整改通知创建成功:")
	PrintJSON(resp.Data)
	return nil
}

func (c *AcceptanceCommand) completeRectify(args []string) error {
	fs := flag.NewFlagSet("complete-rectify", flag.ExitOnError)
	id := fs.String("id", "", "整改通知ID")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == "" {
		fs.PrintDefaults()
		return fmt.Errorf("缺少ID参数")
	}

	req := common.CompleteRectificationRequest{ID: *id}
	respBytes, err := SendRequest("POST", "/api/acceptance/rectification/complete", req)
	if err != nil {
		return err
	}

	var resp common.SuccessResponse
	if err := ParseResponse(respBytes, &resp); err != nil {
		return err
	}

	fmt.Println("整改完成")
	PrintJSON(resp.Data)
	return nil
}

func (c *AcceptanceCommand) recheck(args []string) error {
	fs := flag.NewFlagSet("recheck", flag.ExitOnError)
	acceptanceID := fs.String("acceptance-id", "", "验收记录ID")
	itemsJSON := fs.String("items", "[]", "验收项目清单 (JSON 格式)")
	conclusion := fs.String("conclusion", "", "验收结论")
	status := fs.String("status", "pending", "验收状态")
	supervisorsJSON := fs.String("supervisors", "[]", "验收人员列表")
	acceptanceDate := fs.String("date", time.Now().Format(time.RFC3339), "验收日期")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *acceptanceID == "" || *conclusion == "" {
		fs.PrintDefaults()
		return fmt.Errorf("缺少必需参数")
	}

	var items []common.AcceptanceItem
	if err := json.Unmarshal([]byte(*itemsJSON), &items); err != nil {
		return fmt.Errorf("验收项目清单格式错误: %v", err)
	}

	var supervisors []string
	if err := json.Unmarshal([]byte(*supervisorsJSON), &supervisors); err != nil {
		return fmt.Errorf("验收人员列表格式错误: %v", err)
	}

	date, err := time.Parse(time.RFC3339, *acceptanceDate)
	if err != nil {
		return fmt.Errorf("验收日期格式错误: %v", err)
	}

	req := common.RecheckAcceptanceRequest{
		AcceptanceID:   *acceptanceID,
		Items:          items,
		Conclusion:     *conclusion,
		Status:         common.AcceptanceStatus(*status),
		Supervisors:    supervisors,
		AcceptanceDate: date,
	}

	respBytes, err := SendRequest("POST", "/api/acceptance/recheck", req)
	if err != nil {
		return err
	}

	var resp common.SuccessResponse
	if err := ParseResponse(respBytes, &resp); err != nil {
		return err
	}

	fmt.Println("复查完成:")
	PrintJSON(resp.Data)
	return nil
}
