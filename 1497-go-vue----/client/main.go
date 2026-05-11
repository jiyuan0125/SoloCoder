package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"marketplace/common"
	"os"
	"strconv"
	"strings"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "服务端地址")
	flag.Parse()

	client := NewAPIClient(*serverURL, "")
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("二手闲置交易平台客户端")
	fmt.Println("====================")
	fmt.Println("可用命令:")
	fmt.Println("  register <username> [admin]  - 注册用户")
	fmt.Println("  login <user_id>          - 设置当前用户ID")
	fmt.Println("  publish                 - 发布商品")
	fmt.Println("  list                    - 浏览商品")
	fmt.Println("  search <keyword>        - 搜索商品")
	fmt.Println("  get <item_id>           - 查看商品详情")
	fmt.Println("  offer <item_id> <price> - 出价（分）")
	fmt.Println("  seller-respond <neg_id> <action> [price] - 卖家响应 (accept/reject/counter)")
	fmt.Println("  buyer-respond <neg_id> <action> [price]  - 买家响应 (accept/reject/counter)")
	fmt.Println("  approve <item_id> <approve> - 审核商品 (true/false)")
	fmt.Println("  remove <item_id>       - 下架商品")
	fmt.Println("  pay <order_id>         - 付款")
	fmt.Println("  ship <order_id> <logistics_no> - 发货")
	fmt.Println("  confirm <order_id>    - 确认收货")
	fmt.Println("  get-order <order_id>    - 查看订单")
	fmt.Println("  help                    - 显示帮助")
	fmt.Println("  exit                    - 退出")
	fmt.Println("====================")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]

		switch cmd {
		case "exit":
			fmt.Println("再见！")
			return

		case "help":
			fmt.Println("可用命令:")
			fmt.Println("  register <username> [admin]  - 注册用户")
			fmt.Println("  login <user_id>          - 设置当前用户ID")
			fmt.Println("  publish                 - 发布商品")
			fmt.Println("  list                    - 浏览商品")
			fmt.Println("  search <keyword>        - 搜索商品")
			fmt.Println("  get <item_id>           - 查看商品详情")
			fmt.Println("  offer <item_id> <price> - 出价（分）")
			fmt.Println("  seller-respond <neg_id> <action> [price] - 卖家响应")
			fmt.Println("  buyer-respond <neg_id> <action> [price]  - 买家响应")
			fmt.Println("  approve <item_id> <approve> - 审核商品")
			fmt.Println("  remove <item_id>       - 下架商品")
			fmt.Println("  pay <order_id>         - 付款")
			fmt.Println("  ship <order_id> <logistics_no> - 发货")
			fmt.Println("  confirm <order_id>    - 确认收货")
			fmt.Println("  get-order <order_id>    - 查看订单")
			fmt.Println("  help                    - 显示帮助")
			fmt.Println("  exit                    - 退出")

		case "register":
			if len(parts) < 2 {
				fmt.Println("用法: register <username> [admin]")
				continue
			}
			role := common.UserRoleUser
			if len(parts) >= 3 && parts[2] == "admin" {
				role = common.UserRoleAdmin
			}
			user, err := client.Register(parts[1], role)
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			fmt.Printf("注册成功，用户ID: %s\n", user.ID)
			client.SetUserID(user.ID)

		case "login":
			if len(parts) < 2 {
				fmt.Println("用法: login <user_id>")
				continue
			}
			client.SetUserID(parts[1])
			fmt.Printf("已设置当前用户ID: %s\n", parts[1])

		case "publish":
			fmt.Println("发布商品")
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("标题: ")
			title, _ := reader.ReadString('\n')
			title = strings.TrimSpace(title)
			fmt.Print("描述: ")
			desc, _ := reader.ReadString('\n')
			desc = strings.TrimSpace(desc)
			fmt.Print("分类 (数码电子/家居用品/图书文具/服饰鞋包/运动户外/其他): ")
			cat, _ := reader.ReadString('\n')
			cat = strings.TrimSpace(cat)
			fmt.Print("新旧程度 (全新未拆/几乎全新/轻微使用痕迹/明显使用痕迹): ")
			cond, _ := reader.ReadString('\n')
			cond = strings.TrimSpace(cond)
			fmt.Print("原价(分): ")
			origPriceStr, _ := reader.ReadString('\n')
			origPriceStr = strings.TrimSpace(origPriceStr)
			origPrice, _ := strconv.ParseInt(origPriceStr, 10, 64)
			fmt.Print("售价(分): ")
			priceStr, _ := reader.ReadString('\n')
			priceStr = strings.TrimSpace(priceStr)
			price, _ := strconv.ParseInt(priceStr, 10, 64)

			req := common.CreateItemRequest{
				Title:         title,
				Description:   desc,
				Category:    common.Category(cat),
				Condition:   common.Condition(cond),
				OriginalPrice: origPrice,
				Price:         price,
			}

			item, err := client.CreateItem(req)
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			fmt.Printf("商品已发布，ID: %s\n", item.ID)

		case "list":
			items, err := client.SearchItems(common.SearchItemsRequest{})
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			if len(items) == 0 {
				fmt.Println("暂无商品")
			}
			for _, item := range items {
				fmt.Printf("ID: %s\n", item.ID)
				fmt.Printf("  标题: %s\n", item.Title)
				fmt.Printf("  分类: %s, 新旧: %s\n", item.Category, item.Condition)
				fmt.Printf("  原价: %d分, 售价: %d分\n", item.OriginalPrice, item.Price)
				fmt.Printf("  状态: %s\n", item.Status)
				fmt.Println("---")
			}

		case "search":
			if len(parts) < 2 {
				fmt.Println("用法: search <keyword>")
				continue
			}
			items, err := client.SearchItems(common.SearchItemsRequest{Keyword: parts[1]})
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			if len(items) == 0 {
				fmt.Println("未找到商品")
			}
			for _, item := range items {
				fmt.Printf("ID: %s, 标题: %s, 价格: %d分\n", item.ID, item.Title, item.Price)
			}

		case "get":
			if len(parts) < 2 {
				fmt.Println("用法: get <item_id>")
				continue
			}
			item, err := client.GetItem(parts[1])
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			data, _ := json.MarshalIndent(item, "", "  ")
			fmt.Println(string(data))

		case "offer":
			if len(parts) < 3 {
				fmt.Println("用法: offer <item_id> <price>")
				continue
			}
			price, _ := strconv.ParseInt(parts[2], 10, 64)
			neg, err := client.MakeOffer(parts[1], price)
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			fmt.Printf("出价成功，议价ID: %s\n", neg.ID)

		case "seller-respond":
			if len(parts) < 3 {
				fmt.Println("用法: seller-respond <neg_id> <action> [price]")
				continue
			}
			action := parts[2]
			var counterPrice int64 = 0
			if len(parts) >= 4 {
				counterPrice, _ = strconv.ParseInt(parts[3], 10, 64)
			}
			neg, err := client.SellerRespond(parts[1], action, counterPrice)
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			fmt.Printf("响应成功，状态: %s\n", neg.Status)

		case "buyer-respond":
			if len(parts) < 3 {
				fmt.Println("用法: buyer-respond <neg_id> <action> [price]")
				continue
			}
			action := parts[2]
			var counterPrice int64 = 0
			if len(parts) >= 4 {
				counterPrice, _ = strconv.ParseInt(parts[3], 10, 64)
			}
			neg, err := client.BuyerRespond(parts[1], action, counterPrice)
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			fmt.Printf("响应成功，状态: %s\n", neg.Status)

		case "approve":
			if len(parts) < 3 {
				fmt.Println("用法: approve <item_id> <approve>")
				continue
			}
			approved := parts[2] == "true"
			item, err := client.ApproveItem(parts[1], approved, "")
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			fmt.Printf("审核完成，状态: %s\n", item.Status)

		case "remove":
			if len(parts) < 2 {
				fmt.Println("用法: remove <item_id>")
				continue
			}
			err := client.RemoveItem(parts[1])
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			fmt.Println("下架成功")

		case "pay":
			if len(parts) < 2 {
				fmt.Println("用法: pay <order_id>")
				continue
			}
			order, err := client.PayOrder(parts[1])
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			fmt.Printf("付款成功，状态: %s\n", order.Status)

		case "ship":
			if len(parts) < 3 {
				fmt.Println("用法: ship <order_id> <logistics_no>")
				continue
			}
			order, err := client.ShipOrder(parts[1], parts[2])
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			fmt.Printf("发货成功，状态: %s\n", order.Status)

		case "confirm":
			if len(parts) < 2 {
				fmt.Println("用法: confirm <order_id>")
				continue
			}
			order, err := client.ConfirmReceiveOrder(parts[1])
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			fmt.Printf("确认收货成功，状态: %s\n", order.Status)

		case "get-order":
			if len(parts) < 2 {
				fmt.Println("用法: get-order <order_id>")
				continue
			}
			order, err := client.GetOrder(parts[1])
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			data, _ := json.MarshalIndent(order, "", "  ")
			fmt.Println(string(data))

		default:
			fmt.Println("未知命令，输入 help 查看帮助")
		}
	}
}
