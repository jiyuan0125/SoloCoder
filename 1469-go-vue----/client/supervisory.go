package main

import (
	"flag"
	"fmt"
	"time"

	"supervision-log-system/common"
)

type SupervisoryCommand struct{}

func (c *SupervisoryCommand) Name() string {
	return "supervisory"
}

func (c *SupervisoryCommand) Description() string {
	return "旁站记录管理"
}

func (c *SupervisoryCommand) Execute(args []string) error {
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
	case "submit":
		return c.submit(args)
	case "supplement":
		return c.supplement(args)
	case "--help":
		c.printHelp()
		return nil
	default:
		return fmt.Errorf("未知旁站记录命令: %s", subCommand)
	}
}

func (c *SupervisoryCommand) printHelp() {
	fmt.Println("旁站记录管理命令:")
	fmt.Println("  create      创建旁站记录")
	fmt.Println("  list      列出所有旁站记录")
	fmt.Println("  submit    提交旁站记录")
	fmt.Println("  supplement 追加补充说明")
	fmt.Println("  --help    显示帮助")
	fmt.Println("\n示例:")
	fmt.Println("  client supervisory create --help")
}

func (c *SupervisoryCommand) create(args []string) error {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	projectPart := fs.String("project-part", "", "工程部位")
	process := fs.String("process", "", "施工工序")
	startTime := fs.String("start-time", "", "开始时间 (RFC3339 格式)")
	endTime := fs.String("end-time", "", "结束时间 (RFC3339 格式)")
	constructionDesc := fs.String("construction-desc", "", "施工情况描述")
	issuesFound := fs.String("issues-found", "", "发现问题")
	supervisor := fs.String("supervisor", "", "旁站监理人员")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *projectPart == "" || *process == "" || *startTime == "" || *endTime == "" || *supervisor == "" {
		fmt.Println("必须提供所有必需参数")
		fs.PrintDefaults()
		return fmt.Errorf("缺少必需参数")
	}

	start, err := time.Parse(time.RFC3339, *startTime)
	if err != nil {
		return fmt.Errorf("开始时间格式错误: %v", err)
	}

	end, err := time.Parse(time.RFC3339, *endTime)
	if err != nil {
		return fmt.Errorf("结束时间格式错误: %v", err)
	}

	req := common.CreateSupervisoryRecordRequest{
		ProjectPart:      *projectPart,
		Process:          *process,
		StartTime:        start,
		EndTime:          end,
		ConstructionDesc: *constructionDesc,
		IssuesFound:      *issuesFound,
		Supervisor:       *supervisor,
	}

	respBytes, err := SendRequest("POST", "/api/supervisory", req)
	if err != nil {
		return err
	}

	var resp common.SuccessResponse
	if err := ParseResponse(respBytes, &resp); err != nil {
		return err
	}

	fmt.Println("旁站记录创建成功:")
	PrintJSON(resp.Data)
	return nil
}

func (c *SupervisoryCommand) list() error {
	respBytes, err := SendRequest("GET", "/api/supervisory", nil)
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

func (c *SupervisoryCommand) submit(args []string) error {
	fs := flag.NewFlagSet("submit", flag.ExitOnError)
	id := fs.String("id", "", "旁站记录ID")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == "" {
		fs.PrintDefaults()
		return fmt.Errorf("缺少ID参数")
	}

	req := common.IDResponse{ID: *id}
	respBytes, err := SendRequest("POST", "/api/supervisory/submit", req)
	if err != nil {
		return err
	}

	var resp common.SuccessResponse
	if err := ParseResponse(respBytes, &resp); err != nil {
		return err
	}

	fmt.Println("旁站记录提交成功")
	PrintJSON(resp.Data)
	return nil
}

func (c *SupervisoryCommand) supplement(args []string) error {
	fs := flag.NewFlagSet("supplement", flag.ExitOnError)
	id := fs.String("id", "", "旁站记录ID")
	supplement := fs.String("supplement", "", "补充说明内容")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == "" || *supplement == "" {
		fs.PrintDefaults()
		return fmt.Errorf("缺少必需参数")
	}

	req := common.AddSupervisorySupplementRequest{
		ID:         *id,
		Supplement: *supplement,
	}

	respBytes, err := SendRequest("POST", "/api/supervisory/supplement", req)
	if err != nil {
		return err
	}

	var resp common.SuccessResponse
	if err := ParseResponse(respBytes, &resp); err != nil {
		return err
	}

	fmt.Println("补充说明添加成功")
	PrintJSON(resp.Data)
	return nil
}
