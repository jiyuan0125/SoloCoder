package printer

import (
	"fmt"
	"strings"

	"forum/common"
)

func PrintUser(u *common.UserResponse) {
	fmt.Println("=== 用户信息 ===")
	fmt.Printf("ID: %s\n", u.ID)
	fmt.Printf("用户名: %s\n", u.Username)
	fmt.Printf("管理员: %t\n", u.IsAdmin)
	fmt.Printf("发帖数: %d\n", u.PostCount)
	fmt.Printf("回复数: %d\n", u.ReplyCount)
	fmt.Printf("获赞数: %d\n", u.LikeCount)
	fmt.Printf("创建时间: %s\n", u.CreatedAt)
}

func PrintUsers(users []*common.UserResponse) {
	if len(users) == 0 {
		fmt.Println("没有用户")
		return
	}
	fmt.Println("=== 用户列表 ===")
	for _, u := range users {
		fmt.Printf("ID: %s | 用户名: %s | 管理员: %t | 发帖: %d | 回复: %d | 获赞: %d\n",
			u.ID, u.Username, u.IsAdmin, u.PostCount, u.ReplyCount, u.LikeCount)
	}
}

func PrintPost(p *common.PostResponse) {
	fmt.Println("=== 帖子信息 ===")
	fmt.Printf("ID: %s\n", p.ID)
	fmt.Printf("标题: %s\n", p.Title)
	fmt.Printf("作者: %s (ID: %s)\n", p.Username, p.UserID)
	fmt.Printf("分类: %s\n", p.Category)
	if len(p.Tags) > 0 {
		fmt.Printf("标签: %s\n", strings.Join(p.Tags, ", "))
	}
	fmt.Printf("浏览量: %d | 回复数: %d | 点赞数: %d\n", p.ViewCount, p.ReplyCount, p.LikeCount)
	fmt.Printf("置顶: %t | 加精: %t\n", p.IsTop, p.IsEssence)
	if p.BestReplyID != "" {
		fmt.Printf("最佳回复ID: %s\n", p.BestReplyID)
	}
	fmt.Printf("创建时间: %s\n", p.CreatedAt)
	fmt.Printf("更新时间: %s\n", p.UpdatedAt)
	fmt.Println("--- 内容 ---")
	fmt.Println(p.Content)
}

func PrintPosts(posts []*common.PostResponse) {
	if len(posts) == 0 {
		fmt.Println("没有帖子")
		return
	}
	fmt.Println("=== 帖子列表 ===")
	for i, p := range posts {
		var flags []string
		if p.IsTop {
			flags = append(flags, "置顶")
		}
		if p.IsEssence {
			flags = append(flags, "精华")
		}
		flagStr := ""
		if len(flags) > 0 {
			flagStr = "[" + strings.Join(flags, ",") + "] "
		}
		tagStr := ""
		if len(p.Tags) > 0 {
			tagStr = " [" + strings.Join(p.Tags, ",") + "]"
		}
		fmt.Printf("%d. %s%s (ID: %s) - %s\n", i+1, flagStr, p.Title, p.ID, p.Username)
		fmt.Printf("   分类: %s%s | 浏览: %d | 回复: %d | 点赞: %d | 创建: %s\n",
			p.Category, tagStr, p.ViewCount, p.ReplyCount, p.LikeCount, p.CreatedAt)
	}
}

func PrintReply(r *common.ReplyResponse, indent int) {
	prefix := strings.Repeat("  ", indent)
	bestMark := ""
	if r.IsBest {
		bestMark = " [最佳答案]"
	}
	fmt.Printf("%s--- 第 %d 楼%s ---\n", prefix, r.Floor, bestMark)
	fmt.Printf("%sID: %s | 作者: %s (ID: %s)\n", prefix, r.ID, r.Username, r.UserID)
	fmt.Printf("%s点赞数: %d | 时间: %s\n", prefix, r.LikeCount, r.CreatedAt)
	fmt.Printf("%s内容: %s\n", prefix, r.Content)
	if len(r.Children) > 0 {
		for _, child := range r.Children {
			PrintReply(&child, indent+1)
		}
	}
}

func PrintReplies(replies []common.ReplyResponse) {
	if len(replies) == 0 {
		fmt.Println("没有回复")
		return
	}
	fmt.Println("=== 回复列表 ===")
	for _, r := range replies {
		PrintReply(&r, 0)
	}
}

func PrintPostDetail(detail *common.PostDetailResponse) {
	PrintPost(&detail.Post)
	fmt.Println()
	PrintReplies(detail.Replies)
	if len(detail.History) > 0 {
		fmt.Println()
		fmt.Println("=== 编辑历史 ===")
		for i, h := range detail.History {
			fmt.Printf("--- 版本 %d ---\n", i+1)
			fmt.Printf("编辑时间: %s | 编辑者ID: %s\n", h.EditedAt, h.EditorID)
			fmt.Printf("标题: %s\n", h.Title)
			if len(h.Tags) > 0 {
				fmt.Printf("标签: %s\n", strings.Join(h.Tags, ", "))
			}
			fmt.Printf("内容: %s\n", h.Content)
		}
	}
}

func PrintSearchResult(result *common.SearchResponse) {
	fmt.Printf("搜索结果: 共 %d 条\n", result.Total)
	PrintPosts(result.Posts)
}

func PrintReport(r *common.ReportResponse) {
	fmt.Println("=== 举报信息 ===")
	fmt.Printf("ID: %s\n", r.ID)
	fmt.Printf("帖子ID: %s\n", r.PostID)
	fmt.Printf("举报者ID: %s\n", r.ReporterID)
	fmt.Printf("原因: %s\n", r.Reason)
	fmt.Printf("状态: %s\n", r.Status)
	fmt.Printf("创建时间: %s\n", r.CreatedAt)
	if r.ReviewerID != "" {
		fmt.Printf("审核者ID: %s | 审核时间: %s\n", r.ReviewerID, r.ReviewedAt)
	}
}

func PrintReports(reports []*common.ReportResponse) {
	if len(reports) == 0 {
		fmt.Println("没有待处理的举报")
		return
	}
	fmt.Println("=== 举报列表 ===")
	for _, r := range reports {
		PrintReport(r)
		fmt.Println()
	}
}

func PrintLeaderboard(lb *common.LeaderboardResponse) {
	if len(lb.Users) == 0 {
		fmt.Println("排行榜为空")
		return
	}
	fmt.Println("=== 社区贡献排行榜 (按获赞数) ===")
	for _, u := range lb.Users {
		fmt.Printf("%d. %s (ID: %s) - 获赞: %d\n",
			u.Rank, u.Username, u.UserID, u.LikeCount)
	}
}

func PrintTagCloud(tc *common.TagCloudResponse) {
	if len(tc.Tags) == 0 {
		fmt.Println("没有标签")
		return
	}
	fmt.Println("=== 热门标签 (标签云) ===")
	for _, t := range tc.Tags {
		fmt.Printf("%s (帖子数: %d)\n", t.Name, t.PostCount)
	}
}

func PrintHotPosts(hp *common.HotPostsResponse) {
	if len(hp.Posts) == 0 {
		fmt.Println("今天没有热门帖子")
		return
	}
	fmt.Printf("=== %s 热门帖子 (浏览量0.3 + 回复数0.7) ===\n", hp.Date)
	for i, p := range hp.Posts {
		fmt.Printf("%d. %s (作者: %s) - 综合得分: %.2f\n",
			i+1, p.Title, p.Username, p.Score)
	}
}

func PrintSuccess(msg string) {
	fmt.Printf("[成功] %s\n", msg)
}

func PrintError(err error) {
	fmt.Printf("[错误] %s\n", err.Error())
}

func PrintHelp() {
	fmt.Println("=== 论坛客户端使用帮助 ===")
	fmt.Println()
	fmt.Println("用户管理:")
	fmt.Println("  create-user <username> [--admin]          创建用户")
	fmt.Println("  get-user --id <id> | --username <name>    获取用户信息")
	fmt.Println("  list-users                                  列出所有用户")
	fmt.Println()
	fmt.Println("帖子管理:")
	fmt.Println("  create-post <user_id> <title> <category> [--tags <tag1,tag2>] --content <content>")
	fmt.Println("  get-post <post_id> [--viewer <viewer_id>]  获取帖子详情")
	fmt.Println("  list-posts                                  列出所有帖子")
	fmt.Println("  update-post <post_id> <user_id> [--title <title>] [--content <content>] [--category <cat>] [--tags <tags>]")
	fmt.Println("  delete-post <user_id> <post_id>            删除帖子")
	fmt.Println()
	fmt.Println("搜索:")
	fmt.Println("  search [--keyword <word>] [--category <cat>] [--tag <tag>] [--user <id>] [--limit <n>] [--offset <n>]")
	fmt.Println()
	fmt.Println("回复管理:")
	fmt.Println("  reply <user_id> <post_id> <content> [--parent <parent_id>]")
	fmt.Println("  set-best-reply <user_id> <post_id> <reply_id>")
	fmt.Println()
	fmt.Println("点赞:")
	fmt.Println("  like <user_id> <target_id> post|reply")
	fmt.Println("  unlike <user_id> <target_id> post|reply")
	fmt.Println()
	fmt.Println("管理员功能:")
	fmt.Println("  set-top <admin_id> <post_id> true|false    置顶/取消置顶")
	fmt.Println("  set-essence <admin_id> <post_id> true|false 加精/取消加精")
	fmt.Println()
	fmt.Println("举报管理:")
	fmt.Println("  report <user_id> <post_id> <reason>        举报帖子")
	fmt.Println("  list-reports                                列出待处理举报")
	fmt.Println("  review-report <admin_id> <report_id> resolved|rejected")
	fmt.Println()
	fmt.Println("统计功能:")
	fmt.Println("  leaderboard                                 查看贡献排行榜")
	fmt.Println("  tags [--limit <n>]                          查看热门标签云")
	fmt.Println("  hotposts                                    查看今日热门帖子")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  create-user alice")
	fmt.Println("  create-user admin --admin")
	fmt.Println("  create-post <user_id> \"Hello World\" tech --tags go,forum --content \"This is my first post!\"")
	fmt.Println("  get-post <post_id>")
	fmt.Println("  search --keyword hello")
	fmt.Println("  like <user_id> <post_id> post")
}
