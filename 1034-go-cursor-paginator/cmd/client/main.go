package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"cursor-paginator/pkg/api"
)

const serverURL = "http://localhost:8080"

type Client struct {
	baseURL    string
	nextCursor string
	prevCursor string
	hasNext    bool
	hasPrev    bool
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

func (c *Client) fetchUsers(cursor, previous, sortField string, sortOrder api.SortOrder, limit int) (*api.PageResponse, error) {
	query := url.Values{}
	if cursor != "" {
		query.Set("cursor", cursor)
	}
	if previous != "" {
		query.Set("previous", previous)
	}
	if sortField != "" {
		query.Set("sort_field", sortField)
	}
	if sortOrder != "" {
		query.Set("sort_order", string(sortOrder))
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}

	url := fmt.Sprintf("%s/users?%s", c.baseURL, query.Encode())
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error: %s - %s", resp.Status, string(body))
	}

	var pageResp api.PageResponse
	if err := json.NewDecoder(resp.Body).Decode(&pageResp); err != nil {
		return nil, err
	}

	c.nextCursor = pageResp.NextCursor
	c.prevCursor = pageResp.PrevCursor
	c.hasNext = pageResp.HasNext
	c.hasPrev = pageResp.HasPrev

	return &pageResp, nil
}

func (c *Client) printResponse(resp *api.PageResponse) {
	fmt.Printf("\n=== 分页结果 ===\n")
	fmt.Printf("总数: %d, 每页: %d\n", resp.Total, resp.Limit)
	fmt.Printf("有下一页: %v, 有上一页: %v\n", resp.HasNext, resp.HasPrev)
	if resp.NextCursor != "" {
		fmt.Printf("下一页游标: %s\n", resp.NextCursor)
	}
	if resp.PrevCursor != "" {
		fmt.Printf("上一页游标: %s\n", resp.PrevCursor)
	}

	fmt.Println("\n用户列表:")
	for i, data := range resp.Data {
		var user api.User
		if err := json.Unmarshal(data, &user); err != nil {
			fmt.Printf("  [%d] 解析错误: %v\n", i+1, err)
			continue
		}
		fmt.Printf("  [%d] ID: %d, 名称: %s, 邮箱: %s\n",
			i+1, user.ID, user.Name, user.Email)
	}
	fmt.Println()
}

func (c *Client) interactiveMode(sortField string, sortOrder api.SortOrder, limit int) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n=== 交互模式 ===")
	fmt.Println("命令: n=下一页, p=上一页, j=跳转到游标, q=退出")

	currentCursor := ""
	for {
		resp, err := c.fetchUsers(currentCursor, "", sortField, sortOrder, limit)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			return
		}
		c.printResponse(resp)

		fmt.Print("请输入命令 (n/p/j/q): ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(strings.ToLower(cmd))

		switch cmd {
		case "n", "next":
			if !c.hasNext {
				fmt.Println("没有下一页了！")
			} else {
				currentCursor = c.nextCursor
			}
		case "p", "prev", "previous":
			if !c.hasPrev {
				fmt.Println("没有上一页了！")
			} else {
				currentCursor = c.prevCursor
			}
		case "j", "jump":
			fmt.Print("请输入游标: ")
			jumpCursor, _ := reader.ReadString('\n')
			jumpCursor = strings.TrimSpace(jumpCursor)
			if jumpCursor != "" {
				currentCursor = jumpCursor
			}
		case "q", "quit", "exit":
			fmt.Println("退出交互模式")
			return
		default:
			fmt.Println("未知命令，请重新输入")
		}
	}
}

func main() {
	var (
		cursor     string
		previous   string
		sortField  string
		sortOrder  string
		limit      int
		interactive bool
	)

	flag.StringVar(&cursor, "cursor", "", "下一页游标")
	flag.StringVar(&previous, "previous", "", "上一页游标")
	flag.StringVar(&sortField, "sort", "id", "排序字段 (id, name, email, created_at)")
	flag.StringVar(&sortOrder, "order", "asc", "排序方向 (asc, desc)")
	flag.IntVar(&limit, "limit", 10, "每页条数 (1-100)")
	flag.BoolVar(&interactive, "interactive", false, "交互模式")
	flag.Parse()

	client := NewClient(serverURL)

	if interactive {
		client.interactiveMode(sortField, api.SortOrder(sortOrder), limit)
		return
	}

	resp, err := client.fetchUsers(cursor, previous, sortField, api.SortOrder(sortOrder), limit)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	client.printResponse(resp)
}
