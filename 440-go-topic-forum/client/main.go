package main

import (
	"os"
	"strconv"
	"strings"

	"forum/client/api"
	"forum/client/printer"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printer.PrintHelp()
		return
	}

	cmd := args[0]

	switch cmd {
	case "help", "--help", "-h":
		printer.PrintHelp()

	case "create-user":
		handleCreateUser(args[1:])

	case "get-user":
		handleGetUser(args[1:])

	case "list-users":
		handleListUsers()

	case "create-post":
		handleCreatePost(args[1:])

	case "get-post":
		handleGetPost(args[1:])

	case "list-posts":
		handleListPosts()

	case "update-post":
		handleUpdatePost(args[1:])

	case "delete-post":
		handleDeletePost(args[1:])

	case "search":
		handleSearch(args[1:])

	case "reply":
		handleReply(args[1:])

	case "set-best-reply":
		handleSetBestReply(args[1:])

	case "like":
		handleLike(args[1:])

	case "unlike":
		handleUnlike(args[1:])

	case "set-top":
		handleSetTop(args[1:])

	case "set-essence":
		handleSetEssence(args[1:])

	case "report":
		handleReport(args[1:])

	case "list-reports":
		handleListReports()

	case "review-report":
		handleReviewReport(args[1:])

	case "leaderboard":
		handleLeaderboard()

	case "tags":
		handleTags(args[1:])

	case "hotposts":
		handleHotPosts()

	default:
		printer.PrintError(&unknownCommandError{cmd: cmd})
		printer.PrintHelp()
	}
}

type unknownCommandError struct {
	cmd string
}

func (e *unknownCommandError) Error() string {
	return "未知命令: " + e.cmd
}

func parseNamedArgs(args []string) map[string]string {
	result := make(map[string]string)
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--") {
			key := strings.TrimPrefix(args[i], "--")
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				result[key] = args[i+1]
				i++
			} else {
				result[key] = "true"
			}
		}
	}
	return result
}

func handleCreateUser(args []string) {
	if len(args) < 1 {
		printer.PrintError(&missingArgError{arg: "username"})
		return
	}

	username := args[0]
	isAdmin := false

	namedArgs := parseNamedArgs(args[1:])
	if namedArgs["admin"] == "true" {
		isAdmin = true
	}

	user, err := api.CreateUser(username, isAdmin)
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintSuccess("用户创建成功")
	printer.PrintUser(user)
}

type missingArgError struct {
	arg string
}

func (e *missingArgError) Error() string {
	return "缺少参数: " + e.arg
}

func handleGetUser(args []string) {
	namedArgs := parseNamedArgs(args)
	id := namedArgs["id"]
	username := namedArgs["username"]

	if id == "" && username == "" {
		printer.PrintError(&missingArgError{arg: "--id 或 --username"})
		return
	}

	user, err := api.GetUser(id, username)
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintUser(user)
}

func handleListUsers() {
	users, err := api.ListUsers()
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintUsers(users)
}

func handleCreatePost(args []string) {
	if len(args) < 3 {
		printer.PrintError(&missingArgError{arg: "user_id, title, category"})
		return
	}

	userID := args[0]
	title := args[1]
	category := args[2]

	namedArgs := parseNamedArgs(args[3:])
	content := namedArgs["content"]
	tagsStr := namedArgs["tags"]

	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i, t := range tags {
			tags[i] = strings.TrimSpace(t)
		}
	}

	if content == "" {
		printer.PrintError(&missingArgError{arg: "--content"})
		return
	}

	post, err := api.CreatePost(userID, title, content, category, tags)
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintSuccess("帖子创建成功")
	printer.PrintPost(post)
}

func handleGetPost(args []string) {
	if len(args) < 1 {
		printer.PrintError(&missingArgError{arg: "post_id"})
		return
	}

	postID := args[0]
	namedArgs := parseNamedArgs(args[1:])
	viewerID := namedArgs["viewer"]

	detail, err := api.GetPost(postID, viewerID)
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintPostDetail(detail)
}

func handleListPosts() {
	posts, err := api.ListPosts()
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintPosts(posts)
}

func handleUpdatePost(args []string) {
	if len(args) < 2 {
		printer.PrintError(&missingArgError{arg: "post_id, user_id"})
		return
	}

	postID := args[0]
	userID := args[1]

	namedArgs := parseNamedArgs(args[2:])
	title := namedArgs["title"]
	content := namedArgs["content"]
	category := namedArgs["category"]
	tagsStr := namedArgs["tags"]

	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i, t := range tags {
			tags[i] = strings.TrimSpace(t)
		}
	} else if _, exists := namedArgs["tags"]; exists {
		tags = []string{}
	}

	post, err := api.UpdatePost(postID, userID, title, content, category, tags)
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintSuccess("帖子更新成功")
	printer.PrintPost(post)
}

func handleDeletePost(args []string) {
	if len(args) < 2 {
		printer.PrintError(&missingArgError{arg: "user_id, post_id"})
		return
	}

	userID := args[0]
	postID := args[1]

	err := api.DeletePost(userID, postID)
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintSuccess("帖子删除成功")
}

func handleSearch(args []string) {
	namedArgs := parseNamedArgs(args)
	keyword := namedArgs["keyword"]
	category := namedArgs["category"]
	tag := namedArgs["tag"]
	userID := namedArgs["user"]

	limit, _ := strconv.Atoi(namedArgs["limit"])
	offset, _ := strconv.Atoi(namedArgs["offset"])

	result, err := api.SearchPosts(keyword, category, tag, userID, limit, offset)
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintSearchResult(result)
}

func handleReply(args []string) {
	if len(args) < 3 {
		printer.PrintError(&missingArgError{arg: "user_id, post_id, content"})
		return
	}

	userID := args[0]
	postID := args[1]
	content := args[2]

	namedArgs := parseNamedArgs(args[3:])
	parentID := namedArgs["parent"]

	reply, err := api.CreateReply(userID, postID, content, parentID)
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintSuccess("回复创建成功")
	printer.PrintReply(reply, 0)
}

func handleSetBestReply(args []string) {
	if len(args) < 3 {
		printer.PrintError(&missingArgError{arg: "user_id, post_id, reply_id"})
		return
	}

	userID := args[0]
	postID := args[1]
	replyID := args[2]

	err := api.SetBestReply(userID, postID, replyID)
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintSuccess("最佳回复设置成功")
}

func handleLike(args []string) {
	if len(args) < 3 {
		printer.PrintError(&missingArgError{arg: "user_id, target_id, target_type (post/reply)"})
		return
	}

	userID := args[0]
	targetID := args[1]
	targetType := args[2]

	if targetType != "post" && targetType != "reply" {
		printer.PrintError(&invalidArgError{arg: "target_type", value: targetType, expected: "post 或 reply"})
		return
	}

	err := api.Like(userID, targetID, targetType)
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintSuccess("点赞成功")
}

type invalidArgError struct {
	arg      string
	value    string
	expected string
}

func (e *invalidArgError) Error() string {
	return "参数 " + e.arg + " 无效: " + e.value + " (期望: " + e.expected + ")"
}

func handleUnlike(args []string) {
	if len(args) < 3 {
		printer.PrintError(&missingArgError{arg: "user_id, target_id, target_type (post/reply)"})
		return
	}

	userID := args[0]
	targetID := args[1]
	targetType := args[2]

	if targetType != "post" && targetType != "reply" {
		printer.PrintError(&invalidArgError{arg: "target_type", value: targetType, expected: "post 或 reply"})
		return
	}

	err := api.Unlike(userID, targetID, targetType)
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintSuccess("取消点赞成功")
}

func handleSetTop(args []string) {
	if len(args) < 3 {
		printer.PrintError(&missingArgError{arg: "admin_id, post_id, is_top (true/false)"})
		return
	}

	adminID := args[0]
	postID := args[1]
	isTopStr := strings.ToLower(args[2])

	var isTop bool
	if isTopStr == "true" {
		isTop = true
	} else if isTopStr == "false" {
		isTop = false
	} else {
		printer.PrintError(&invalidArgError{arg: "is_top", value: isTopStr, expected: "true 或 false"})
		return
	}

	err := api.SetTop(adminID, postID, isTop)
	if err != nil {
		printer.PrintError(err)
		return
	}

	if isTop {
		printer.PrintSuccess("置顶成功")
	} else {
		printer.PrintSuccess("取消置顶成功")
	}
}

func handleSetEssence(args []string) {
	if len(args) < 3 {
		printer.PrintError(&missingArgError{arg: "admin_id, post_id, is_essence (true/false)"})
		return
	}

	adminID := args[0]
	postID := args[1]
	isEssenceStr := strings.ToLower(args[2])

	var isEssence bool
	if isEssenceStr == "true" {
		isEssence = true
	} else if isEssenceStr == "false" {
		isEssence = false
	} else {
		printer.PrintError(&invalidArgError{arg: "is_essence", value: isEssenceStr, expected: "true 或 false"})
		return
	}

	err := api.SetEssence(adminID, postID, isEssence)
	if err != nil {
		printer.PrintError(err)
		return
	}

	if isEssence {
		printer.PrintSuccess("加精成功")
	} else {
		printer.PrintSuccess("取消加精成功")
	}
}

func handleReport(args []string) {
	if len(args) < 3 {
		printer.PrintError(&missingArgError{arg: "user_id, post_id, reason"})
		return
	}

	userID := args[0]
	postID := args[1]
	reason := strings.Join(args[2:], " ")

	report, err := api.CreateReport(userID, postID, reason)
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintSuccess("举报成功")
	printer.PrintReport(report)
}

func handleListReports() {
	reports, err := api.GetPendingReports()
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintReports(reports)
}

func handleReviewReport(args []string) {
	if len(args) < 3 {
		printer.PrintError(&missingArgError{arg: "admin_id, report_id, status (resolved/rejected)"})
		return
	}

	adminID := args[0]
	reportID := args[1]
	status := strings.ToLower(args[2])

	if status != "resolved" && status != "rejected" {
		printer.PrintError(&invalidArgError{arg: "status", value: status, expected: "resolved 或 rejected"})
		return
	}

	err := api.ReviewReport(adminID, reportID, status)
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintSuccess("举报审核完成")
}

func handleLeaderboard() {
	lb, err := api.GetLeaderboard()
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintLeaderboard(lb)
}

func handleTags(args []string) {
	namedArgs := parseNamedArgs(args)
	limit, _ := strconv.Atoi(namedArgs["limit"])

	tc, err := api.GetTagCloud(limit)
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintTagCloud(tc)
}

func handleHotPosts() {
	hp, err := api.GetHotPosts()
	if err != nil {
		printer.PrintError(err)
		return
	}

	printer.PrintHotPosts(hp)
}
