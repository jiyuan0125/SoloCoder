package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"feedback-system/internal/client/api"
	"feedback-system/pkg/common"
)

func main() {
	serverAddr := flag.String("server", "http://localhost:8080", "Server address")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printHelp()
		os.Exit(1)
	}

	client := api.NewClient(*serverAddr)
	cmd := strings.ToLower(args[0])

	switch cmd {
	case "submit":
		handleSubmit(client)
	case "list":
		handleList(client, args[1:])
	case "get":
		handleGet(client, args[1:])
	case "note":
		handleAddNote(client, args[1:])
	case "status":
		handleUpdateStatus(client, args[1:])
	case "stats":
		handleStats(client)
	case "help":
		printHelp()
	default:
		fmt.Printf("未知命令: %s\n\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("用户反馈收集系统客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  feedback-client [命令] [选项]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  submit          提交反馈（用户功能）")
	fmt.Println("  list            查看反馈列表（管理员功能）")
	fmt.Println("  get <id>        查看单条反馈详情（管理员功能）")
	fmt.Println("  note <id>       添加内部备注（管理员功能）")
	fmt.Println("  status <id>     更新反馈状态（管理员功能）")
	fmt.Println("  stats           查看统计概览（管理员功能）")
	fmt.Println("  help            显示帮助信息")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -server <addr>  指定服务端地址（默认: http://localhost:8080）")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  feedback-client submit")
	fmt.Println("  feedback-client list --type bug --status pending")
	fmt.Println("  feedback-client get abc123")
	fmt.Println("  feedback-client note abc123")
	fmt.Println("  feedback-client status abc123")
	fmt.Println("  feedback-client stats")
}

func handleSubmit(client *api.Client) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("请输入用户标识: ")
	userID, _ := reader.ReadString('\n')
	userID = strings.TrimSpace(userID)
	if userID == "" {
		fmt.Println("错误: 用户标识不能为空")
		os.Exit(1)
	}

	fmt.Println("\n请选择反馈类型:")
	fmt.Println("  1. 功能建议 (suggestion)")
	fmt.Println("  2. Bug报告 (bug)")
	fmt.Println("  3. 体验投诉 (complaint)")
	fmt.Print("请输入选项 (1-3): ")
	typeInput, _ := reader.ReadString('\n')
	typeInput = strings.TrimSpace(typeInput)

	var feedbackType string
	switch typeInput {
	case "1", "suggestion":
		feedbackType = "suggestion"
	case "2", "bug":
		feedbackType = "bug"
	case "3", "complaint":
		feedbackType = "complaint"
	default:
		fmt.Println("错误: 无效的反馈类型")
		os.Exit(1)
	}

	fmt.Print("\n请输入评分 (1-5星): ")
	ratingInput, _ := reader.ReadString('\n')
	ratingInput = strings.TrimSpace(ratingInput)
	rating, err := strconv.Atoi(ratingInput)
	if err != nil {
		fmt.Println("错误: 评分必须是1-5之间的数字")
		os.Exit(1)
	}

	fmt.Printf("\n请输入反馈描述（最多%d字）:\n", common.MaxDescriptionLen)
	description, _ := reader.ReadString('\n')
	description = strings.TrimSpace(description)

	req := common.SubmitFeedbackRequest{
		UserID:      userID,
		Type:        feedbackType,
		Rating:      common.Rating(rating),
		Description: description,
	}

	resp, err := client.SubmitFeedback(req)
	if err != nil {
		fmt.Printf("提交失败: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("\n提交成功！反馈ID: %s\n", resp.FeedbackID)
	} else {
		fmt.Printf("\n提交失败: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleList(client *api.Client, args []string) {
	page := 1
	pageSize := common.DefaultPageSize
	var filterType, filterStatus string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--type" && i+1 < len(args):
			filterType = args[i+1]
			i++
		case arg == "--status" && i+1 < len(args):
			filterStatus = args[i+1]
			i++
		case arg == "--page" && i+1 < len(args):
			if p, err := strconv.Atoi(args[i+1]); err == nil {
				page = p
			}
			i++
		case arg == "--page-size" && i+1 < len(args):
			if ps, err := strconv.Atoi(args[i+1]); err == nil {
				pageSize = ps
			}
			i++
		}
	}

	req := common.ListFeedbackRequest{
		Page:     page,
		PageSize: pageSize,
		Type:     filterType,
		Status:   filterStatus,
	}

	resp, err := client.ListFeedbacks(req)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("查询失败: %s\n", resp.Message)
		os.Exit(1)
	}

	fmt.Printf("\n反馈列表（第 %d/%d 页，共 %d 条）:\n", resp.Page, resp.TotalPages, resp.Total)
	fmt.Println(strings.Repeat("-", 80))

	for _, f := range resp.Feedbacks {
		fmt.Printf("ID: %s\n", f.ID)
		fmt.Printf("  用户: %s\n", f.UserID)
		fmt.Printf("  类型: %s (%s)\n", f.Type, f.Type.DisplayName())
		fmt.Printf("  评分: %d星\n", f.Rating)
		fmt.Printf("  状态: %s (%s)\n", f.Status, f.Status.DisplayName())
		desc := f.Description
		if len(desc) > 50 {
			desc = desc[:50] + "..."
		}
		fmt.Printf("  描述: %s\n", desc)
		fmt.Printf("  创建时间: %s\n", f.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println(strings.Repeat("-", 80))
	}
}

func handleGet(client *api.Client, args []string) {
	if len(args) < 1 {
		fmt.Println("用法: feedback-client get <feedback-id>")
		os.Exit(1)
	}

	feedbackID := args[0]
	resp, err := client.GetFeedback(feedbackID)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("查询失败: %s\n", resp.Message)
		os.Exit(1)
	}

	f := resp.Feedback
	fmt.Println("\n反馈详情:")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("ID:          %s\n", f.ID)
	fmt.Printf("用户:        %s\n", f.UserID)
	fmt.Printf("类型:        %s (%s)\n", f.Type, f.Type.DisplayName())
	fmt.Printf("评分:        %d星\n", f.Rating)
	fmt.Printf("状态:        %s (%s)\n", f.Status, f.Status.DisplayName())
	fmt.Println("描述:")
	fmt.Println(f.Description)
	fmt.Println(strings.Repeat("-", 80))

	if f.InternalNote != "" {
		fmt.Println("内部备注:")
		fmt.Println(f.InternalNote)
		fmt.Println(strings.Repeat("-", 80))
	}

	if f.ProcessNote != "" {
		fmt.Println("处理说明:")
		fmt.Println(f.ProcessNote)
		fmt.Println(strings.Repeat("-", 80))
	}

	fmt.Printf("创建时间:    %s\n", f.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("更新时间:    %s\n", f.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 80))
}

func handleAddNote(client *api.Client, args []string) {
	if len(args) < 1 {
		fmt.Println("用法: feedback-client note <feedback-id>")
		os.Exit(1)
	}

	feedbackID := args[0]
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("请输入备注内容（最多%d字，空行结束）:\n", common.MaxNoteLen)
	var lines []string
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimRight(line, "\n\r")
		if line == "" {
			break
		}
		lines = append(lines, line)
	}

	note := strings.Join(lines, "\n")
	if note == "" {
		fmt.Println("错误: 备注不能为空")
		os.Exit(1)
	}

	resp, err := client.AddNote(feedbackID, note)
	if err != nil {
		fmt.Printf("添加备注失败: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Println("\n备注添加成功！")
	} else {
		fmt.Printf("\n添加备注失败: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleUpdateStatus(client *api.Client, args []string) {
	if len(args) < 1 {
		fmt.Println("用法: feedback-client status <feedback-id>")
		os.Exit(1)
	}

	feedbackID := args[0]
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n请选择新状态:")
	fmt.Println("  1. 待处理 (pending)")
	fmt.Println("  2. 处理中 (in_process)")
	fmt.Println("  3. 已关闭 (closed)")
	fmt.Print("请输入选项 (1-3): ")

	statusInput, _ := reader.ReadString('\n')
	statusInput = strings.TrimSpace(statusInput)

	var newStatus string
	switch statusInput {
	case "1", "pending":
		newStatus = "pending"
	case "2", "in_process":
		newStatus = "in_process"
	case "3", "closed":
		newStatus = "closed"
	default:
		fmt.Println("错误: 无效的状态")
		os.Exit(1)
	}

	var processNote string
	if newStatus == "closed" {
		fmt.Printf("\n请输入处理说明:\n")
		processNote, _ = reader.ReadString('\n')
		processNote = strings.TrimSpace(processNote)
	}

	resp, err := client.UpdateStatus(feedbackID, newStatus, processNote)
	if err != nil {
		fmt.Printf("更新状态失败: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Println("\n状态更新成功！")
	} else {
		fmt.Printf("\n状态更新失败: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleStats(client *api.Client) {
	resp, err := client.GetStatistics()
	if err != nil {
		fmt.Printf("查询统计失败: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("查询统计失败: %s\n", resp.Message)
		os.Exit(1)
	}

	fmt.Println("\n统计概览")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("\n整体统计:\n")
	fmt.Printf("  总反馈数:   %d\n", resp.Data.Overall.TotalCount)
	fmt.Printf("  平均评分:   %.1f星\n", resp.Data.Overall.AverageRating)

	fmt.Println("\n按类型统计:")
	fmt.Println(strings.Repeat("-", 80))
	for _, t := range resp.Data.ByType {
		fmt.Printf("\n  类型: %s\n", t.TypeName)
		fmt.Printf("    数量: %d\n", t.Count)
		fmt.Printf("    平均评分: %.1f星\n", t.AverageRating)
	}
	fmt.Println(strings.Repeat("=", 80))
}
