package main

import (
	"autorepair/common"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	if strings.HasSuffix(baseURL, "/") {
		baseURL = baseURL[:len(baseURL)-1]
	}
	return &Client{baseURL: baseURL}
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp common.Response
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Success {
		if apiResp.Message != "" {
			return nil, fmt.Errorf(apiResp.Message)
		}
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	return respBytes, nil
}

func printUsage() {
	fmt.Println("汽车维修管理系统 - 命令行客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client [全局选项] <命令> [子命令] [参数]")
	fmt.Println()
	fmt.Println("全局选项:")
	fmt.Println("  -server string  服务端地址 (默认 \"localhost:9002\")")
	fmt.Println("  -help, -h       显示帮助信息")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  order       工单管理")
	fmt.Println("  item        维修项目管理")
	fmt.Println("  part        配件管理")
	fmt.Println("  replenish   补货待办管理")
	fmt.Println()
	fmt.Println("使用 'client <命令> -help' 查看命令详细用法")
}

func printOrderUsage() {
	fmt.Println("工单管理")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client order <子命令> [参数]")
	fmt.Println()
	fmt.Println("子命令:")
	fmt.Println("  create          创建工单")
	fmt.Println("    参数:")
	fmt.Println("      -plate string      车牌号 (必填)")
	fmt.Println("      -name string       客户姓名 (必填)")
	fmt.Println("      -phone string      联系电话 (必填)")
	fmt.Println("      -desc string       问题描述")
	fmt.Println("      -fault string      故障类型 (默认 maintenance)")
	fmt.Println("        可选值: maintenance(保养), engine(发动机), transmission(变速箱),")
	fmt.Println("                  chassis(底盘), electrical(电气), ac(空调), body(钣金喷漆)")
	fmt.Println("  list            列出所有工单")
	fmt.Println("  get <id>        查看指定工单详情")
	fmt.Println("  assign <id>     接单")
	fmt.Println("  complete <id>   完成工单")
	fmt.Println("  cancel <id>     取消工单")
}

func printItemUsage() {
	fmt.Println("维修项目管理")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client item <子命令> [参数]")
	fmt.Println()
	fmt.Println("子命令:")
	fmt.Println("  create          创建维修项目")
	fmt.Println("    参数:")
	fmt.Println("      -order-id int      工单ID (必填)")
	fmt.Println("      -name string       项目名称 (必填)")
	fmt.Println("      -level string      技师等级 (默认 intermediate)")
	fmt.Println("        可选值: apprentice(学徒工), intermediate(中级工), senior(高级工)")
	fmt.Println("      -hours float       预估工时 (必填)")
	fmt.Println("  complete        完成维修项目")
	fmt.Println("    参数:")
	fmt.Println("      -item-id int       项目ID (必填)")
	fmt.Println("      -hours float       实际工时 (必填)")
	fmt.Println("      -reason string     超时原因 (工时超50%以上必填)")
	fmt.Println("  add-part        向项目添加配件")
	fmt.Println("    参数:")
	fmt.Println("      -item-id int       项目ID (必填)")
	fmt.Println("      -part-id int       配件ID (必填)")
	fmt.Println("      -qty int           数量 (必填)")
	fmt.Println("  return-part     退回配件")
	fmt.Println("    参数:")
	fmt.Println("      -used-id int       出库配件ID (必填)")
	fmt.Println("      -qty int           退回数量 (必填)")
}

func printPartUsage() {
	fmt.Println("配件管理")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client part <子命令> [参数]")
	fmt.Println()
	fmt.Println("子命令:")
	fmt.Println("  create          创建配件")
	fmt.Println("    参数:")
	fmt.Println("      -code string       配件编码 (必填)")
	fmt.Println("      -name string       配件名称 (必填)")
	fmt.Println("      -spec string       规格")
	fmt.Println("      -price int         单价(分) (默认 0)")
	fmt.Println("      -stock int         库存数量 (默认 0)")
	fmt.Println("      -warning int       预警值 (默认 0)")
	fmt.Println("  list            列出所有配件")
	fmt.Println("  get <id>        查看指定配件详情")
	fmt.Println("  set-price       设置配件价格")
	fmt.Println("    参数:")
	fmt.Println("      -part-id int       配件ID (必填)")
	fmt.Println("      -price int         新单价(分) (必填)")
	fmt.Println("  set-stock       设置配件库存")
	fmt.Println("    参数:")
	fmt.Println("      -part-id int       配件ID (必填)")
	fmt.Println("      -stock int         新库存数量 (必填)")
}

func printReplenishUsage() {
	fmt.Println("补货待办管理")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client replenish <子命令> [参数]")
	fmt.Println()
	fmt.Println("子命令:")
	fmt.Println("  list            列出所有补货待办")
	fmt.Println("  resolve <id>    标记补货待办为已处理")
}

func (c *Client) handleOrderCreate(args []string) error {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	plate := fs.String("plate", "", "车牌号")
	name := fs.String("name", "", "客户姓名")
	phone := fs.String("phone", "", "联系电话")
	desc := fs.String("desc", "", "问题描述")
	fault := fs.String("fault", "maintenance", "故障类型")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *plate == "" || *name == "" || *phone == "" {
		printOrderUsage()
		return fmt.Errorf("缺少必填参数")
	}

	req := common.CreateOrderRequest{
		PlateNumber:  *plate,
		CustomerName: *name,
		Phone:        *phone,
		Description:  *desc,
		FaultType:    common.FaultType(*fault),
	}

	respBytes, err := c.doRequest(http.MethodPost, "/api/orders", req)
	if err != nil {
		return err
	}

	var apiResp common.Response
	json.Unmarshal(respBytes, &apiResp)
	dataBytes, _ := json.MarshalIndent(apiResp.Data, "", "  ")
	fmt.Println("工单创建成功:")
	fmt.Println(string(dataBytes))
	return nil
}

func (c *Client) handleOrderList() error {
	respBytes, err := c.doRequest(http.MethodGet, "/api/orders", nil)
	if err != nil {
		return err
	}

	var apiResp common.Response
	json.Unmarshal(respBytes, &apiResp)
	dataBytes, _ := json.MarshalIndent(apiResp.Data, "", "  ")
	fmt.Println(string(dataBytes))
	return nil
}

func (c *Client) handleOrderGet(idStr string) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的ID: %s", idStr)
	}

	respBytes, err := c.doRequest(http.MethodGet, "/api/orders/"+strconv.FormatInt(id, 10), nil)
	if err != nil {
		return err
	}

	var apiResp common.Response
	json.Unmarshal(respBytes, &apiResp)
	dataBytes, _ := json.MarshalIndent(apiResp.Data, "", "  ")
	fmt.Println(string(dataBytes))
	return nil
}

func (c *Client) handleOrderAssign(idStr string) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的ID: %s", idStr)
	}

	req := common.AssignOrderRequest{OrderID: id}
	respBytes, err := c.doRequest(http.MethodPost, "/api/orders/assign", req)
	if err != nil {
		return err
	}

	var apiResp common.Response
	json.Unmarshal(respBytes, &apiResp)
	dataBytes, _ := json.MarshalIndent(apiResp.Data, "", "  ")
	fmt.Println("接单成功:")
	fmt.Println(string(dataBytes))
	return nil
}

func (c *Client) handleOrderComplete(idStr string) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的ID: %s", idStr)
	}

	req := common.CompleteOrderRequest{OrderID: id}
	respBytes, err := c.doRequest(http.MethodPost, "/api/orders/complete", req)
	if err != nil {
		return err
	}

	var apiResp common.Response
	json.Unmarshal(respBytes, &apiResp)
	dataBytes, _ := json.MarshalIndent(apiResp.Data, "", "  ")
	fmt.Println("工单完成:")
	fmt.Println(string(dataBytes))
	return nil
}

func (c *Client) handleOrderCancel(idStr string) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的ID: %s", idStr)
	}

	req := common.CancelOrderRequest{OrderID: id}
	respBytes, err := c.doRequest(http.MethodPost, "/api/orders/cancel", req)
	if err != nil {
		return err
	}

	var apiResp common.Response
	json.Unmarshal(respBytes, &apiResp)
	dataBytes, _ := json.MarshalIndent(apiResp.Data, "", "  ")
	fmt.Println("工单已取消:")
	fmt.Println(string(dataBytes))
	return nil
}

func (c *Client) handleItemCreate(args []string) error {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	orderID := fs.Int64("order-id", 0, "工单ID")
	name := fs.String("name", "", "项目名称")
	level := fs.String("level", "intermediate", "技师等级")
	hours := fs.Float64("hours", 0, "预估工时")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *orderID == 0 || *name == "" || *hours <= 0 {
		printItemUsage()
		return fmt.Errorf("缺少必填参数")
	}

	req := common.CreateRepairItemRequest{
		OrderID:         *orderID,
		Name:            *name,
		TechnicianLevel: common.TechnicianLevel(*level),
		EstimatedHours:  *hours,
	}

	respBytes, err := c.doRequest(http.MethodPost, "/api/repair-items", req)
	if err != nil {
		return err
	}

	var apiResp common.Response
	json.Unmarshal(respBytes, &apiResp)
	dataBytes, _ := json.MarshalIndent(apiResp.Data, "", "  ")
	fmt.Println("维修项目创建成功:")
	fmt.Println(string(dataBytes))
	return nil
}

func (c *Client) handleItemComplete(args []string) error {
	fs := flag.NewFlagSet("complete", flag.ExitOnError)
	itemID := fs.Int64("item-id", 0, "项目ID")
	hours := fs.Float64("hours", 0, "实际工时")
	reason := fs.String("reason", "", "超时原因")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *itemID == 0 || *hours <= 0 {
		printItemUsage()
		return fmt.Errorf("缺少必填参数")
	}

	req := common.CompleteRepairItemRequest{
		ItemID:         *itemID,
		ActualHours:    *hours,
		OvertimeReason: *reason,
	}

	respBytes, err := c.doRequest(http.MethodPost, "/api/repair-items/complete", req)
	if err != nil {
		return err
	}

	var apiResp common.Response
	json.Unmarshal(respBytes, &apiResp)
	dataBytes, _ := json.MarshalIndent(apiResp.Data, "", "  ")
	fmt.Println("维修项目完成:")
	fmt.Println(string(dataBytes))
	return nil
}

func (c *Client) handleItemAddPart(args []string) error {
	fs := flag.NewFlagSet("add-part", flag.ExitOnError)
	itemID := fs.Int64("item-id", 0, "项目ID")
	partID := fs.Int64("part-id", 0, "配件ID")
	qty := fs.Int64("qty", 0, "数量")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *itemID == 0 || *partID == 0 || *qty <= 0 {
		printItemUsage()
		return fmt.Errorf("缺少必填参数")
	}

	req := common.AddPartToItemRequest{
		ItemID:   *itemID,
		PartID:   *partID,
		Quantity: *qty,
	}

	respBytes, err := c.doRequest(http.MethodPost, "/api/used-parts", req)
	if err != nil {
		return err
	}

	var apiResp common.Response
	json.Unmarshal(respBytes, &apiResp)
	dataBytes, _ := json.MarshalIndent(apiResp.Data, "", "  ")
	fmt.Println("配件添加成功:")
	fmt.Println(string(dataBytes))
	return nil
}

func (c *Client) handleItemReturnPart(args []string) error {
	fs := flag.NewFlagSet("return-part", flag.ExitOnError)
	usedID := fs.Int64("used-id", 0, "出库配件ID")
	qty := fs.Int64("qty", 0, "退回数量")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *usedID == 0 || *qty <= 0 {
		printItemUsage()
		return fmt.Errorf("缺少必填参数")
	}

	req := common.ReturnPartRequest{
		UsedPartID: *usedID,
		Quantity:   *qty,
	}

	_, err := c.doRequest(http.MethodPost, "/api/used-parts/return", req)
	if err != nil {
		return err
	}

	fmt.Println("配件退回成功")
	return nil
}

func (c *Client) handlePartCreate(args []string) error {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	code := fs.String("code", "", "配件编码")
	name := fs.String("name", "", "配件名称")
	spec := fs.String("spec", "", "规格")
	price := fs.Int64("price", 0, "单价(分)")
	stock := fs.Int64("stock", 0, "库存数量")
	warning := fs.Int64("warning", 0, "预警值")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *code == "" || *name == "" {
		printPartUsage()
		return fmt.Errorf("缺少必填参数")
	}

	req := common.CreatePartRequest{
		Code:         *code,
		Name:         *name,
		Spec:         *spec,
		UnitPrice:    *price,
		StockQty:     *stock,
		WarningLevel: *warning,
	}

	respBytes, err := c.doRequest(http.MethodPost, "/api/parts", req)
	if err != nil {
		return err
	}

	var apiResp common.Response
	json.Unmarshal(respBytes, &apiResp)
	dataBytes, _ := json.MarshalIndent(apiResp.Data, "", "  ")
	fmt.Println("配件创建成功:")
	fmt.Println(string(dataBytes))
	return nil
}

func (c *Client) handlePartList() error {
	respBytes, err := c.doRequest(http.MethodGet, "/api/parts", nil)
	if err != nil {
		return err
	}

	var apiResp common.Response
	json.Unmarshal(respBytes, &apiResp)
	dataBytes, _ := json.MarshalIndent(apiResp.Data, "", "  ")
	fmt.Println(string(dataBytes))
	return nil
}

func (c *Client) handlePartGet(idStr string) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的ID: %s", idStr)
	}

	respBytes, err := c.doRequest(http.MethodGet, "/api/parts/"+strconv.FormatInt(id, 10), nil)
	if err != nil {
		return err
	}

	var apiResp common.Response
	json.Unmarshal(respBytes, &apiResp)
	dataBytes, _ := json.MarshalIndent(apiResp.Data, "", "  ")
	fmt.Println(string(dataBytes))
	return nil
}

func (c *Client) handlePartSetPrice(args []string) error {
	fs := flag.NewFlagSet("set-price", flag.ExitOnError)
	partID := fs.Int64("part-id", 0, "配件ID")
	price := fs.Int64("price", 0, "新单价(分)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *partID == 0 {
		printPartUsage()
		return fmt.Errorf("缺少必填参数")
	}

	req := common.UpdatePartPriceRequest{
		PartID:    *partID,
		UnitPrice: *price,
	}

	respBytes, err := c.doRequest(http.MethodPost, "/api/parts/price", req)
	if err != nil {
		return err
	}

	var apiResp common.Response
	json.Unmarshal(respBytes, &apiResp)
	dataBytes, _ := json.MarshalIndent(apiResp.Data, "", "  ")
	fmt.Println("价格更新成功:")
	fmt.Println(string(dataBytes))
	return nil
}

func (c *Client) handlePartSetStock(args []string) error {
	fs := flag.NewFlagSet("set-stock", flag.ExitOnError)
	partID := fs.Int64("part-id", 0, "配件ID")
	stock := fs.Int64("stock", 0, "新库存数量")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *partID == 0 {
		printPartUsage()
		return fmt.Errorf("缺少必填参数")
	}

	req := common.UpdatePartStockRequest{
		PartID:   *partID,
		StockQty: *stock,
	}

	respBytes, err := c.doRequest(http.MethodPost, "/api/parts/stock", req)
	if err != nil {
		return err
	}

	var apiResp common.Response
	json.Unmarshal(respBytes, &apiResp)
	dataBytes, _ := json.MarshalIndent(apiResp.Data, "", "  ")
	fmt.Println("库存更新成功:")
	fmt.Println(string(dataBytes))
	return nil
}

func (c *Client) handleReplenishList() error {
	respBytes, err := c.doRequest(http.MethodGet, "/api/replenishments", nil)
	if err != nil {
		return err
	}

	var apiResp common.Response
	json.Unmarshal(respBytes, &apiResp)
	dataBytes, _ := json.MarshalIndent(apiResp.Data, "", "  ")
	fmt.Println(string(dataBytes))
	return nil
}

func (c *Client) handleReplenishResolve(idStr string) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的ID: %s", idStr)
	}

	req := common.ResolveReplenishmentRequest{TodoID: id}
	respBytes, err := c.doRequest(http.MethodPost, "/api/replenishments/resolve", req)
	if err != nil {
		return err
	}

	var apiResp common.Response
	json.Unmarshal(respBytes, &apiResp)
	dataBytes, _ := json.MarshalIndent(apiResp.Data, "", "  ")
	fmt.Println("补货待办已处理:")
	fmt.Println(string(dataBytes))
	return nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverAddr := "localhost:9002"
	args := os.Args[1:]

	for i, arg := range args {
		if arg == "-server" && i+1 < len(args) {
			serverAddr = args[i+1]
			args = append(args[:i], args[i+2:]...)
			break
		} else if strings.HasPrefix(arg, "-server=") {
			serverAddr = arg[len("-server="):]
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	if args[0] == "-help" || args[0] == "-h" || args[0] == "--help" {
		printUsage()
		return
	}

	client := NewClient(serverAddr)
	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "order":
		if len(rest) == 0 {
			printOrderUsage()
			os.Exit(1)
		}
		subCmd := rest[0]
		subArgs := rest[1:]
		switch subCmd {
		case "create":
			if err := client.handleOrderCreate(subArgs); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "list":
			if err := client.handleOrderList(); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "get":
			if len(subArgs) != 1 {
				printOrderUsage()
				os.Exit(1)
			}
			if err := client.handleOrderGet(subArgs[0]); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "assign":
			if len(subArgs) != 1 {
				printOrderUsage()
				os.Exit(1)
			}
			if err := client.handleOrderAssign(subArgs[0]); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "complete":
			if len(subArgs) != 1 {
				printOrderUsage()
				os.Exit(1)
			}
			if err := client.handleOrderComplete(subArgs[0]); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "cancel":
			if len(subArgs) != 1 {
				printOrderUsage()
				os.Exit(1)
			}
			if err := client.handleOrderCancel(subArgs[0]); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "-help", "-h", "--help":
			printOrderUsage()
		default:
			fmt.Fprintf(os.Stderr, "未知的子命令: %s\n", subCmd)
			printOrderUsage()
			os.Exit(1)
		}

	case "item":
		if len(rest) == 0 {
			printItemUsage()
			os.Exit(1)
		}
		subCmd := rest[0]
		subArgs := rest[1:]
		switch subCmd {
		case "create":
			if err := client.handleItemCreate(subArgs); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "complete":
			if err := client.handleItemComplete(subArgs); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "add-part":
			if err := client.handleItemAddPart(subArgs); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "return-part":
			if err := client.handleItemReturnPart(subArgs); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "-help", "-h", "--help":
			printItemUsage()
		default:
			fmt.Fprintf(os.Stderr, "未知的子命令: %s\n", subCmd)
			printItemUsage()
			os.Exit(1)
		}

	case "part":
		if len(rest) == 0 {
			printPartUsage()
			os.Exit(1)
		}
		subCmd := rest[0]
		subArgs := rest[1:]
		switch subCmd {
		case "create":
			if err := client.handlePartCreate(subArgs); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "list":
			if err := client.handlePartList(); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "get":
			if len(subArgs) != 1 {
				printPartUsage()
				os.Exit(1)
			}
			if err := client.handlePartGet(subArgs[0]); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "set-price":
			if err := client.handlePartSetPrice(subArgs); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "set-stock":
			if err := client.handlePartSetStock(subArgs); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "-help", "-h", "--help":
			printPartUsage()
		default:
			fmt.Fprintf(os.Stderr, "未知的子命令: %s\n", subCmd)
			printPartUsage()
			os.Exit(1)
		}

	case "replenish":
		if len(rest) == 0 {
			printReplenishUsage()
			os.Exit(1)
		}
		subCmd := rest[0]
		subArgs := rest[1:]
		switch subCmd {
		case "list":
			if err := client.handleReplenishList(); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "resolve":
			if len(subArgs) != 1 {
				printReplenishUsage()
				os.Exit(1)
			}
			if err := client.handleReplenishResolve(subArgs[0]); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		case "-help", "-h", "--help":
			printReplenishUsage()
		default:
			fmt.Fprintf(os.Stderr, "未知的子命令: %s\n", subCmd)
			printReplenishUsage()
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "未知的命令: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
