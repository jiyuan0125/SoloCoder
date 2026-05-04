package client

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/example/food-menu/common"
)

type AdminCLI struct {
	client *APIClient
	reader *bufio.Reader
}

func NewAdminCLI(client *APIClient) *AdminCLI {
	return &AdminCLI{
		client: client,
		reader: bufio.NewReader(os.Stdin),
	}
}

func (a *AdminCLI) Run() {
	fmt.Println("========================================")
	fmt.Println("       餐厅菜品管理系统 - 管理员模式")
	fmt.Println("========================================")
	fmt.Println()

	for {
		a.showMenu()
		choice := a.readChoice()

		switch choice {
		case "1":
			a.addDish()
		case "2":
			a.editDish()
		case "3":
			a.listDishes()
		case "4":
			a.setDishStatus(true)
		case "5":
			a.setDishStatus(false)
		case "6":
			a.setTodayRecommend()
		case "7":
			a.showCategorySummary()
		case "8":
			fmt.Println("退出管理员模式")
			return
		default:
			fmt.Println("无效的选择，请重新输入")
		}

		fmt.Println()
		fmt.Print("按回车键继续...")
		a.reader.ReadLine()
	}
}

func (a *AdminCLI) showMenu() {
	fmt.Println("请选择操作：")
	fmt.Println("1. 添加菜品")
	fmt.Println("2. 编辑菜品")
	fmt.Println("3. 查看所有菜品")
	fmt.Println("4. 上架菜品")
	fmt.Println("5. 下架菜品")
	fmt.Println("6. 设置今日推荐")
	fmt.Println("7. 查看分类统计")
	fmt.Println("8. 退出")
	fmt.Print("请输入选择 (1-8): ")
}

func (a *AdminCLI) readChoice() string {
	input, _ := a.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (a *AdminCLI) readString(prompt string) string {
	fmt.Print(prompt)
	input, _ := a.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (a *AdminCLI) readFloat(prompt string) (float64, error) {
	input := a.readString(prompt)
	return strconv.ParseFloat(input, 64)
}

func (a *AdminCLI) addDish() {
	fmt.Println("\n--- 添加菜品 ---")

	fmt.Println("可用分类：")
	for i, cat := range common.ValidCategories {
		fmt.Printf("%d. %s\n", i+1, cat)
	}

	catInput := a.readString("请选择分类编号: ")
	catIndex, err := strconv.Atoi(catInput)
	if err != nil || catIndex < 1 || catIndex > len(common.ValidCategories) {
		fmt.Println("无效的分类编号")
		return
	}
	category := common.ValidCategories[catIndex-1]

	name := a.readString("菜品名称: ")
	if name == "" {
		fmt.Println("菜品名称不能为空")
		return
	}

	price, err := a.readFloat("菜品价格: ")
	if err != nil || price <= 0 {
		fmt.Println("价格必须大于0")
		return
	}

	description := a.readString("菜品描述 (可选): ")
	imageURL := a.readString("图片URL (可选): ")

	req := &common.CreateDishRequest{
		Name:        name,
		Price:       price,
		Description: description,
		ImageURL:    imageURL,
		Category:    category,
	}

	dish, err := a.client.CreateDish(req)
	if err != nil {
		fmt.Printf("添加失败: %v\n", err)
		return
	}

	fmt.Printf("添加成功！菜品ID: %s\n", dish.ID)
}

func (a *AdminCLI) editDish() {
	fmt.Println("\n--- 编辑菜品 ---")

	id := a.readString("请输入菜品ID: ")
	if id == "" {
		fmt.Println("菜品ID不能为空")
		return
	}

	dish, err := a.client.GetDish(id)
	if err != nil {
		fmt.Printf("获取菜品失败: %v\n", err)
		return
	}

	fmt.Printf("\n当前菜品信息:\n")
	fmt.Printf("名称: %s\n", dish.Name)
	fmt.Printf("价格: %.2f\n", dish.Price)
	fmt.Printf("描述: %s\n", dish.Description)
	fmt.Printf("图片URL: %s\n", dish.ImageURL)
	fmt.Printf("分类: %s\n", dish.Category)
	fmt.Println()

	name := a.readString("新名称 (直接回车保持不变): ")
	priceInput := a.readString("新价格 (直接回车保持不变): ")
	description := a.readString("新描述 (直接回车保持不变): ")
	imageURL := a.readString("新图片URL (直接回车保持不变): ")

	fmt.Println("\n可选新分类 (直接回车保持不变):")
	for i, cat := range common.ValidCategories {
		fmt.Printf("%d. %s\n", i+1, cat)
	}
	catInput := a.readString("请选择新分类编号 (直接回车保持不变): ")

	req := &common.UpdateDishRequest{}
	if name != "" {
		req.Name = name
	}
	if priceInput != "" {
		price, err := strconv.ParseFloat(priceInput, 64)
		if err != nil || price <= 0 {
			fmt.Println("价格必须大于0")
			return
		}
		req.Price = price
	}
	if description != "" {
		req.Description = description
	}
	if imageURL != "" {
		req.ImageURL = imageURL
	}
	if catInput != "" {
		catIndex, err := strconv.Atoi(catInput)
		if err != nil || catIndex < 1 || catIndex > len(common.ValidCategories) {
			fmt.Println("无效的分类编号")
			return
		}
		req.Category = common.ValidCategories[catIndex-1]
	}

	if req.Name == "" && req.Price == 0 && req.Description == "" && req.ImageURL == "" && req.Category == "" {
		fmt.Println("没有修改任何内容")
		return
	}

	updatedDish, err := a.client.UpdateDish(id, req)
	if err != nil {
		fmt.Printf("更新失败: %v\n", err)
		return
	}

	fmt.Printf("更新成功！\n")
	fmt.Printf("名称: %s\n", updatedDish.Name)
	fmt.Printf("价格: %.2f\n", updatedDish.Price)
	fmt.Printf("分类: %s\n", updatedDish.Category)
}

func (a *AdminCLI) listDishes() {
	fmt.Println("\n--- 所有菜品列表 ---")

	listResp, err := a.client.ListDishes("", false)
	if err != nil {
		fmt.Printf("获取列表失败: %v\n", err)
		return
	}

	if len(listResp.Dishes) == 0 {
		fmt.Println("暂无菜品")
		return
	}

	fmt.Printf("共 %d 道菜品:\n", listResp.Total)
	fmt.Println("----------------------------------------")
	for _, dish := range listResp.Dishes {
		status := "在售"
		if dish.Status == common.StatusOffSale {
			status = "已下架"
		}
		recommend := ""
		if dish.IsRecommend {
			recommend = " [今日推荐]"
		}
		fmt.Printf("ID: %s\n", dish.ID)
		fmt.Printf("名称: %s%s\n", dish.Name, recommend)
		fmt.Printf("价格: ¥%.2f\n", dish.Price)
		fmt.Printf("分类: %s | 状态: %s\n", dish.Category, status)
		fmt.Printf("描述: %s\n", dish.Description)
		fmt.Println("----------------------------------------")
	}
}

func (a *AdminCLI) setDishStatus(onSale bool) {
	action := "上架"
	if !onSale {
		action = "下架"
	}

	fmt.Printf("\n--- %s菜品 ---\n", action)

	id := a.readString("请输入菜品ID: ")
	if id == "" {
		fmt.Println("菜品ID不能为空")
		return
	}

	var err error
	if onSale {
		err = a.client.SetDishOnSale(id)
	} else {
		err = a.client.SetDishOffSale(id)
	}

	if err != nil {
		fmt.Printf("%s失败: %v\n", action, err)
		return
	}

	fmt.Printf("%s成功！\n", action)
}

func (a *AdminCLI) setTodayRecommend() {
	fmt.Println("\n--- 设置今日推荐 ---")

	listResp, err := a.client.ListDishes("", true)
	if err != nil {
		fmt.Printf("获取在售菜品失败: %v\n", err)
		return
	}

	if len(listResp.Dishes) == 0 {
		fmt.Println("暂无在售菜品，无法设置推荐")
		return
	}

	fmt.Println("在售菜品列表：")
	for i, dish := range listResp.Dishes {
		fmt.Printf("%d. ID: %s | 名称: %s | 价格: ¥%.2f\n",
			i+1, dish.ID, dish.Name, dish.Price)
	}

	input := a.readString("请输入要推荐的菜品编号（多个用逗号分隔，例如: 1,3,5）: ")
	if input == "" {
		fmt.Println("未选择任何菜品，将清除今日推荐")
		_, err := a.client.SetTodayRecommend([]string{})
		if err != nil {
			fmt.Printf("清除推荐失败: %v\n", err)
			return
		}
		fmt.Println("已清除今日推荐")
		return
	}

	parts := strings.Split(input, ",")
	var dishIDs []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		index, err := strconv.Atoi(part)
		if err != nil || index < 1 || index > len(listResp.Dishes) {
			fmt.Printf("无效的编号: %s\n", part)
			return
		}
		dishIDs = append(dishIDs, listResp.Dishes[index-1].ID)
	}

	resp, err := a.client.SetTodayRecommend(dishIDs)
	if err != nil {
		fmt.Printf("设置推荐失败: %v\n", err)
		return
	}

	fmt.Println("\n今日推荐设置成功！")
	fmt.Printf("推荐日期: %s\n", resp.RecommendInfo.Date)
	fmt.Println("推荐菜品:")
	for _, dish := range resp.Dishes {
		fmt.Printf("  - %s (¥%.2f)\n", dish.Name, dish.Price)
	}
}

func (a *AdminCLI) showCategorySummary() {
	fmt.Println("\n--- 分类统计 ---")

	summaryResp, err := a.client.GetCategorySummary()
	if err != nil {
		fmt.Printf("获取统计失败: %v\n", err)
		return
	}

	fmt.Println("----------------------------------------")
	fmt.Printf("%-6s | %-8s | %-8s\n", "分类", "总数", "在售")
	fmt.Println("----------------------------------------")
	for _, s := range summaryResp.Summaries {
		fmt.Printf("%-6s | %-8d | %-8d\n", s.Category, s.TotalCount, s.OnSaleCount)
	}
	fmt.Println("----------------------------------------")
}
