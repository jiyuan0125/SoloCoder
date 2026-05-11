package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"pos-system/pkg/common"
)

var serverURL = "http://localhost:8080"

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Server URL")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n=== POS 门店管理系统客户端 ===")
		fmt.Println("1. 商品管理")
		fmt.Println("2. 会员管理")
		fmt.Println("3. POS收银")
		fmt.Println("4. 日结")
		fmt.Println("5. 盘点")
		fmt.Println("6. 入库")
		fmt.Println("0. 退出")
		fmt.Print("请选择操作: ")

		choice := readInput(reader)

		switch choice {
		case "1":
			productMenu(reader)
		case "2":
			memberMenu(reader)
		case "3":
			posCheckout(reader)
		case "4":
			dailyClosingMenu(reader)
		case "5":
			stocktakingMenu(reader)
		case "6":
			stockIn(reader)
		case "0":
			fmt.Println("退出系统...")
			return
		default:
			fmt.Println("无效选项，请重试")
		}
	}
}

func readInput(reader *bufio.Reader) string {
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func readInt(reader *bufio.Reader) int {
	input := readInput(reader)
	n, _ := strconv.Atoi(input)
	return n
}

func readInt64(reader *bufio.Reader) int64 {
	input := readInput(reader)
	n, _ := strconv.ParseInt(input, 10, 64)
	return n
}

func readFloat64(reader *bufio.Reader) float64 {
	input := readInput(reader)
	f, _ := strconv.ParseFloat(input, 64)
	return f
}

func productMenu(reader *bufio.Reader) {
	for {
		fmt.Println("\n=== 商品管理 ===")
		fmt.Println("1. 添加商品")
		fmt.Println("2. 查看商品列表")
		fmt.Println("3. 按条码查询商品")
		fmt.Println("0. 返回")
		fmt.Print("请选择: ")

		choice := readInput(reader)

		switch choice {
		case "1":
			createProduct(reader)
		case "2":
			listProducts()
		case "3":
			getProductByBarcode(reader)
		case "0":
			return
		default:
			fmt.Println("无效选项")
		}
	}
}

func createProduct(reader *bufio.Reader) {
	fmt.Print("商品名称: ")
	name := readInput(reader)
	fmt.Print("商品条码: ")
	barcode := readInput(reader)
	fmt.Print("单价(分): ")
	price := readInt64(reader)
	fmt.Print("库存数量: ")
	stockQty := readInt(reader)

	req := common.CreateProductReq{
		Name:     name,
		Barcode:  barcode,
		Price:    price,
		StockQty: stockQty,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/products", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Code != 0 {
		fmt.Printf("错误: %s\n", result.Message)
		return
	}

	dataMap, _ := result.Data.(map[string]interface{})
	fmt.Printf("商品创建成功，ID: %v\n", dataMap["id"])
}

func listProducts() {
	resp, err := http.Get(serverURL + "/api/products")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Code != 0 {
		fmt.Printf("错误: %s\n", result.Message)
		return
	}

	products, _ := result.Data.([]interface{})
	fmt.Println("\n商品列表:")
	for _, p := range products {
		prod := p.(map[string]interface{})
		fmt.Printf("ID: %v, 名称: %v, 条码: %v, 价格: %v分, 库存: %v\n",
			prod["id"], prod["name"], prod["barcode"], prod["price"], prod["stock_qty"])
	}
}

func getProductByBarcode(reader *bufio.Reader) {
	fmt.Print("输入条码: ")
	barcode := readInput(reader)

	resp, err := http.Get(serverURL + "/api/products/barcode?barcode=" + barcode)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Code != 0 {
		fmt.Printf("错误: %s\n", result.Message)
		return
	}

	prod, _ := result.Data.(map[string]interface{})
	fmt.Printf("ID: %v, 名称: %v, 条码: %v, 价格: %v分, 库存: %v\n",
		prod["id"], prod["name"], prod["barcode"], prod["price"], prod["stock_qty"])
}

func memberMenu(reader *bufio.Reader) {
	for {
		fmt.Println("\n=== 会员管理 ===")
		fmt.Println("1. 创建会员等级")
		fmt.Println("2. 创建会员")
		fmt.Println("0. 返回")
		fmt.Print("请选择: ")

		choice := readInput(reader)

		switch choice {
		case "1":
			createMemberLevel(reader)
		case "2":
			createMember(reader)
		case "0":
			return
		default:
			fmt.Println("无效选项")
		}
	}
}

func createMemberLevel(reader *bufio.Reader) {
	fmt.Print("等级名称: ")
	name := readInput(reader)
	fmt.Print("折扣率(0-1, 如0.9表示9折): ")
	rate := readFloat64(reader)

	req := common.CreateMemberLevelReq{
		Name:         name,
		DiscountRate: rate,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/member-levels", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Code != 0 {
		fmt.Printf("错误: %s\n", result.Message)
		return
	}

	dataMap, _ := result.Data.(map[string]interface{})
	fmt.Printf("会员等级创建成功，ID: %v\n", dataMap["id"])
}

func createMember(reader *bufio.Reader) {
	fmt.Print("会员名称: ")
	name := readInput(reader)
	fmt.Print("电话: ")
	phone := readInput(reader)
	fmt.Print("会员等级ID: ")
	levelID := readInput(reader)

	req := common.CreateMemberReq{
		Name:          name,
		Phone:         phone,
		MemberLevelID: levelID,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/members", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Code != 0 {
		fmt.Printf("错误: %s\n", result.Message)
		return
	}

	dataMap, _ := result.Data.(map[string]interface{})
	fmt.Printf("会员创建成功，ID: %v\n", dataMap["id"])
}

func posCheckout(reader *bufio.Reader) {
	fmt.Println("\n=== POS收银 ===")
	fmt.Print("收银员ID: ")
	cashierID := readInput(reader)
	fmt.Print("会员ID(可选，直接回车跳过): ")
	memberID := readInput(reader)
	fmt.Print("支付方式(cash/wechat/alipay): ")
	paymentMethod := readInput(reader)

	var items []common.AddOrderItem
	for {
		fmt.Print("输入商品条码(输入done完成): ")
		barcode := readInput(reader)
		if barcode == "done" {
			break
		}

		resp, err := http.Get(serverURL + "/api/products/barcode?barcode=" + barcode)
		if err != nil {
			fmt.Printf("查询商品失败: %v\n", err)
			continue
		}

		var result common.Response
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		if result.Code != 0 {
			fmt.Printf("未找到商品: %s\n", result.Message)
			continue
		}

		prod, _ := result.Data.(map[string]interface{})
		fmt.Printf("商品: %v, 价格: %v分, 库存: %v\n", prod["name"], prod["price"], prod["stock_qty"])

		fmt.Print("数量: ")
		qty := readInt(reader)

		items = append(items, common.AddOrderItem{
			ProductID: prod["id"].(string),
			Quantity:  qty,
		})
	}

	if len(items) == 0 {
		fmt.Println("没有选择任何商品")
		return
	}

	req := common.CreateOrderReq{
		CashierID:     cashierID,
		MemberID:      memberID,
		Items:         items,
		PaymentMethod: paymentMethod,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/orders", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Code != 0 {
		fmt.Printf("错误: %s\n", result.Message)
		return
	}

	dataMap, _ := result.Data.(map[string]interface{})
	fmt.Println("\n=== 结算成功 ===")
	fmt.Printf("订单号: %v\n", dataMap["order_id"])
	fmt.Printf("总金额: %v分\n", dataMap["total_amount"])
	fmt.Printf("折扣金额: %v分\n", dataMap["discount_amount"])
	fmt.Printf("实付金额: %v分\n", dataMap["payable_amount"])
	fmt.Printf("库存状态: %v\n", dataMap["stock_status"])
}

func dailyClosingMenu(reader *bufio.Reader) {
	for {
		fmt.Println("\n=== 日结管理 ===")
		fmt.Println("1. 创建日结")
		fmt.Println("2. 查看日结")
		fmt.Println("0. 返回")
		fmt.Print("请选择: ")

		choice := readInput(reader)

		switch choice {
		case "1":
			createDailyClosing(reader)
		case "2":
			getDailyClosing(reader)
		case "0":
			return
		default:
			fmt.Println("无效选项")
		}
	}
}

func createDailyClosing(reader *bufio.Reader) {
	fmt.Print("日期(YYYY-MM-DD): ")
	date := readInput(reader)

	req := common.DailyClosingReq{
		Date: date,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/daily-closing", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Code != 0 {
		fmt.Printf("错误: %s\n", result.Message)
		return
	}

	dataMap, _ := result.Data.(map[string]interface{})
	fmt.Println("\n=== 日结完成 ===")
	fmt.Printf("日结单号: %v\n", dataMap["closing_id"])
	fmt.Printf("日期: %v\n", dataMap["date"])
	fmt.Printf("总金额: %v分\n", dataMap["total_amount"])
	fmt.Printf("订单数: %v\n", dataMap["order_count"])

	stats, _ := dataMap["payment_stats"].([]interface{})
	fmt.Println("按支付方式统计:")
	for _, s := range stats {
		stat := s.(map[string]interface{})
		fmt.Printf("  %v: %v分 (%v笔)\n", stat["payment_method"], stat["total_amount"], stat["order_count"])
	}
}

func getDailyClosing(reader *bufio.Reader) {
	fmt.Print("日期(YYYY-MM-DD): ")
	date := readInput(reader)

	resp, err := http.Get(serverURL + "/api/daily-closing?date=" + date)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Code != 0 {
		fmt.Printf("错误: %s\n", result.Message)
		return
	}

	dataMap, _ := result.Data.(map[string]interface{})
	fmt.Println("\n=== 日结详情 ===")
	fmt.Printf("日结单号: %v\n", dataMap["closing_id"])
	fmt.Printf("日期: %v\n", dataMap["date"])
	fmt.Printf("总金额: %v分\n", dataMap["total_amount"])
	fmt.Printf("订单数: %v\n", dataMap["order_count"])
	fmt.Printf("完成时间: %v\n", dataMap["closed_at"])
}

func stocktakingMenu(reader *bufio.Reader) {
	for {
		fmt.Println("\n=== 盘点管理 ===")
		fmt.Println("1. 创建盘点任务")
		fmt.Println("2. 查看盘点任务列表")
		fmt.Println("3. 查看盘点任务详情")
		fmt.Println("4. 提交盘点结果")
		fmt.Println("5. 完成盘点")
		fmt.Println("0. 返回")
		fmt.Print("请选择: ")

		choice := readInput(reader)

		switch choice {
		case "1":
			createStocktaking(reader)
		case "2":
			listStocktakings()
		case "3":
			getStocktakingDetail(reader)
		case "4":
			submitStocktaking(reader)
		case "5":
			completeStocktaking(reader)
		case "0":
			return
		default:
			fmt.Println("无效选项")
		}
	}
}

func createStocktaking(reader *bufio.Reader) {
	fmt.Print("操作员ID: ")
	operatorID := readInput(reader)

	req := common.CreateStocktakingReq{
		OperatorID: operatorID,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/stocktaking", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Code != 0 {
		fmt.Printf("错误: %s\n", result.Message)
		return
	}

	dataMap, _ := result.Data.(map[string]interface{})
	fmt.Printf("盘点任务创建成功，ID: %v, 状态: %v\n", dataMap["stocktaking_id"], dataMap["status"])
}

func listStocktakings() {
	resp, err := http.Get(serverURL + "/api/stocktaking")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Code != 0 {
		fmt.Printf("错误: %s\n", result.Message)
		return
	}

	dataMap, _ := result.Data.(map[string]interface{})
	stocktakings, _ := dataMap["stocktakings"].([]interface{})

	fmt.Println("\n盘点任务列表:")
	for _, st := range stocktakings {
		s := st.(map[string]interface{})
		fmt.Printf("ID: %v, 状态: %v, 操作员: %v, 创建时间: %v\n",
			s["id"], s["status"], s["operator_id"], s["created_at"])
	}
}

func getStocktakingDetail(reader *bufio.Reader) {
	fmt.Print("盘点任务ID: ")
	id := readInput(reader)

	resp, err := http.Get(serverURL + "/api/stocktaking/" + id)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Code != 0 {
		fmt.Printf("错误: %s\n", result.Message)
		return
	}

	st, _ := result.Data.(map[string]interface{})
	fmt.Printf("\n=== 盘点任务详情 ===\n")
	fmt.Printf("ID: %v\n", st["id"])
	fmt.Printf("状态: %v\n", st["status"])
	fmt.Printf("操作员: %v\n", st["operator_id"])
	fmt.Printf("创建时间: %v\n", st["created_at"])

	items, _ := st["items"].([]interface{})
	fmt.Println("\n盘点商品:")
	for _, item := range items {
		i := item.(map[string]interface{})
		fmt.Printf("  商品: %v (条码: %v), 账面: %v, 实盘: %v, 差异: %v, 差异金额: %v分\n",
			i["product_name"], i["barcode"], i["snapshot_qty"], i["actual_qty"], i["diff_qty"], i["diff_amount"])
	}

	pendingOps, _ := st["pending_ops"].([]interface{})
	if len(pendingOps) > 0 {
		fmt.Println("\n盘点期间变动:")
		for _, op := range pendingOps {
			o := op.(map[string]interface{})
			fmt.Printf("  类型: %v, 商品ID: %v, 数量: %v, 时间: %v\n",
				o["operation_type"], o["product_id"], o["quantity"], o["operation_time"])
		}
	}
}

func submitStocktaking(reader *bufio.Reader) {
	fmt.Print("盘点任务ID: ")
	id := readInput(reader)

	resp, err := http.Get(serverURL + "/api/stocktaking/" + id)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	var result common.Response
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	if result.Code != 0 {
		fmt.Printf("错误: %s\n", result.Message)
		return
	}

	st, _ := result.Data.(map[string]interface{})
	items, _ := st["items"].([]interface{})

	var submitItems []common.SubmitStocktakingItem
	for _, item := range items {
		i := item.(map[string]interface{})
		fmt.Printf("商品: %v (条码: %v), 账面库存: %v\n",
			i["product_name"], i["barcode"], i["snapshot_qty"])
		fmt.Print("实盘数量: ")
		actualQty := readInt(reader)

		submitItems = append(submitItems, common.SubmitStocktakingItem{
			ProductID: i["product_id"].(string),
			ActualQty: actualQty,
		})
	}

	req := common.SubmitStocktakingReq{
		StocktakingID: id,
		Items:         submitItems,
	}

	body, _ := json.Marshal(req)
	resp2, err := http.Post(serverURL+"/api/stocktaking/submit", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp2.Body.Close()

	var result2 common.Response
	json.NewDecoder(resp2.Body).Decode(&result2)

	if result2.Code != 0 {
		fmt.Printf("错误: %s\n", result2.Message)
		return
	}

	fmt.Println("盘点结果提交成功")
}

func completeStocktaking(reader *bufio.Reader) {
	fmt.Print("盘点任务ID: ")
	id := readInput(reader)

	req := common.CompleteStocktakingReq{
		StocktakingID: id,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/stocktaking/complete", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Code != 0 {
		fmt.Printf("错误: %s\n", result.Message)
		return
	}

	dataMap, _ := result.Data.(map[string]interface{})
	fmt.Printf("盘点完成，ID: %v, 状态: %v\n", dataMap["stocktaking_id"], dataMap["status"])
}

func stockIn(reader *bufio.Reader) {
	fmt.Print("商品ID: ")
	productID := readInput(reader)
	fmt.Print("入库数量: ")
	qty := readInt(reader)
	fmt.Print("操作员ID: ")
	operatorID := readInput(reader)

	req := common.StockInReq{
		ProductID:  productID,
		Quantity:   qty,
		OperatorID: operatorID,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/stock-in", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result common.Response
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Code != 0 {
		fmt.Printf("错误: %s\n", result.Message)
		return
	}

	dataMap, _ := result.Data.(map[string]interface{})
	fmt.Printf("入库成功，操作ID: %v\n", dataMap["id"])
}
