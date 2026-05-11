package main

import (
	"bytes"
	"constructionms/common"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const defaultServerURL = "http://localhost:8080"

var serverURL string

type Client struct {
	serverURL string
}

func NewClient(url string) *Client {
	return &Client{serverURL: url}
}

func (c *Client) doRequest(method, path string, body interface{}) (*common.Response, error) {
	var jsonBody []byte
	var err error

	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求参数失败: %v", err)
		}
	}

	req, err := http.NewRequest(method, c.serverURL+path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var response common.Response
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v (响应内容: %s)", err, string(bodyBytes))
	}

	return &response, nil
}

func (c *Client) CreateProject(req *common.CreateProjectRequest) error {
	resp, err := c.doRequest(http.MethodPost, "/projects", req)
	if err != nil {
		return err
	}

	if resp.Code != 200 {
		return fmt.Errorf(resp.Message)
	}

	var project common.Project
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &project)

	fmt.Printf("\n✓ 项目创建成功！\n")
	fmt.Printf("  项目编号: %s\n", project.ID)
	fmt.Printf("  项目名称: %s\n", project.Name)
	fmt.Printf("  建设单位: %s\n", project.Client)
	fmt.Printf("  承包单位: %s\n", project.Contractor)
	fmt.Printf("  合同金额: %.2f 元\n", float64(project.ContractFee)/100.0)
	fmt.Printf("  计划开工: %s\n", project.PlanStart.Format("2006-01-02"))
	fmt.Printf("  计划竣工: %s\n", project.PlanEnd.Format("2006-01-02"))
	fmt.Printf("  项目经理: %s\n", project.ProjectManager)
	fmt.Printf("\n  里程碑节点:\n")
	for _, m := range project.Milestones {
		fmt.Printf("    - %s (计划完成: %s)\n", m.Name, m.PlanFinishDate.Format("2006-01-02"))
	}
	return nil
}

func (c *Client) ListProjects() error {
	resp, err := c.doRequest(http.MethodGet, "/projects", nil)
	if err != nil {
		return err
	}

	if resp.Code != 200 {
		return fmt.Errorf(resp.Message)
	}

	var list common.ProjectListResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &list)

	if len(list.Projects) == 0 {
		fmt.Println("暂无项目")
		return nil
	}

	fmt.Println("\n项目列表:")
	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("%-12s %-25s %-15s %-10s %-10s %-10s\n", 
		"项目编号", "项目名称", "项目经理", "进度", "状态", "质量风险")
	fmt.Println(strings.Repeat("-", 100))

	for _, p := range list.Projects {
		riskStatus := "正常"
		if p.IsRiskProject {
			riskStatus = "\033[31m风险\033[0m"
		}
		fmt.Printf("%-12s %-25s %-15s %-10.1f %-10s %-10s\n",
			p.ID, truncate(p.Name, 24), truncate(p.ProjectManager, 14), 
			p.Progress*100, p.Status, riskStatus)
	}
	fmt.Println(strings.Repeat("-", 100))
	return nil
}

func (c *Client) GetProject(projectID string) error {
	resp, err := c.doRequest(http.MethodGet, "/projects/"+projectID, nil)
	if err != nil {
		return err
	}

	if resp.Code != 200 {
		return fmt.Errorf(resp.Message)
	}

	var detail common.ProjectDetailResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &detail)

	fmt.Printf("\n=== 项目详情 ===\n")
	fmt.Printf("项目编号: %s\n", detail.ID)
	fmt.Printf("项目名称: %s\n", detail.Name)
	fmt.Printf("建设单位: %s\n", detail.Client)
	fmt.Printf("承包单位: %s\n", detail.Contractor)
	fmt.Printf("合同金额: %.2f 元\n", float64(detail.ContractFee)/100.0)
	fmt.Printf("计划开工: %s\n", detail.PlanStart.Format("2006-01-02"))
	fmt.Printf("计划竣工: %s\n", detail.PlanEnd.Format("2006-01-02"))
	fmt.Printf("项目经理: %s\n", detail.ProjectManager)
	fmt.Printf("整体进度: %.1f%%\n", detail.OverallProgress*100)
	
	riskStatus := "正常"
	if detail.IsRiskProject {
		riskStatus = "\033[31m质量风险项目（不合格检查超过5次）\033[0m"
	}
	fmt.Printf("质量状态: %s\n\n", riskStatus)

	fmt.Println("--- 里程碑进度 ---")
	for _, m := range detail.Milestones {
		delayStatus := ""
		if m.IsDelayed {
			delayStatus = " \033[31m[已延期]\033[0m"
		}
		actualDate := "-"
		if m.ActualFinishDate != nil {
			actualDate = m.ActualFinishDate.Format("2006-01-02")
		}
		fmt.Printf("%s: 进度 %d%%  (计划: %s, 实际: %s)%s\n", 
			m.Name, m.Progress, m.PlanFinishDate.Format("2006-01-02"), actualDate, delayStatus)
	}

	fmt.Println("\n--- 成本预算 ---")
	for _, b := range detail.BudgetSummary {
		budgetStr := fmt.Sprintf("%.2f", float64(b.Budget)/100.0)
		actualStr := fmt.Sprintf("%.2f", float64(b.Actual)/100.0)
		if b.OverBudget {
			fmt.Printf("\033[31m%s: 预算 %s 元, 实际 %s 元 [超预算]\033[0m\n",
				b.Name, budgetStr, actualStr)
		} else {
			fmt.Printf("%s: 预算 %s 元, 实际 %s 元\n",
				b.Name, budgetStr, actualStr)
		}
	}

	if len(detail.QualityChecks) > 0 {
		fmt.Println("\n--- 质量检查 ---")
		for _, q := range detail.QualityChecks {
			result := "合格"
			if !q.IsPass {
				result = "\033[31m不合格\033[0m"
			}
			closed := "已关闭"
			if !q.Closed {
				closed = "\033[33m待整改\033[0m"
			}
			fmt.Printf("%s - %s [%s] 检查人: %s 日期: %s [%s]\n",
				q.ID, q.CheckItem, result, q.Inspector, 
				q.CheckDate.Format("2006-01-02"), closed)
			
			if q.Rectification != nil {
				rectResult := "合格"
				if !q.Rectification.Passed {
					rectResult = "整改不合格"
				}
				fmt.Printf("  整改: %s 整改人: %s 日期: %s [%s]\n",
					q.Rectification.Result, q.Rectification.Inspector,
					q.Rectification.Date.Format("2006-01-02"), rectResult)
			}
		}
	}
	return nil
}

func (c *Client) UpdateProgress(projectID string, req *common.UpdateProgressRequest) error {
	resp, err := c.doRequest(http.MethodPost, "/projects/"+projectID+"/progress", req)
	if err != nil {
		return err
	}

	if resp.Code != 200 {
		return fmt.Errorf(resp.Message)
	}

	fmt.Printf("\n✓ 进度更新成功！\n")
	fmt.Printf("  项目: %s\n", projectID)
	fmt.Printf("  里程碑: %s\n", req.MilestoneName)
	fmt.Printf("  进度: %d%%\n", req.Progress)
	return nil
}

func (c *Client) AddExpense(projectID string, req *common.AddExpenseRequest) error {
	resp, err := c.doRequest(http.MethodPost, "/projects/"+projectID+"/expenses", req)
	if err != nil {
		return err
	}

	if resp.Code != 200 {
		return fmt.Errorf(resp.Message)
	}

	var result map[string]interface{}
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)

	fmt.Printf("\n✓ 支出记录添加成功！\n")
	fmt.Printf("  类别: %s\n", req.Category)
	fmt.Printf("  金额: %.2f 元\n", float64(req.Amount)/100.0)
	fmt.Printf("  日期: %s\n", req.Date)
	fmt.Printf("  凭证号: %s\n", req.VoucherNo)

	if overBudget, ok := result["over_budget"].(bool); ok && overBudget {
		fmt.Printf("\n\033[31m⚠ 警告: 该类别已超预算\033[0m\n")
	}
	return nil
}

func (c *Client) AddQualityCheck(projectID string, req *common.AddQualityCheckRequest) error {
	resp, err := c.doRequest(http.MethodPost, "/projects/"+projectID+"/quality", req)
	if err != nil {
		return err
	}

	if resp.Code != 200 {
		return fmt.Errorf(resp.Message)
	}

	var result map[string]interface{}
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)

	fmt.Printf("\n✓ 质量检查记录添加成功！\n")
	fmt.Printf("  检查项目: %s\n", req.CheckItem)
	fmt.Printf("  检查结果: %s\n", req.Result)
	fmt.Printf("  检查人: %s\n", req.Inspector)
	fmt.Printf("  检查日期: %s\n", req.CheckDate)

	if isRisk, ok := result["is_risk_project"].(bool); ok && isRisk {
		fmt.Printf("\n\033[31m⚠ 警告: 该项目已标记为质量风险项目（不合格检查超过5次）\033[0m\n")
	}
	return nil
}

func (c *Client) AddRectification(projectID string, req *common.RectificationRequest) error {
	resp, err := c.doRequest(http.MethodPost, "/projects/"+projectID+"/rectification", req)
	if err != nil {
		return err
	}

	if resp.Code != 200 {
		return fmt.Errorf(resp.Message)
	}

	fmt.Printf("\n✓ 整改记录添加成功！\n")
	fmt.Printf("  检查ID: %s\n", req.CheckID)
	fmt.Printf("  整改结果: %s\n", req.Result)
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func printUsage() {
	fmt.Println("\n建筑工程项目管理系统 - 客户端命令行工具")
	fmt.Println("\n用法:")
	fmt.Println("  constructionms-client [命令] [选项]")
	fmt.Println("\n命令:")
	fmt.Println("  create-project    创建新项目")
	fmt.Println("  list-projects     列出所有项目")
	fmt.Println("  get-project       查看项目详情")
	fmt.Println("  update-progress   更新里程碑进度")
	fmt.Println("  add-expense       添加支出记录")
	fmt.Println("  add-quality       添加质量检查记录")
	fmt.Println("  rectify           提交整改记录")
	fmt.Println("\n全局选项:")
	fmt.Println("  --server          服务端地址 (默认: http://localhost:8080)")
	fmt.Println("  --help            显示帮助信息")
	fmt.Println("\n示例:")
	fmt.Println("  # 创建项目")
	fmt.Println("  constructionms-client create-project --name \"国贸中心\" --client \"某地产公司\" --contractor \"中建集团\" --contract-fee 500000000 --plan-start 2024-01-01 --plan-end 2026-12-31 --pm \"张三\"")
	fmt.Println("")
	fmt.Println("  # 查看项目列表")
	fmt.Println("  constructionms-client list-projects")
	fmt.Println("")
	fmt.Println("  # 查看项目详情")
	fmt.Println("  constructionms-client get-project --id PRJ-000001")
	fmt.Println("")
	fmt.Println("  # 更新进度")
	fmt.Println("  constructionms-client update-progress --id PRJ-000001 --milestone \"基础工程\" --progress 50")
	fmt.Println("")
	fmt.Println("  # 添加支出")
	fmt.Println("  constructionms-client add-expense --id PRJ-000001 --category \"材料费\" --amount 10000000 --date 2024-03-15 --voucher INV001")
	fmt.Println("")
	fmt.Println("  # 添加质量检查")
	fmt.Println("  constructionms-client add-quality --id PRJ-000001 --item \"钢筋绑扎\" --result \"合格\" --inspector \"李四\" --date 2024-03-15")
}

func main() {
	flag.StringVar(&serverURL, "server", defaultServerURL, "服务端地址")
	flag.Usage = printUsage

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	args := os.Args[1:]
	command := args[0]

	flag.CommandLine.Parse(os.Args[2:])

	client := NewClient(serverURL)

	switch command {
	case "create-project":
		createCmd := flag.NewFlagSet("create-project", flag.ExitOnError)
		name := createCmd.String("name", "", "项目名称")
		clientName := createCmd.String("client", "", "建设单位")
		contractor := createCmd.String("contractor", "", "承包单位")
		contractFeeStr := createCmd.String("contract-fee", "", "合同金额（元）")
		planStart := createCmd.String("plan-start", "", "计划开工日期 (YYYY-MM-DD)")
		planEnd := createCmd.String("plan-end", "", "计划竣工日期 (YYYY-MM-DD)")
		pm := createCmd.String("pm", "", "项目经理")
		createCmd.Parse(args[1:])

		if *name == "" || *contractFeeStr == "" || *planStart == "" || *planEnd == "" {
			fmt.Println("错误: 缺少必要参数")
			createCmd.PrintDefaults()
			os.Exit(1)
		}

		feeFloat, err := strconv.ParseFloat(*contractFeeStr, 64)
		if err != nil {
			fmt.Printf("错误: 合同金额格式无效: %v\n", err)
			os.Exit(1)
		}
		contractFee := int64(feeFloat * 100)

		req := &common.CreateProjectRequest{
			Name:          *name,
			Client:        *clientName,
			Contractor:    *contractor,
			ContractFee:   contractFee,
			PlanStart:     *planStart,
			PlanEnd:       *planEnd,
			ProjectManager: *pm,
		}

		if err := client.CreateProject(req); err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}

	case "list-projects":
		if err := client.ListProjects(); err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}

	case "get-project":
		getCmd := flag.NewFlagSet("get-project", flag.ExitOnError)
		id := getCmd.String("id", "", "项目编号")
		getCmd.Parse(args[1:])

		if *id == "" {
			fmt.Println("错误: 缺少项目编号")
			os.Exit(1)
		}

		if err := client.GetProject(*id); err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}

	case "update-progress":
		updateCmd := flag.NewFlagSet("update-progress", flag.ExitOnError)
		id := updateCmd.String("id", "", "项目编号")
		milestone := updateCmd.String("milestone", "", "里程碑名称 (基础工程/主体结构/装饰装修/竣工验收)")
		progressStr := updateCmd.String("progress", "", "进度百分比 (0-100)")
		actualFinish := updateCmd.String("actual-finish", "", "实际完成日期 (YYYY-MM-DD)")
		updateCmd.Parse(args[1:])

		if *id == "" || *milestone == "" || *progressStr == "" {
			fmt.Println("错误: 缺少必要参数")
			updateCmd.PrintDefaults()
			os.Exit(1)
		}

		progress, err := strconv.Atoi(*progressStr)
		if err != nil {
			fmt.Printf("错误: 进度格式无效: %v\n", err)
			os.Exit(1)
		}

		req := &common.UpdateProgressRequest{
			MilestoneName:    *milestone,
			Progress:         progress,
			ActualFinishDate: *actualFinish,
		}

		if err := client.UpdateProgress(*id, req); err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}

	case "add-expense":
		expenseCmd := flag.NewFlagSet("add-expense", flag.ExitOnError)
		id := expenseCmd.String("id", "", "项目编号")
		category := expenseCmd.String("category", "", "支出类别 (人工费/材料费/机械费/管理费)")
		amountStr := expenseCmd.String("amount", "", "支出金额（元）")
		date := expenseCmd.String("date", "", "支出日期 (YYYY-MM-DD)")
		voucher := expenseCmd.String("voucher", "", "凭证编号")
		expenseCmd.Parse(args[1:])

		if *id == "" || *category == "" || *amountStr == "" || *date == "" {
			fmt.Println("错误: 缺少必要参数")
			expenseCmd.PrintDefaults()
			os.Exit(1)
		}

		amountFloat, err := strconv.ParseFloat(*amountStr, 64)
		if err != nil {
			fmt.Printf("错误: 金额格式无效: %v\n", err)
			os.Exit(1)
		}
		amount := int64(amountFloat * 100)

		req := &common.AddExpenseRequest{
			Category:  *category,
			Amount:    amount,
			Date:      *date,
			VoucherNo: *voucher,
		}

		if err := client.AddExpense(*id, req); err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}

	case "add-quality":
		qualityCmd := flag.NewFlagSet("add-quality", flag.ExitOnError)
		id := qualityCmd.String("id", "", "项目编号")
		item := qualityCmd.String("item", "", "检查项目")
		result := qualityCmd.String("result", "", "检查结果 (合格/不合格)")
		inspector := qualityCmd.String("inspector", "", "检查人")
		date := qualityCmd.String("date", "", "检查日期 (YYYY-MM-DD)")
		qualityCmd.Parse(args[1:])

		if *id == "" || *item == "" || *result == "" || *inspector == "" || *date == "" {
			fmt.Println("错误: 缺少必要参数")
			qualityCmd.PrintDefaults()
			os.Exit(1)
		}

		req := &common.AddQualityCheckRequest{
			CheckItem: *item,
			Result:    *result,
			Inspector: *inspector,
			CheckDate: *date,
		}

		if err := client.AddQualityCheck(*id, req); err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}

	case "rectify":
		rectifyCmd := flag.NewFlagSet("rectify", flag.ExitOnError)
		id := rectifyCmd.String("id", "", "项目编号")
		checkID := rectifyCmd.String("check-id", "", "质量检查记录ID")
		result := rectifyCmd.String("result", "", "整改结果 (合格/不合格)")
		inspector := rectifyCmd.String("inspector", "", "整改人")
		date := rectifyCmd.String("date", "", "整改日期 (YYYY-MM-DD)")
		rectifyCmd.Parse(args[1:])

		if *id == "" || *checkID == "" || *result == "" || *inspector == "" || *date == "" {
			fmt.Println("错误: 缺少必要参数")
			rectifyCmd.PrintDefaults()
			os.Exit(1)
		}

		req := &common.RectificationRequest{
			CheckID:   *checkID,
			Result:    *result,
			Inspector: *inspector,
			Date:      *date,
		}

		if err := client.AddRectification(*id, req); err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Printf("错误: 未知命令 '%s'\n", command)
		printUsage()
		os.Exit(1)
	}
}
