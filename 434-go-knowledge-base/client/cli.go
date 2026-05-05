package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"kb/common"
	"os"
	"strconv"
	"strings"
)

type CLI struct {
	client *APIClient
	reader *bufio.Scanner
}

func NewCLI() *CLI {
	return &CLI{
		client: &APIClient{
			UserID:     "default_user",
			Department: "default",
			IsLoggedIn: true,
		},
		reader: bufio.NewScanner(os.Stdin),
	}
}

func (c *CLI) Run() {
	fmt.Println("=== 知识库客户端 ===")
	fmt.Println("输入 'help' 查看可用命令")
	fmt.Println()

	for {
		fmt.Print("kb> ")
		if !c.reader.Scan() {
			break
		}
		line := strings.TrimSpace(c.reader.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]

		switch strings.ToLower(cmd) {
		case "help", "h":
			c.printHelp()
		case "quit", "exit", "q":
			fmt.Println("再见！")
			return
		case "setuser":
			c.handleSetUser(parts)
		case "create":
			c.handleCreate(parts)
		case "get":
			c.handleGet(parts)
		case "update":
			c.handleUpdate(parts)
		case "publish":
			c.handlePublish(parts)
		case "archive":
			c.handleArchive(parts)
		case "versions":
			c.handleVersions(parts)
		case "rollback":
			c.handleRollback(parts)
		case "diff":
			c.handleDiff(parts)
		case "search":
			c.handleSearch(parts)
		case "fav":
			c.handleAddFav(parts)
		case "unfav":
			c.handleRemoveFav(parts)
		case "favs":
			c.handleGetFavs()
		case "hot":
			c.handleHot(parts)
		case "stats":
			c.handleStats()
		default:
			fmt.Printf("未知命令: %s，输入 'help' 查看帮助\n", cmd)
		}
	}
}

func (c *CLI) printHelp() {
	fmt.Println("可用命令:")
	fmt.Println("  help, h                    显示帮助信息")
	fmt.Println("  quit, exit, q              退出程序")
	fmt.Println()
	fmt.Println("用户设置:")
	fmt.Println("  setuser <user_id> <dept> [logged_in=true/false]  设置当前用户")
	fmt.Println()
	fmt.Println("文章操作:")
	fmt.Println("  create <title> <category> <access_level> [tags...]  创建新文章(草稿)")
	fmt.Println("  get <article_id>         获取文章详情")
	fmt.Println("  update <article_id> <title> [tags...]  更新草稿文章")
	fmt.Println("  publish <article_id>      发布文章")
	fmt.Println("  archive <article_id>      归档文章(不可逆)")
	fmt.Println()
	fmt.Println("版本管理:")
	fmt.Println("  versions <article_id>     查看文章的所有版本")
	fmt.Println("  rollback <article_id> <version_num>  回滚到指定版本(生成新版本)")
	fmt.Println("  diff <article_id> <v1> <v2>  对比两个版本差异")
	fmt.Println()
	fmt.Println("搜索与收藏:")
	fmt.Println("  search <keyword...>       全文搜索文章")
	fmt.Println("  fav <article_id>          添加收藏")
	fmt.Println("  unfav <article_id>        取消收藏")
	fmt.Println("  favs                      查看我的收藏")
	fmt.Println()
	fmt.Println("统计信息:")
	fmt.Println("  hot [limit]               查看热门文章排行")
	fmt.Println("  stats                     查看知识库容量统计")
	fmt.Println()
	fmt.Println("访问权限级别(access_level):")
	fmt.Println("  public      公开,所有人可见")
	fmt.Println("  logged_in   登录用户可见")
	fmt.Println("  department  指定部门可见")
	fmt.Println()
}

func (c *CLI) handleSetUser(parts []string) {
	if len(parts) < 3 {
		fmt.Println("用法: setuser <user_id> <dept> [logged_in=true/false]")
		return
	}

	c.client.UserID = parts[1]
	c.client.Department = parts[2]

	if len(parts) >= 4 {
		c.client.IsLoggedIn = parts[3] == "true" || parts[3] == "1"
	} else {
		c.client.IsLoggedIn = true
	}

	fmt.Printf("当前用户: %s, 部门: %s, 登录状态: %v\n",
		c.client.UserID, c.client.Department, c.client.IsLoggedIn)
}

func (c *CLI) handleCreate(parts []string) {
	if len(parts) < 4 {
		fmt.Println("用法: create <title> <category> <access_level> [tags...]")
		fmt.Println("  提示: 标题含空格请用引号包裹，内容将在下一步输入")
		return
	}

	title := parts[1]
	category := parts[2]
	accessLevel := common.AccessLevel(parts[3])

	var tags []string
	if len(parts) > 4 {
		tags = parts[4:]
	}

	fmt.Print("请输入文章内容 (空行结束):\n> ")
	var contentParts []string
	for {
		if !c.reader.Scan() {
			break
		}
		line := c.reader.Text()
		if line == "" {
			break
		}
		contentParts = append(contentParts, line)
	}
	content := strings.Join(contentParts, "\n")

	if content == "" {
		fmt.Println("内容不能为空")
		return
	}

	var deptIDs []string
	if accessLevel == common.AccessDepartment {
		fmt.Print("请输入可见部门ID (多个用空格分隔): ")
		if c.reader.Scan() {
			deptIDs = strings.Fields(c.reader.Text())
		}
	}

	resp, err := c.client.CreateArticle(title, content, category, tags, accessLevel, deptIDs)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("错误 [%d]: %s\n", resp.Code, resp.Message)
		return
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println("文章创建成功:")
	fmt.Println(string(data))
}

func (c *CLI) handleGet(parts []string) {
	if len(parts) < 2 {
		fmt.Println("用法: get <article_id>")
		return
	}

	resp, err := c.client.GetArticle(parts[1])
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("错误 [%d]: %s\n", resp.Code, resp.Message)
		return
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
}

func (c *CLI) handleUpdate(parts []string) {
	if len(parts) < 3 {
		fmt.Println("用法: update <article_id> <title> [tags...]")
		return
	}

	articleID := parts[1]
	title := parts[2]

	var tags []string
	if len(parts) > 3 {
		tags = parts[3:]
	}

	fmt.Print("请输入新的文章内容 (空行结束):\n> ")
	var contentParts []string
	for {
		if !c.reader.Scan() {
			break
		}
		line := c.reader.Text()
		if line == "" {
			break
		}
		contentParts = append(contentParts, line)
	}
	content := strings.Join(contentParts, "\n")

	resp, err := c.client.UpdateArticle(articleID, title, content, "", tags, common.AccessPublic, nil, nil)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("错误 [%d]: %s\n", resp.Code, resp.Message)
		return
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println("文章更新成功:")
	fmt.Println(string(data))
}

func (c *CLI) handlePublish(parts []string) {
	if len(parts) < 2 {
		fmt.Println("用法: publish <article_id>")
		return
	}

	resp, err := c.client.PublishArticle(parts[1])
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("错误 [%d]: %s\n", resp.Code, resp.Message)
		return
	}

	fmt.Println("文章发布成功")
}

func (c *CLI) handleArchive(parts []string) {
	if len(parts) < 2 {
		fmt.Println("用法: archive <article_id>")
		return
	}

	fmt.Printf("警告: 归档是不可逆操作，确定要归档文章 %s 吗? (yes/no): ", parts[1])
	if !c.reader.Scan() {
		return
	}
	confirm := strings.ToLower(c.reader.Text())
	if confirm != "yes" && confirm != "y" {
		fmt.Println("已取消")
		return
	}

	resp, err := c.client.ArchiveArticle(parts[1])
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("错误 [%d]: %s\n", resp.Code, resp.Message)
		return
	}

	fmt.Println("文章已归档")
}

func (c *CLI) handleVersions(parts []string) {
	if len(parts) < 2 {
		fmt.Println("用法: versions <article_id>")
		return
	}

	resp, err := c.client.GetVersions(parts[1])
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("错误 [%d]: %s\n", resp.Code, resp.Message)
		return
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
}

func (c *CLI) handleRollback(parts []string) {
	if len(parts) < 3 {
		fmt.Println("用法: rollback <article_id> <version_num>")
		return
	}

	versionNum, err := strconv.Atoi(parts[2])
	if err != nil {
		fmt.Println("版本号必须是数字")
		return
	}

	resp, err := c.client.RollbackArticle(parts[1], versionNum)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("错误 [%d]: %s\n", resp.Code, resp.Message)
		return
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println("回滚成功(已生成新版本):")
	fmt.Println(string(data))
}

func (c *CLI) handleDiff(parts []string) {
	if len(parts) < 4 {
		fmt.Println("用法: diff <article_id> <v1> <v2>")
		return
	}

	v1, err1 := strconv.Atoi(parts[2])
	v2, err2 := strconv.Atoi(parts[3])
	if err1 != nil || err2 != nil {
		fmt.Println("版本号必须是数字")
		return
	}

	resp, err := c.client.CompareVersions(parts[1], v1, v2)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("错误 [%d]: %s\n", resp.Code, resp.Message)
		return
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
}

func (c *CLI) handleSearch(parts []string) {
	if len(parts) < 2 {
		fmt.Println("用法: search <keyword...>")
		return
	}

	keyword := strings.Join(parts[1:], " ")

	resp, err := c.client.Search(keyword)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("错误 [%d]: %s\n", resp.Code, resp.Message)
		return
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println("搜索结果:")
	fmt.Println(string(data))
}

func (c *CLI) handleAddFav(parts []string) {
	if len(parts) < 2 {
		fmt.Println("用法: fav <article_id>")
		return
	}

	resp, err := c.client.AddFavorite(parts[1])
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("错误 [%d]: %s\n", resp.Code, resp.Message)
		return
	}

	fmt.Println("已添加收藏")
}

func (c *CLI) handleRemoveFav(parts []string) {
	if len(parts) < 2 {
		fmt.Println("用法: unfav <article_id>")
		return
	}

	resp, err := c.client.RemoveFavorite(parts[1])
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("错误 [%d]: %s\n", resp.Code, resp.Message)
		return
	}

	fmt.Println("已取消收藏")
}

func (c *CLI) handleGetFavs() {
	resp, err := c.client.GetFavorites()
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("错误 [%d]: %s\n", resp.Code, resp.Message)
		return
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println("我的收藏:")
	fmt.Println(string(data))
}

func (c *CLI) handleHot(parts []string) {
	limit := 10
	if len(parts) >= 2 {
		if l, err := strconv.Atoi(parts[1]); err == nil && l > 0 {
			limit = l
		}
	}

	resp, err := c.client.GetHotArticles(limit)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("错误 [%d]: %s\n", resp.Code, resp.Message)
		return
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println("热门文章排行:")
	fmt.Println(string(data))
}

func (c *CLI) handleStats() {
	resp, err := c.client.GetStats()
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Code != common.ErrCodeSuccess {
		fmt.Printf("错误 [%d]: %s\n", resp.Code, resp.Message)
		return
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println("知识库统计:")
	fmt.Println(string(data))
}
