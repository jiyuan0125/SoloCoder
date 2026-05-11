package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"merchant-mgmt-system/internal/core"
	"merchant-mgmt-system/pkg/common"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/")}
}

func (c *Client) do(method, path string, body interface{}, result interface{}) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if result != nil && len(data) > 0 {
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("解析响应失败: %v, 原始响应: %s", err, string(data))
		}
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	baseURL := os.Getenv("SERVER_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := NewClient(baseURL)

	cmd := strings.ToLower(os.Args[1])
	args := os.Args[2:]

	switch cmd {
	case "customer":
		handleCustomerCmd(client, args)
	case "follow":
		handleFollowCmd(client, args)
	case "contract":
		handleContractCmd(client, args)
	case "dashboard":
		handleDashboardCmd(client)
	case "help":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func handleCustomerCmd(client *Client, args []string) {
	if len(args) == 0 {
		fmt.Println("用法: client customer <list|get|create> [参数]")
		fmt.Println("\n子命令:")
		fmt.Println("  list [status]          列出客户，可选过滤状态")
		fmt.Println("  get <id>               获取单个客户详情")
		fmt.Println("  create <name> <contact> <phone> <area> <type> <move_in>")
		fmt.Println("                         创建新客户")
		os.Exit(1)
	}

	subCmd := strings.ToLower(args[0])
	switch subCmd {
	case "list":
		listCustomers(client, args[1:])
	case "get":
		getCustomer(client, args[1:])
	case "create":
		createCustomer(client, args[1:])
	default:
		fmt.Printf("未知客户子命令: %s\n", subCmd)
		os.Exit(1)
	}
}

func createCustomer(client *Client, args []string) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	name := fs.String("name", "", "客户名称")
	contact := fs.String("contact", "", "联系人")
	phone := fs.String("phone", "", "联系电话")
	area := fs.Float64("area", 0, "意向面积")
	bType := fs.String("type", "", "业态类型 (RESTAURANT/RETAIL/OFFICE/WAREHOUSE)")
	moveIn := fs.String("move-in", "", "预计入驻时间 (YYYY-MM-DD)")
	fs.Parse(args)

	if *name == "" || *contact == "" || *phone == "" || *area <= 0 || *bType == "" || *moveIn == "" {
		fmt.Println("请提供所有必需参数:")
		fmt.Println("  --name 客户名称")
		fmt.Println("  --contact 联系人")
		fmt.Println("  --phone 联系电话")
		fmt.Println("  --area 意向面积")
		fmt.Println("  --type 业态类型")
		fmt.Println("  --move-in 预计入驻时间")
		os.Exit(1)
	}

	bt, err := core.ValidateBusinessType(*bType)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	moveInDate, err := core.ParseDate(*moveIn)
	if err != nil {
		fmt.Printf("无效的日期格式: %v\n", err)
		os.Exit(1)
	}

	req := common.CreateCustomerRequest{
		CustomerName:   *name,
		ContactPerson:  *contact,
		ContactPhone:   *phone,
		IntentArea:     *area,
		IntentBusiness: bt,
		ExpectedMoveIn: moveInDate,
	}

	var result common.CreateCustomerResponse
	if err := client.do(http.MethodPost, "/api/customers", &req, &result); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("创建失败: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Println("创建成功!")
	printCustomer(result.Customer)
}

func listCustomers(client *Client, args []string) {
	path := "/api/customers"
	if len(args) > 0 {
		if _, err := core.ValidateStatus(args[0]); err != nil {
			fmt.Printf("无效的状态: %s\n", args[0])
			os.Exit(1)
		}
		path = "/api/customers?status=" + args[0]
	}

	var result common.ListCustomersResponse
	if err := client.do(http.MethodGet, path, nil, &result); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("查询失败: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Printf("共 %d 个客户:\n\n", len(result.Customers))
	for _, c := range result.Customers {
		printCustomer(c)
		fmt.Println()
	}
}

func getCustomer(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("请提供客户ID")
		os.Exit(1)
	}

	var result common.GetCustomerResponse
	if err := client.do(http.MethodGet, "/api/customers/"+args[0], nil, &result); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("查询失败: %s\n", result.Message)
		os.Exit(1)
	}

	printCustomer(result.Customer)
}

func printCustomer(c *common.Customer) {
	fmt.Printf("ID:          %s\n", c.ID)
	fmt.Printf("客户名称:    %s\n", c.CustomerName)
	fmt.Printf("联系人:      %s\n", c.ContactPerson)
	fmt.Printf("联系电话:    %s\n", c.ContactPhone)
	fmt.Printf("意向面积:    %.2f ㎡\n", c.IntentArea)
	fmt.Printf("意向业态:    %s\n", c.IntentBusiness)
	fmt.Printf("预计入驻:    %s\n", c.ExpectedMoveIn.Format("2006-01-02"))
	fmt.Printf("状态:        %s\n", c.Status)
	fmt.Printf("创建时间:    %s\n", c.CreatedAt.Format("2006-01-02 15:04:05"))
}

func handleFollowCmd(client *Client, args []string) {
	if len(args) == 0 {
		fmt.Println("用法: client follow <list|create|close> [参数]")
		fmt.Println("\n子命令:")
		fmt.Println("  list [customer_id]    列出跟进记录")
		fmt.Println("  create                创建跟进记录")
		fmt.Println("  close <id>            关闭跟进记录")
		os.Exit(1)
	}

	subCmd := strings.ToLower(args[0])
	switch subCmd {
	case "list":
		listFollows(client, args[1:])
	case "create":
		createFollow(client, args[1:])
	case "close":
		closeFollow(client, args[1:])
	default:
		fmt.Printf("未知跟进子命令: %s\n", subCmd)
		os.Exit(1)
	}
}

func createFollow(client *Client, args []string) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	customerID := fs.String("customer-id", "", "客户ID")
	method := fs.String("method", "", "跟进方式 (PHONE/WECHAT/MEETING/EMAIL)")
	content := fs.String("content", "", "跟进内容")
	nextPlan := fs.String("next-plan", "", "下次计划时间 (YYYY-MM-DD HH:MM 可选)")
	fs.Parse(args)

	if *customerID == "" || *method == "" || *content == "" {
		fmt.Println("请提供必需参数:")
		fmt.Println("  --customer-id 客户ID")
		fmt.Println("  --method 跟进方式")
		fmt.Println("  --content 跟进内容")
		fmt.Println("  --next-plan 下次计划时间 (可选)")
		os.Exit(1)
	}

	fm, err := core.ValidateFollowMethod(*method)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var nextPlanTime *time.Time
	if *nextPlan != "" {
		t, err := parseDateTime(*nextPlan)
		if err != nil {
			fmt.Printf("无效的计划时间: %v\n", err)
			os.Exit(1)
		}
		nextPlanTime = &t
	}

	req := common.CreateFollowRequest{
		CustomerID:   *customerID,
		Method:       fm,
		Content:      *content,
		NextPlanTime: nextPlanTime,
	}

	var result common.CreateFollowResponse
	if err := client.do(http.MethodPost, "/api/follows", &req, &result); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("创建失败: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Println("创建成功!")
	printFollow(result.Follow)
}

func listFollows(client *Client, args []string) {
	path := "/api/follows"
	if len(args) > 0 {
		path = "/api/follows?customer_id=" + args[0]
	}

	var result common.ListFollowsResponse
	if err := client.do(http.MethodGet, path, nil, &result); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("查询失败: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Printf("共 %d 条跟进记录:\n\n", len(result.Follows))
	for _, f := range result.Follows {
		printFollow(f)
		fmt.Println()
	}
}

func closeFollow(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("请提供跟进记录ID")
		os.Exit(1)
	}

	req := common.CloseFollowRequest{ID: args[0]}
	var result common.CloseFollowResponse
	if err := client.do(http.MethodPost, "/api/follows/close", &req, &result); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("关闭失败: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Println("跟进记录已关闭")
}

func printFollow(f *common.FollowRecord) {
	fmt.Printf("ID:          %s\n", f.ID)
	fmt.Printf("客户ID:      %s\n", f.CustomerID)
	fmt.Printf("跟进方式:    %s\n", f.Method)
	fmt.Printf("跟进内容:    %s\n", f.Content)
	if f.NextPlanTime != nil {
		fmt.Printf("下次计划:    %s\n", f.NextPlanTime.Format("2006-01-02 15:04"))
	}
	fmt.Printf("状态:        %s\n", f.Status)
	fmt.Printf("创建时间:    %s\n", f.CreatedAt.Format("2006-01-02 15:04:05"))
}

func handleContractCmd(client *Client, args []string) {
	if len(args) == 0 {
		fmt.Println("用法: client contract <list|create> [参数]")
		fmt.Println("\n子命令:")
		fmt.Println("  list                   列出合同")
		fmt.Println("  create                 创建合同")
		os.Exit(1)
	}

	subCmd := strings.ToLower(args[0])
	switch subCmd {
	case "list":
		listContracts(client)
	case "create":
		createContract(client, args[1:])
	default:
		fmt.Printf("未知合同子命令: %s\n", subCmd)
		os.Exit(1)
	}
}

func createContract(client *Client, args []string) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	customerID := fs.String("customer-id", "", "客户ID")
	area := fs.Float64("area", 0, "租赁面积")
	rate := fs.Float64("rate", 0, "月租金单价 (元/㎡)")
	startDate := fs.String("start", "", "租期开始日期 (YYYY-MM-DD)")
	endDate := fs.String("end", "", "租期结束日期 (YYYY-MM-DD)")
	freeDays := fs.Int("free-days", 0, "免租期天数")
	fs.Parse(args)

	if *customerID == "" || *area <= 0 || *rate <= 0 || *startDate == "" || *endDate == "" {
		fmt.Println("请提供必需参数:")
		fmt.Println("  --customer-id 客户ID")
		fmt.Println("  --area 租赁面积")
		fmt.Println("  --rate 月租金单价")
		fmt.Println("  --start 开始日期")
		fmt.Println("  --end 结束日期")
		fmt.Println("  --free-days 免租期天数 (可选，默认0)")
		os.Exit(1)
	}

	start, err := core.ParseDate(*startDate)
	if err != nil {
		fmt.Printf("无效的开始日期: %v\n", err)
		os.Exit(1)
	}

	end, err := core.ParseDate(*endDate)
	if err != nil {
		fmt.Printf("无效的结束日期: %v\n", err)
		os.Exit(1)
	}

	req := common.CreateContractRequest{
		CustomerID:      *customerID,
		LeaseArea:       *area,
		MonthlyRentRate: *rate,
		StartDate:       start,
		EndDate:         end,
		FreeRentDays:    *freeDays,
	}

	var result common.CreateContractResponse
	if err := client.do(http.MethodPost, "/api/contracts", &req, &result); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("签约失败: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Println("签约成功!")
	printContract(result.Contract)
}

func listContracts(client *Client) {
	var result common.ListContractsResponse
	if err := client.do(http.MethodGet, "/api/contracts", nil, &result); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("查询失败: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Printf("共 %d 个合同:\n\n", len(result.Contracts))
	for _, c := range result.Contracts {
		printContract(c)
		fmt.Println()
	}
}

func printContract(c *common.Contract) {
	fmt.Printf("ID:          %s\n", c.ID)
	fmt.Printf("客户ID:      %s\n", c.CustomerID)
	fmt.Printf("客户名称:    %s\n", c.CustomerName)
	fmt.Printf("租赁面积:    %.2f ㎡\n", c.LeaseArea)
	fmt.Printf("月租金单价:  %.2f 元/㎡\n", c.MonthlyRentRate)
	fmt.Printf("租期:        %s 至 %s\n", c.StartDate.Format("2006-01-02"), c.EndDate.Format("2006-01-02"))
	fmt.Printf("免租期:      %d 天\n", c.FreeRentDays)
	fmt.Printf("总租金:      %.2f 元\n", c.TotalRent)
	fmt.Printf("签约时间:    %s\n", c.CreatedAt.Format("2006-01-02 15:04:05"))
}

func handleDashboardCmd(client *Client) {
	var result common.GetDashboardResponse
	if err := client.do(http.MethodGet, "/api/dashboard", nil, &result); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("查询失败: %s\n", result.Message)
		os.Exit(1)
	}

	stats := result.Stats
	fmt.Println("==================== 数据看板 ====================")
	fmt.Printf("本月新增客户: %d\n", stats.MonthlyNewCustomers)
	fmt.Printf("本月签约数:   %d\n", stats.MonthlySignedContracts)
	fmt.Printf("签约率:       %.2f%%\n", stats.SigningRate)
	fmt.Printf("平均跟进次数: %.2f\n", stats.AverageFollowCount)
	fmt.Println("=================================================")
}

func parseDateTime(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04",
		"2006-01-02",
	}
	var lastErr error
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, s, time.Local)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

func parseDate(s string) (time.Time, error) {
	return core.ParseDate(s)
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func printUsage() {
	fmt.Println("招商管理系统 - 命令行客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client <command> [subcommand] [flags]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  customer    客户管理")
	fmt.Println("  follow      跟进管理")
	fmt.Println("  contract    合同管理")
	fmt.Println("  dashboard   数据看板")
	fmt.Println("  help        显示帮助")
	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  SERVER_URL  服务端地址 (默认 http://localhost:8080)")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  client customer create --name \"测试公司\" --contact \"张三\" --phone \"138-0000-0000\" --area 100 --type OFFICE --move-in 2026-06-01")
	fmt.Println("  client customer list")
	fmt.Println("  client follow create --customer-id <id> --method PHONE --content \"电话沟通良好\"")
	fmt.Println("  client contract list")
	fmt.Println("  client dashboard")
}
