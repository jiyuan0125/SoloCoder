package client

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/example/food-menu/common"
)

type CustomerCLI struct {
	client *APIClient
	reader *bufio.Reader
}

func NewCustomerCLI(client *APIClient) *CustomerCLI {
	return &CustomerCLI{
		client: client,
		reader: bufio.NewReader(os.Stdin),
	}
}

func (c *CustomerCLI) Run() {
	fmt.Println("========================================")
	fmt.Println("       餐厅菜品管理系统 - 顾客模式")
	fmt.Println("========================================")
	fmt.Println()

	for {
		c.showMenu()
		choice := c.readChoice()

		switch choice {
		case "1":
			c.showTodayRecommend()
		case "2":
			c.browseByCategory()
		case "3":
			c.searchDishes()
		case "4":
			c.viewDishDetail()
		case "5":
			fmt.Println("退出顾客模式")
			return
		default:
			fmt.Println("无效的选择，请重新输入")
		}

		fmt.Println()
		fmt.Print("按回车键继续...")
		c.reader.ReadLine()
	}
}

func (c *CustomerCLI) showMenu() {
	fmt.Println("请选择操作：")
	fmt.Println("1. 查看今日推荐")
	fmt.Println("2. 按分类浏览")
	fmt.Println("3. 搜索菜品")
	fmt.Println("4. 查看菜品详情")
	fmt.Println("5. 退出")
	fmt.Print("请输入选择 (1-5): ")
}

func (c *CustomerCLI) readChoice() string {
	input, _ := c.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (c *CustomerCLI) readString(prompt string) string {
	fmt.Print(prompt)
	input, _ := c.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (c *CustomerCLI) showTodayRecommend() {
	fmt.Println("\n--- 今日推荐 ---")

	resp, err := c.client.GetTodayRecommend()
	if err != nil {
		fmt.Printf("获取推荐失败: %v\n", err)
		return
	}

	if len(resp.Dishes) == 0 {
		fmt.Println("今日暂无推荐菜品")
		return
	}

	fmt.Printf("推荐日期: %s\n", resp.RecommendInfo.Date)
	fmt.Println("----------------------------------------")
	for i, dish := range resp.Dishes {
		fmt.Printf("[%d] %s\n", i+1, dish.Name)
		fmt.Printf("    价格: ¥%.2f\n", dish.Price)
		fmt.Printf("    分类: %s\n", dish.Category)
		if dish.Description != "" {
			fmt.Printf("    描述: %s\n", dish.Description)
		}
		fmt.Println("----------------------------------------")
	}
}

func (c *CustomerCLI) browseByCategory() {
	fmt.Println("\n--- 按分类浏览 ---")

	fmt.Println("可用分类：")
	for i, cat := range common.ValidCategories {
		fmt.Printf("%d. %s\n", i+1, cat)
	}
	fmt.Printf("%d. 全部\n", len(common.ValidCategories)+1)

	input := c.readString("请选择分类编号: ")
	index, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("无效的输入")
		return
	}

	var category string
	if index == len(common.ValidCategories)+1 {
		category = ""
	} else if index >= 1 && index <= len(common.ValidCategories) {
		category = common.ValidCategories[index-1]
	} else {
		fmt.Println("无效的分类编号")
		return
	}

	listResp, err := c.client.SearchDishes(category, "")
	if err != nil {
		fmt.Printf("获取菜品列表失败: %v\n", err)
		return
	}

	if len(listResp.Dishes) == 0 {
		fmt.Println("该分类暂无在售菜品")
		return
	}

	categoryName := category
	if categoryName == "" {
		categoryName = "全部"
	}
	fmt.Printf("\n%s分类 - 共 %d 道在售菜品:\n", categoryName, listResp.Total)
	fmt.Println("----------------------------------------")
	for i, dish := range listResp.Dishes {
		recommend := ""
		if dish.IsRecommend {
			recommend = " [★今日推荐]"
		}
		fmt.Printf("[%d] %s%s\n", i+1, dish.Name, recommend)
		fmt.Printf("    价格: ¥%.2f | 分类: %s\n", dish.Price, dish.Category)
		if dish.Description != "" {
			fmt.Printf("    描述: %s\n", dish.Description)
		}
		fmt.Println("----------------------------------------")
	}
}

func (c *CustomerCLI) searchDishes() {
	fmt.Println("\n--- 搜索菜品 ---")

	fmt.Println("可选分类筛选（直接回车不筛选）：")
	for i, cat := range common.ValidCategories {
		fmt.Printf("%d. %s\n", i+1, cat)
	}

	catInput := c.readString("请选择分类编号 (可选): ")
	var category string
	if catInput != "" {
		catIndex, err := strconv.Atoi(catInput)
		if err != nil || catIndex < 1 || catIndex > len(common.ValidCategories) {
			fmt.Println("无效的分类编号，将不进行分类筛选")
		} else {
			category = common.ValidCategories[catIndex-1]
		}
	}

	keyword := c.readString("请输入搜索关键词 (空则显示全部): ")

	listResp, err := c.client.SearchDishes(category, keyword)
	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return
	}

	if len(listResp.Dishes) == 0 {
		if keyword == "" {
			fmt.Println("暂无在售菜品")
		} else {
			fmt.Printf("未找到包含 \"%s\" 的菜品\n", keyword)
		}
		return
	}

	filterInfo := ""
	if category != "" {
		filterInfo = fmt.Sprintf(" [%s分类]", category)
	}
	if keyword != "" {
		filterInfo += fmt.Sprintf(" (关键词: %s)", keyword)
	}
	fmt.Printf("\n搜索结果%s - 共 %d 道:\n", filterInfo, listResp.Total)
	fmt.Println("----------------------------------------")
	for i, dish := range listResp.Dishes {
		recommend := ""
		if dish.IsRecommend {
			recommend = " [★今日推荐]"
		}
		fmt.Printf("[%d] %s%s\n", i+1, dish.Name, recommend)
		fmt.Printf("    价格: ¥%.2f | 分类: %s\n", dish.Price, dish.Category)
		if dish.Description != "" {
			fmt.Printf("    描述: %s\n", dish.Description)
		}
		fmt.Println("----------------------------------------")
	}
}

func (c *CustomerCLI) viewDishDetail() {
	fmt.Println("\n--- 查看菜品详情 ---")

	id := c.readString("请输入菜品ID: ")
	if id == "" {
		fmt.Println("菜品ID不能为空")
		return
	}

	dish, err := c.client.GetDish(id)
	if err != nil {
		fmt.Printf("获取菜品失败: %v\n", err)
		return
	}

	fmt.Println("----------------------------------------")
	fmt.Printf("菜品名称: %s\n", dish.Name)
	fmt.Printf("价格: ¥%.2f\n", dish.Price)
	fmt.Printf("分类: %s\n", dish.Category)

	if dish.Status == common.StatusOnSale {
		fmt.Printf("状态: 在售\n")
	} else {
		fmt.Printf("状态: 该菜品暂时不可用\n")
	}

	if dish.IsRecommend {
		fmt.Printf("推荐: ★今日推荐\n")
	}

	if dish.Description != "" {
		fmt.Printf("描述: %s\n", dish.Description)
	}

	if dish.ImageURL != "" {
		fmt.Printf("图片: %s\n", dish.ImageURL)
	}
	fmt.Println("----------------------------------------")
}
