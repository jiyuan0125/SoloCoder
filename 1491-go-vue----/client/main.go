package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"repair-platform/common"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	client := newClient()
	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "help":
		printUsage()
	case "create-repair":
		cmdCreateRepair(client, args)
	case "list-repairs":
		cmdListRepairs(client)
	case "assign-tech":
		cmdAssignTech(client, args)
	case "apply-spare":
		cmdApplySpare(client, args)
	case "complete-repair":
		cmdCompleteRepair(client, args)
	case "add-spare":
		cmdAddSpare(client, args)
	case "list-spares":
		cmdListSpares(client)
	case "update-spare-price":
		cmdUpdateSparePrice(client, args)
	case "list-pos":
		cmdListPOs(client)
	case "add-tech":
		cmdAddTech(client, args)
	case "list-techs":
		cmdListTechs(client)
	default:
		fmt.Printf("未知命令: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("家电维修平台 CLI")
	fmt.Println()
	fmt.Println("用法: repair-cli <命令> [参数]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  help                                      显示帮助")
	fmt.Println()
	fmt.Println("  create-repair <userID> <区域> <品类> <品牌> <型号> <故障描述> [地址]")
	fmt.Println("      品类: 空调|冰箱|洗衣机|电视|热水器|油烟机|燃气灶")
	fmt.Println()
	fmt.Println("  list-repairs                              列出所有维修请求")
	fmt.Println()
	fmt.Println("  assign-tech <请求ID>                      为订单派单")
	fmt.Println()
	fmt.Println("  apply-spare <请求ID> <师傅ID> <编码:数量[,编码:数量...]>")
	fmt.Println("      示例: apply-spare REQ00001 TECH01 AC01:2,AC02:1")
	fmt.Println()
	fmt.Println("  complete-repair <请求ID> <师傅ID> <故障原因> [使用备件] [退回备件]")
	fmt.Println("      使用备件格式: 编码:数量,...")
	fmt.Println("      退回备件格式: 编码:数量,...")
	fmt.Println("      示例: complete-repair REQ00001 TECH01 \"缺氟\" AC01:1 AC02:1")
	fmt.Println()
	fmt.Println("  add-spare <编码> <名称> <适用品类> <适用品牌> <单价> <库存>")
	fmt.Println()
	fmt.Println("  list-spares                               列出所有备件")
	fmt.Println()
	fmt.Println("  update-spare-price <编码> <新价格>")
	fmt.Println()
	fmt.Println("  list-pos                                  列出采购待办")
	fmt.Println()
	fmt.Println("  add-tech <ID> <姓名> <品类1,品类2...> <区域1,区域2...>")
	fmt.Println("      示例: add-tech TECH01 张师傅 空调,冰箱 朝阳区,海淀区")
	fmt.Println()
	fmt.Println("  list-techs                                列出所有师傅")
	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  SERVER_URL                                服务端地址 (默认 http://localhost:8080)")
}

func printJSON(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("输出失败: %v\n", err)
		return
	}
	fmt.Println(string(b))
}

func cmdCreateRepair(c *Client, args []string) {
	if len(args) < 6 {
		fmt.Println("参数不足")
		os.Exit(1)
	}
	userID := args[0]
	area := args[1]
	category := common.ApplianceCategory(args[2])
	brand := args[3]
	model := args[4]
	faultDesc := args[5]
	address := ""
	if len(args) > 6 {
		address = args[6]
	}

	resp, err := c.CreateRepair(userID, address, area, common.ApplianceInfo{
		Category: category,
		Brand:    brand,
		Model:    model,
	}, faultDesc)
	if err != nil {
		fmt.Printf("创建失败: %v\n", err)
		os.Exit(1)
	}
	printJSON(resp)
}

func cmdListRepairs(c *Client) {
	resp, err := c.ListRepairs()
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		os.Exit(1)
	}
	printJSON(resp)
}

func cmdAssignTech(c *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("需要请求ID")
		os.Exit(1)
	}
	resp, err := c.AssignTech(args[0])
	if err != nil {
		fmt.Printf("派单失败: %v\n", err)
		os.Exit(1)
	}
	printJSON(resp)
}

func parseSpareItems(s string) ([]common.ApplySpareItem, error) {
	var items []common.ApplySpareItem
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, ":", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("格式错误: %s", p)
		}
		qty, err := strconv.Atoi(kv[1])
		if err != nil {
			return nil, fmt.Errorf("数量错误: %s", kv[1])
		}
		items = append(items, common.ApplySpareItem{Code: kv[0], Quantity: qty})
	}
	return items, nil
}

func parseUsedItems(s string) []common.UsedPartItem {
	var items []common.UsedPartItem
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, ":", 2)
		if len(kv) != 2 {
			continue
		}
		qty, err := strconv.Atoi(kv[1])
		if err != nil {
			continue
		}
		items = append(items, common.UsedPartItem{Code: kv[0], Quantity: qty})
	}
	return items
}

func parseReturnItems(s string) []common.ReturnPartItem {
	var items []common.ReturnPartItem
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, ":", 2)
		if len(kv) != 2 {
			continue
		}
		qty, err := strconv.Atoi(kv[1])
		if err != nil {
			continue
		}
		items = append(items, common.ReturnPartItem{Code: kv[0], Quantity: qty})
	}
	return items
}

func cmdApplySpare(c *Client, args []string) {
	if len(args) < 3 {
		fmt.Println("参数不足")
		os.Exit(1)
	}
	reqID := args[0]
	techID := args[1]
	items, err := parseSpareItems(args[2])
	if err != nil {
		fmt.Printf("解析备件失败: %v\n", err)
		os.Exit(1)
	}
	resp, err := c.ApplySpare(reqID, techID, items)
	if err != nil {
		fmt.Printf("申请失败: %v\n", err)
		os.Exit(1)
	}
	printJSON(resp)
}

func cmdCompleteRepair(c *Client, args []string) {
	if len(args) < 3 {
		fmt.Println("参数不足")
		os.Exit(1)
	}
	reqID := args[0]
	techID := args[1]
	faultCause := args[2]

	var used []common.UsedPartItem
	if len(args) > 3 {
		used = parseUsedItems(args[3])
	}
	var ret []common.ReturnPartItem
	if len(args) > 4 {
		ret = parseReturnItems(args[4])
	}

	resp, err := c.CompleteRepair(reqID, techID, faultCause, used, ret)
	if err != nil {
		fmt.Printf("完成失败: %v\n", err)
		os.Exit(1)
	}
	printJSON(resp)
}

func cmdAddSpare(c *Client, args []string) {
	if len(args) < 6 {
		fmt.Println("参数不足")
		os.Exit(1)
	}
	code := args[0]
	name := args[1]
	cat := common.ApplianceCategory(args[2])
	brand := args[3]
	price, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		fmt.Printf("价格格式错误: %v\n", err)
		os.Exit(1)
	}
	stock, err := strconv.Atoi(args[5])
	if err != nil {
		fmt.Printf("库存格式错误: %v\n", err)
		os.Exit(1)
	}

	err = c.AddSpare(common.SparePart{
		Code:            code,
		Name:            name,
		ApplicableCat:   cat,
		ApplicableBrand: brand,
		UnitPrice:       price,
		StockQuantity:   stock,
	})
	if err != nil {
		fmt.Printf("添加失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("添加成功")
}

func cmdListSpares(c *Client) {
	resp, err := c.ListSpares()
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		os.Exit(1)
	}
	printJSON(resp)
}

func cmdUpdateSparePrice(c *Client, args []string) {
	if len(args) < 2 {
		fmt.Println("参数不足")
		os.Exit(1)
	}
	code := args[0]
	price, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		fmt.Printf("价格格式错误: %v\n", err)
		os.Exit(1)
	}
	err = c.UpdateSparePrice(code, price)
	if err != nil {
		fmt.Printf("更新失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("更新成功")
}

func cmdListPOs(c *Client) {
	resp, err := c.ListPOs()
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		os.Exit(1)
	}
	printJSON(resp)
}

func cmdAddTech(c *Client, args []string) {
	if len(args) < 4 {
		fmt.Println("参数不足")
		os.Exit(1)
	}
	id := args[0]
	name := args[1]
	catsRaw := strings.Split(args[2], ",")
	areasRaw := strings.Split(args[3], ",")

	var cats []common.ApplianceCategory
	for _, c := range catsRaw {
		c = strings.TrimSpace(c)
		if c != "" {
			cats = append(cats, common.ApplianceCategory(c))
		}
	}
	var areas []string
	for _, a := range areasRaw {
		a = strings.TrimSpace(a)
		if a != "" {
			areas = append(areas, a)
		}
	}

	err := c.AddTech(common.Technician{
		ID:           id,
		Name:         name,
		Specialties:  cats,
		ServiceAreas: areas,
	})
	if err != nil {
		fmt.Printf("添加失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("添加成功")
}

func cmdListTechs(c *Client) {
	resp, err := c.ListTechs()
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		os.Exit(1)
	}
	printJSON(resp)
}
