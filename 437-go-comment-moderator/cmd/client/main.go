package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/comment-moderator/internal/client"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	host := flag.String("host", "localhost", "服务端主机")
	port := flag.Int("port", 8080, "服务端端口")

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]
	subArgs := args[1:]

	apiClient := client.NewAPIClient(*host, *port)

	switch cmd {
	case "health":
		handleHealth(apiClient)
	case "submit":
		handleSubmit(apiClient, subArgs)
	case "edit":
		handleEdit(apiClient, subArgs)
	case "report":
		handleReport(apiClient, subArgs)
	case "view-rejected":
		handleViewRejected(apiClient, subArgs)
	case "pending":
		handlePending(apiClient, subArgs)
	case "approve":
		handleApprove(apiClient, subArgs)
	case "reject":
		handleReject(apiClient, subArgs)
	case "add-sensitive":
		handleAddSensitive(apiClient, subArgs)
	case "remove-sensitive":
		handleRemoveSensitive(apiClient, subArgs)
	case "list-sensitive":
		handleListSensitive(apiClient)
	case "register-moderator":
		handleRegisterModerator(apiClient, subArgs)
	case "moderator-stats":
		handleModeratorStats(apiClient)
	case "audit-logs":
		handleAuditLogs(apiClient)
	case "help":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`评论审核客户端使用说明:

用法: client [全局选项] <命令> [命令选项]

全局选项:
  --host <主机>    服务端主机 (默认: localhost)
  --port <端口>    服务端端口 (默认: 8080)

用户操作:
  submit --user <用户ID> --content <评论内容>    提交评论
  edit --comment <评论ID> --user <用户ID> --content <新内容>  编辑评论
  report --comment <评论ID> --user <用户ID> --reason <举报原因>  举报评论
  view-rejected --comment <评论ID> --user <用户ID>  查看被拒绝的评论原因

审核员操作:
  pending --moderator <审核员ID> [--limit <数量>] [--offset <偏移量>]  获取待审核评论
  approve --moderator <审核员ID> --comments <评论ID1,评论ID2...>  批量通过评论
  reject --moderator <审核员ID> --comments <评论ID1,评论ID2...> --reason <原因>  批量拒绝评论

管理员操作:
  add-sensitive --admin <管理员ID> --word <敏感词> --level <级别>  添加敏感词
  remove-sensitive --admin <管理员ID> --word <敏感词>  删除敏感词
  list-sensitive                                           列出所有敏感词
  register-moderator --admin <管理员ID> --name <审核员姓名>  注册审核员
  moderator-stats                                          查看审核员统计
  audit-logs                                            查看审计日志

系统操作:
  health                                        健康检查
  help                                          显示帮助

敏感词级别: severe(严重), medium(中等), mild(轻微)

示例:
  client --port 8737 submit --user test --content "hello"
  client -port 8737 submit --user test --content "hello"
`)
}

func handleHealth(c *client.APIClient) {
	result, err := c.HealthCheck()
	if err != nil {
		fmt.Printf("健康检查失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(result)
}

func handleSubmit(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("submit", flag.ExitOnError)
	userID := fs.String("user", "", "用户ID")
	content := fs.String("content", "", "评论内容")
	fs.Parse(args)

	if *userID == "" || *content == "" {
		fmt.Println("使用: submit --user <用户ID> --content <评论内容>")
		os.Exit(1)
	}

	resp, err := c.SubmitComment(*userID, *content)
	if err != nil {
		fmt.Printf("提交失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功: %s\n", resp.Message)
	if resp.CommentID != "" {
		fmt.Printf("评论ID: %s\n", resp.CommentID)
	}
	fmt.Printf("状态: %s\n", resp.Status)
}

func handleEdit(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("edit", flag.ExitOnError)
	commentID := fs.String("comment", "", "评论ID")
	userID := fs.String("user", "", "用户ID")
	content := fs.String("content", "", "新内容")
	fs.Parse(args)

	if *commentID == "" || *userID == "" || *content == "" {
		fmt.Println("使用: edit --comment <评论ID> --user <用户ID> --content <新内容>")
		os.Exit(1)
	}

	resp, err := c.EditComment(*commentID, *userID, *content)
	if err != nil {
		fmt.Printf("编辑失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功: %s\n", resp.Message)
	fmt.Printf("状态: %s\n", resp.Status)
}

func handleReport(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	commentID := fs.String("comment", "", "评论ID")
	userID := fs.String("user", "", "用户ID")
	reason := fs.String("reason", "", "举报原因")
	fs.Parse(args)

	if *commentID == "" || *userID == "" {
		fmt.Println("使用: report --comment <评论ID> --user <用户ID> [--reason <举报原因>]")
		os.Exit(1)
	}

	resp, err := c.ReportComment(*commentID, *userID, *reason)
	if err != nil {
		fmt.Printf("举报失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功: %s\n", resp.Message)
	fmt.Printf("举报次数: %d\n", resp.ReportCount)
}

func handleViewRejected(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("view-rejected", flag.ExitOnError)
	commentID := fs.String("comment", "", "评论ID")
	userID := fs.String("user", "", "用户ID")
	fs.Parse(args)

	if *commentID == "" || *userID == "" {
		fmt.Println("使用: view-rejected --comment <评论ID> --user <用户ID>")
		os.Exit(1)
	}

	resp, err := c.ViewRejectedComment(*userID, *commentID)
	if err != nil {
		fmt.Printf("获取失败: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("失败: %s\n", resp.Message)
		os.Exit(1)
	}

	fmt.Println("拒绝原因:")
	fmt.Println(resp.RejectReason)
}

func handlePending(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("pending", flag.ExitOnError)
	moderatorID := fs.String("moderator", "", "审核员ID")
	limit := fs.Int("limit", 50, "返回数量(最大50)")
	offset := fs.Int("offset", 0, "偏移量")
	fs.Parse(args)

	if *moderatorID == "" {
		fmt.Println("使用: pending --moderator <审核员ID> [--limit <数量>] [--offset <偏移量>]")
		os.Exit(1)
	}

	resp, err := c.GetPendingComments(*moderatorID, *limit, *offset)
	if err != nil {
		fmt.Printf("获取失败: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("失败: %s\n", resp.Message)
		os.Exit(1)
	}

	fmt.Printf("待审核评论总数: %d\n", resp.Total)
	fmt.Printf("当前页数量: %d\n", len(resp.Comments))
	fmt.Printf("是否还有更多: %v\n", resp.HasMore)
	fmt.Println()

	for i, comment := range resp.Comments {
		fmt.Printf("=== 评论 %d ===\n", i+1)
		fmt.Printf("  ID: %s\n", comment.ID)
		fmt.Printf("  用户ID: %s\n", comment.UserID)
		fmt.Printf("  内容: %s\n", truncate(comment.Content, 50))
		fmt.Printf("  状态: %s\n", comment.Status)
		fmt.Printf("  创建时间: %s\n", comment.CreatedAt.Format("2006-01-02 15:04:05"))
		if comment.ReportCount > 0 {
			fmt.Printf("  举报次数: %d\n", comment.ReportCount)
		}
		fmt.Println()
	}
}

func handleApprove(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("approve", flag.ExitOnError)
	moderatorID := fs.String("moderator", "", "审核员ID")
	commentsStr := fs.String("comments", "", "评论ID列表(用逗号分隔)")
	fs.Parse(args)

	if *moderatorID == "" || *commentsStr == "" {
		fmt.Println("使用: approve --moderator <审核员ID> --comments <评论ID1,评论ID2...>")
		os.Exit(1)
	}

	commentIDs := strings.Split(*commentsStr, ",")
	for i, id := range commentIDs {
		commentIDs[i] = strings.TrimSpace(id)
	}

	resp, err := c.BatchApprove(*moderatorID, commentIDs)
	if err != nil {
		fmt.Printf("操作失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功处理: %d 条\n", resp.Processed)
	if len(resp.FailedIDs) > 0 {
		fmt.Printf("处理失败: %v\n", resp.FailedIDs)
	}
}

func handleReject(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("reject", flag.ExitOnError)
	moderatorID := fs.String("moderator", "", "审核员ID")
	commentsStr := fs.String("comments", "", "评论ID列表(用逗号分隔)")
	reason := fs.String("reason", "", "拒绝原因")
	fs.Parse(args)

	if *moderatorID == "" || *commentsStr == "" {
		fmt.Println("使用: reject --moderator <审核员ID> --comments <评论ID1,评论ID2...> --reason <拒绝原因")
		os.Exit(1)
	}

	commentIDs := strings.Split(*commentsStr, ",")
	for i, id := range commentIDs {
		commentIDs[i] = strings.TrimSpace(id)
	}

	resp, err := c.BatchReject(*moderatorID, commentIDs, *reason)
	if err != nil {
		fmt.Printf("操作失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功处理: %d 条\n", resp.Processed)
	if len(resp.FailedIDs) > 0 {
		fmt.Printf("处理失败: %v\n", resp.FailedIDs)
	}
}

func handleAddSensitive(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("add-sensitive", flag.ExitOnError)
	adminID := fs.String("admin", "", "管理员ID")
	word := fs.String("word", "", "敏感词")
	levelStr := fs.String("level", "", "级别 (severe/medium/mild)")
	fs.Parse(args)

	if *adminID == "" || *word == "" || *levelStr == "" {
		fmt.Println("使用: add-sensitive --admin <管理员ID> --word <敏感词> --level <级别>")
		os.Exit(1)
	}

	level, err := client.ParseSensitiveWordLevel(*levelStr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	resp, err := c.AddSensitiveWord(*adminID, *word, level)
	if err != nil {
		fmt.Printf("添加失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功: %s\n", resp.Message)
}

func handleRemoveSensitive(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("remove-sensitive", flag.ExitOnError)
	adminID := fs.String("admin", "", "管理员ID")
	word := fs.String("word", "", "敏感词")
	fs.Parse(args)

	if *adminID == "" || *word == "" {
		fmt.Println("使用: remove-sensitive --admin <管理员ID> --word <敏感词>")
		os.Exit(1)
	}

	resp, err := c.RemoveSensitiveWord(*adminID, *word)
	if err != nil {
		fmt.Printf("删除失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功: %s\n", resp.Message)
}

func handleListSensitive(c *client.APIClient) {
	resp, err := c.GetSensitiveWords()
	if err != nil {
		fmt.Printf("获取失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("敏感词总数: %d\n", len(resp.Words))
	fmt.Println()

	for i, word := range resp.Words {
		fmt.Printf("%d. %s [%s]\n", i+1, word.Word, word.Level)
	}
}

func handleRegisterModerator(c *client.APIClient, args []string) {
	fs := flag.NewFlagSet("register-moderator", flag.ExitOnError)
	adminID := fs.String("admin", "", "管理员ID")
	name := fs.String("name", "", "审核员姓名")
	fs.Parse(args)

	if *adminID == "" || *name == "" {
		fmt.Println("使用: register-moderator --admin <管理员ID> --name <审核员姓名>")
		os.Exit(1)
	}

	resp, err := c.RegisterModerator(*adminID, *name)
	if err != nil {
		fmt.Printf("注册失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功: %s\n", resp.Message)
	fmt.Printf("审核员ID: %s\n", resp.ModeratorID)
}

func handleModeratorStats(c *client.APIClient) {
	resp, err := c.GetModeratorStats()
	if err != nil {
		fmt.Printf("获取失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("审核员总数: %d\n", len(resp.Moderators))
	fmt.Println()

	for id, stats := range resp.Moderators {
		fmt.Printf("审核员ID: %s\n", id)
		fmt.Printf("  姓名: %s\n", stats.Name)
		fmt.Printf("  今日处理: %d 条\n", stats.DailyCount)
		fmt.Printf("  总处理: %d 条\n", stats.TotalCount)
		fmt.Println()
	}
}

func handleAuditLogs(c *client.APIClient) {
	resp, err := c.GetAuditLogs()
	if err != nil {
		fmt.Printf("获取失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("审计日志总数: %d\n", len(resp.Logs))
	fmt.Println()

	for i, log := range resp.Logs {
		fmt.Printf("=== 日志 %d ===\n", i+1)
		fmt.Printf("  ID: %s\n", log.ID)
		fmt.Printf("  操作人: %s\n", log.ModeratorID)
		fmt.Printf("  操作: %s\n", log.Action)
		if log.CommentID != "" {
			fmt.Printf("  评论ID: %s\n", log.CommentID)
		}
		fmt.Printf("  时间: %s\n", log.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("  详情: %s\n", log.Details)
		fmt.Println()
	}
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}
