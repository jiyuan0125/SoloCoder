package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"bplus-tree/common"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

func (c *Client) Insert(key string, value string) {
	req := common.InsertRequest{
		Key:   key,
		Value: value,
	}

	var resp common.InsertResponse
	if err := c.post("/insert", req, &resp); err != nil {
		fmt.Printf("插入失败: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("成功插入: %s = %s\n", key, value)
	} else {
		fmt.Println("插入失败")
	}
}

func (c *Client) Search(key string) {
	req := common.SearchRequest{Key: key}

	var resp common.SearchResponse
	if err := c.post("/search", req, &resp); err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return
	}

	if resp.Found {
		fmt.Printf("找到: %s = %v\n", resp.Key, resp.Value)
	} else {
		fmt.Printf("未找到键: %s\n", key)
	}
}

func (c *Client) Delete(key string) {
	req := common.DeleteRequest{Key: key}

	var resp common.DeleteResponse
	if err := c.post("/delete", req, &resp); err != nil {
		fmt.Printf("删除失败: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("成功删除: %s\n", key)
	} else {
		fmt.Printf("键不存在: %s\n", key)
	}
}

func (c *Client) Range(start string, end string) {
	req := common.RangeRequest{
		Start: start,
		End:   end,
	}

	var resp common.RangeResponse
	if err := c.post("/range", req, &resp); err != nil {
		fmt.Printf("范围查询失败: %v\n", err)
		return
	}

	if resp.Count == 0 {
		fmt.Println("范围查询结果为空")
		return
	}

	fmt.Printf("范围查询结果 (%d 条):\n", resp.Count)
	for i, item := range resp.Items {
		fmt.Printf("  %d. %s = %v\n", i+1, item.Key, item.Value)
	}
}

func (c *Client) Scan(direction string, batchSize int, startOffset int) {
	req := common.ScanRequest{
		Direction:   direction,
		BatchSize:   batchSize,
		StartOffset: startOffset,
	}

	var resp common.ScanResponse
	if err := c.post("/scan", req, &resp); err != nil {
		fmt.Printf("扫描失败: %v\n", err)
		return
	}

	if resp.Count == 0 {
		fmt.Println("扫描结果为空")
		return
	}

	dirName := "正向"
	if direction == "backward" {
		dirName = "反向"
	}

	fmt.Printf("%s扫描结果 (%d 条", dirName, resp.Count)
	if resp.HasMore {
		fmt.Printf(", 还有更多数据")
	}
	fmt.Println("):")

	for i, item := range resp.Items {
		fmt.Printf("  %d. %s = %v\n", startOffset+i+1, item.Key, item.Value)
	}
}

func (c *Client) post(endpoint string, req interface{}, resp interface{}) error {
	url := c.baseURL + endpoint

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	httpResp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("服务器返回错误状态: %d", httpResp.StatusCode)
	}

	if err := json.Unmarshal(respBody, resp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	return nil
}
