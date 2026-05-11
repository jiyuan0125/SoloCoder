package main

import (
	"flag"
	"fmt"
	"time"

	"supervision-log-system/common"
)

type DailyLogCommand struct{}

func (c *DailyLogCommand) Name() string {
	return "dailylog"
}

func (c *DailyLogCommand) Description() string {
	return "监理日志管理"
}

func (c *DailyLogCommand) Execute(args []string) error {
	if len(args) < 1 {
		c.printHelp()
		return nil
	}

	subCommand := args[0]
	args = args[1:]

	switch subCommand {
	case "draft":
		return c.generateDraft(args)
	case "submit":
		return c.submit(args)
	case "get":
		return c.get(args)
	case "--help":
		c.printHelp()
		return nil
	default:
		return fmt.Errorf("未知日志管理命令: %s", subCommand)
	}
}

func (c *DailyLogCommand) printHelp() {
	fmt.Println("监理日志管理命令:")
	fmt.Println("  draft       生成日志草稿")
	fmt.Println("  submit      提交日志")
	fmt.Println("  get         获取日志")
	fmt.Println("  --help      显示帮助")
}

func (c *DailyLogCommand) generateDraft(args []string) error {
	fs := flag.NewFlagSet("draft", flag.ExitOnError)
	supervisor := fs.String("supervisor", "", "监理人员")
	logDate := fs.String("date", time.Now().Format("2006-01-02"), "日志日期 (YYYY-MM-DD 格式)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *supervisor == "" {
		fs.PrintDefaults()
		return fmt.Errorf("缺少监理人员参数")
	}

	date, err := time.Parse("2006-01-02", *logDate)
	if err != nil {
		return fmt.Errorf("日期格式错误: %v", err)
	}

	req := common.GetDailyLogRequest{
		Supervisor: *supervisor,
		LogDate:    date,
	}

	respBytes, err := SendRequest("POST", "/api/dailylog/draft", req)
	if err != nil {
		return err
	}

	var resp common.SuccessResponse
	if err := ParseResponse(respBytes, &resp); err != nil {
		return err
	}

	fmt.Println("日志草稿生成成功:")
	PrintJSON(resp.Data)
	return nil
}

func (c *DailyLogCommand) submit(args []string) error {
	fs := flag.NewFlagSet("submit", flag.ExitOnError)
	supervisor := fs.String("supervisor", "", "监理人员")
	logDate := fs.String("date", time.Now().Format("2006-01-02"), "日志日期 (YYYY-MM-DD 格式)")
	additionalContent := fs.String("content", "", "补充内容")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *supervisor == "" {
		fs.PrintDefaults()
		return fmt.Errorf("缺少监理人员参数")
	}

	date, err := time.Parse("2006-01-02", *logDate)
	if err != nil {
		return fmt.Errorf("日期格式错误: %v", err)
	}

	req := common.SubmitDailyLogRequest{
		Supervisor:        *supervisor,
		LogDate:           date,
		AdditionalContent: *additionalContent,
	}

	respBytes, err := SendRequest("POST", "/api/dailylog", req)
	if err != nil {
		return err
	}

	var resp common.SuccessResponse
	if err := ParseResponse(respBytes, &resp); err != nil {
		return err
	}

	fmt.Println("日志提交成功:")
	PrintJSON(resp.Data)
	return nil
}

func (c *DailyLogCommand) get(args []string) error {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	supervisor := fs.String("supervisor", "", "监理人员")
	logDate := fs.String("date", time.Now().Format("2006-01-02"), "日志日期 (YYYY-MM-DD 格式)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *supervisor == "" {
		fs.PrintDefaults()
		return fmt.Errorf("缺少监理人员参数")
	}

	date, err := time.Parse("2006-01-02", *logDate)
	if err != nil {
		return fmt.Errorf("日期格式错误: %v", err)
	}

	req := common.GetDailyLogRequest{
		Supervisor: *supervisor,
		LogDate:    date,
	}

	respBytes, err := SendRequest("GET", "/api/dailylog", req)
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
