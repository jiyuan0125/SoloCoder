package main

import (
	"auction-house/internal/client"
	"auction-house/internal/common"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var apiClient *client.APIClient
var scanner *bufio.Scanner

func main() {
	apiClient = client.NewAPIClient("")
	scanner = bufio.NewScanner(os.Stdin)

	fmt.Println("==============================")
	fmt.Println("  竞拍平台客户端")
	fmt.Println("==============================")

	for {
		printMenu()
		choice := getInput("请选择操作: ")

		switch choice {
		case "1":
			createAuction()
		case "2":
			getSellerAuctions()
		case "3":
			placeBid()
		case "4":
			getBidHistory()
		case "5":
			getAuction()
		case "0":
			fmt.Println("退出客户端")
			return
		default:
			fmt.Println("无效选择，请重新输入")
		}

		fmt.Println()
	}
}

func printMenu() {
	fmt.Println("\n菜单:")
	fmt.Println("  1. 发布拍品 (卖家)")
	fmt.Println("  2. 查看我的拍品 (卖家)")
	fmt.Println("  3. 出价 (买家)")
	fmt.Println("  4. 查询出价历史")
	fmt.Println("  5. 查看拍品详情")
	fmt.Println("  0. 退出")
}

func getInput(prompt string) string {
	fmt.Print(prompt)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func getFloatInput(prompt string) (float64, error) {
	input := getInput(prompt)
	return strconv.ParseFloat(input, 64)
}

func createAuction() {
	fmt.Println("\n--- 发布拍品 ---")

	sellerID := getInput("请输入您的卖家ID: ")
	if sellerID == "" {
		fmt.Println("卖家ID不能为空")
		return
	}

	name := getInput("拍品名称: ")
	if name == "" {
		fmt.Println("拍品名称不能为空")
		return
	}

	description := getInput("拍品描述 (最多500字): ")
	if len(description) > 500 {
		fmt.Println("描述不能超过500字")
		return
	}

	startPrice, err := getFloatInput("起拍价: ")
	if err != nil || startPrice <= 0 {
		fmt.Println("起拍价必须大于零")
		return
	}

	bidIncrement, err := getFloatInput("每次加价最小幅度: ")
	if err != nil || bidIncrement <= 0 {
		fmt.Println("加价幅度必须大于零")
		return
	}

	deadlineStr := getInput("竞拍截止时间 (格式: 2006-01-02 15:04:05): ")
	deadline, err := time.ParseInLocation("2006-01-02 15:04:05", deadlineStr, time.Local)
	if err != nil {
		fmt.Println("时间格式错误，请使用: 2006-01-02 15:04:05")
		return
	}

	req := &common.CreateAuctionRequest{
		Name:         name,
		Description:  description,
		StartPrice:   startPrice,
		BidIncrement: bidIncrement,
		Deadline:     deadline.Format(time.RFC3339),
		SellerID:     sellerID,
	}

	resp, err := apiClient.CreateAuction(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("操作失败: %s (错误码: %d)\n", resp.Message, resp.Code)
		return
	}

	fmt.Println("拍品发布成功!")
	client.PrintAuction(resp.Data)
}

func getSellerAuctions() {
	fmt.Println("\n--- 查看我的拍品 ---")

	sellerID := getInput("请输入您的卖家ID: ")
	if sellerID == "" {
		fmt.Println("卖家ID不能为空")
		return
	}

	resp, err := apiClient.GetSellerAuctions(sellerID)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("操作失败: %s (错误码: %d)\n", resp.Message, resp.Code)
		return
	}

	if len(resp.Data) == 0 {
		fmt.Println("您还没有发布任何拍品")
		return
	}

	fmt.Printf("找到 %d 个拍品:\n", len(resp.Data))
	for i, auction := range resp.Data {
		fmt.Printf("\n拍品 %d:\n", i+1)
		client.PrintAuction(auction)
	}
}

func placeBid() {
	fmt.Println("\n--- 出价 ---")

	bidderID := getInput("请输入您的买家ID: ")
	if bidderID == "" {
		fmt.Println("买家ID不能为空")
		return
	}

	auctionID := getInput("拍品ID: ")
	if auctionID == "" {
		fmt.Println("拍品ID不能为空")
		return
	}

	price, err := getFloatInput("出价金额: ")
	if err != nil || price <= 0 {
		fmt.Println("出价金额必须大于零")
		return
	}

	req := &common.BidRequest{
		AuctionID: auctionID,
		BidderID:  bidderID,
		Price:     price,
	}

	resp, err := apiClient.PlaceBid(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("出价失败: %s (错误码: %d)\n", resp.Message, resp.Code)
		return
	}

	fmt.Println("出价成功!")
	fmt.Printf("出价ID: %s\n", resp.Data.ID)
	fmt.Printf("拍品ID: %s\n", resp.Data.AuctionID)
	fmt.Printf("出价人: %s\n", resp.Data.BidderID)
	fmt.Printf("出价金额: %.2f\n", resp.Data.Price)
	fmt.Printf("出价时间: %s\n", resp.Data.BidTime.Format("2006-01-02 15:04:05"))
}

func getBidHistory() {
	fmt.Println("\n--- 查询出价历史 ---")

	auctionID := getInput("拍品ID: ")
	if auctionID == "" {
		fmt.Println("拍品ID不能为空")
		return
	}

	resp, err := apiClient.GetBidHistory(auctionID)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("操作失败: %s (错误码: %d)\n", resp.Message, resp.Code)
		return
	}

	if len(resp.Data) == 0 {
		fmt.Println("该拍品暂无出价记录")
		return
	}

	fmt.Printf("找到 %d 条出价记录 (按价格从高到低排序):\n", len(resp.Data))
	for i, bid := range resp.Data {
		client.PrintBid(bid, i+1)
	}
}

func getAuction() {
	fmt.Println("\n--- 查看拍品详情 ---")

	auctionID := getInput("拍品ID: ")
	if auctionID == "" {
		fmt.Println("拍品ID不能为空")
		return
	}

	resp, err := apiClient.GetAuction(auctionID)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("操作失败: %s (错误码: %d)\n", resp.Message, resp.Code)
		return
	}

	client.PrintAuction(resp.Data)
}
