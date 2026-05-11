package main

import (
	"flag"
	"fmt"
	"time"

	"supervision-log-system/common"
)

type IssueCommand struct{}

func (c *IssueCommand) Name() string {
	return "issue"
}

func (c *IssueCommand) Description() string {
	return "问题跟踪管理"
}

func (c *IssueCommand) Execute(args []string) error {
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
	case "update-status":
		return c.updateStatus(args)
	case "check-overdue":
		return c.checkOverdue()
	case "--help":
		c.printHelp()
		return nil
	default:
		return fmt.Errorf("未知问题管理命令: %s", subCommand)
	}
}

func (c *IssueCommand) printHelp() {
	fmt.Println("问题跟踪管理命令:")
	fmt.Println("  create          创建问题")
	fmt.Println("  list            列出所有问题")
	fmt.Println("  update-status   更新问题状态")
	fmt.Println("  check-overdue   检查逾期问题")
	fmt.Println("  --help          显示帮助")
}

func (c *IssueCommand) create(args []string) error {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	severity := fs.String("severity", "general", "严重程度 (general/serious/major)")
	description := fs.String("description", "", "问题描述")
	discoveredDate := fs.String("discovered-date", time.Now().Format(time.RFC3339), "发现日期")
	responsibleUnit := fs.String("responsible-unit", "", "责任单位")
	rectificationReq := fs.String("rectification-req", "", "整改要求")
	planFinishDate := fs.String("plan-finish-date", "", "计划整改日期 (RFC3339 格式)")
	status := fs.String("status", "open", "问题状态 (open/in_progress/resolved/closed)")
	supervisor := fs.String("supervisor", "", "发现问题的监理人员")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *description == "" || *responsibleUnit == "" || *planFinishDate == "" || *supervisor == "" {
		fmt.Println("必须提供所有必需参数")
		fs.PrintDefaults()
		return fmt.Errorf("缺少必需参数")
	}

	discovered, err := time.Parse(time.RFC3339, *discoveredDate)
	if err != nil {
		return fmt.Errorf("发现日期格式错误: %v", err)
	}

	planFinish, err := time.Parse(time.RFC3339, *planFinishDate)
	if err != nil {
		return fmt.Errorf("计划整改日期格式错误: %v", err)
	}

	req := common.CreateIssueRequest{
		Severity:         common.Severity(*severity),
		Description:      *description,
		DiscoveredDate:   discovered,
		ResponsibleUnit:  *responsibleUnit,
		RectificationReq: *rectificationReq,
		PlanFinishDate:   planFinish,
		Status:           common.IssueStatus(*status),
		Supervisor:       *supervisor,
	}

	respBytes, err := SendRequest("POST", "/api/issues", req)
	if err != nil {
		return err
	}

	var resp common.SuccessResponse
	if err := ParseResponse(respBytes, &resp); err != nil {
		return err
	}

	fmt.Println("问题创建成功:")
	PrintJSON(resp.Data)
	return nil
}

func (c *IssueCommand) list() error {
	respBytes, err := SendRequest("GET", "/api/issues", nil)
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

func (c *IssueCommand) updateStatus(args []string) error {
	fs := flag.NewFlagSet("update-status", flag.ExitOnError)
	id := fs.String("id", "", "问题ID")
	status := fs.String("status", "", "新状态 (open/in_progress/resolved/closed)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == "" || *status == "" {
		fs.PrintDefaults()
		return fmt.Errorf("缺少必需参数")
	}

	req := common.UpdateIssueStatusRequest{
		ID:     *id,
		Status: common.IssueStatus(*status),
	}

	respBytes, err := SendRequest("PUT", "/api/issues/status", req)
	if err != nil {
		return err
	}

	var resp common.SuccessResponse
	if err := ParseResponse(respBytes, &resp); err != nil {
		return err
	}

	fmt.Println("问题状态更新成功")
	PrintJSON(resp.Data)
	return nil
}

func (c *IssueCommand) checkOverdue() error {
	respBytes, err := SendRequest("POST", "/api/issues/overdue", nil)
	if err != nil {
		return err
	}

	var resp common.SuccessResponse
	if err := ParseResponse(respBytes, &resp); err != nil {
		return err
	}

	fmt.Println("逾期问题检查完成")
	PrintJSON(resp.Data)
	return nil
}
