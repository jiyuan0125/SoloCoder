package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ip-location/common"
)

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: serverURL,
	}
}

func (c *Client) post(endpoint string, reqBody, respBody interface{}) error {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	url := c.serverURL + endpoint
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("服务端返回错误: %s", string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if err := json.Unmarshal(body, respBody); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	return nil
}

func (c *Client) get(endpoint string, respBody interface{}) error {
	url := c.serverURL + endpoint
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("服务端返回错误: %s", string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if err := json.Unmarshal(body, respBody); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	return nil
}

func (c *Client) Query(ip string) error {
	req := common.QueryRequest{IP: ip}
	var resp common.QueryResponse

	if err := c.post("/query", &req, &resp); err != nil {
		return err
	}

	printQueryResult(resp)
	return nil
}

func (c *Client) BatchQuery(ips []string) error {
	req := common.BatchQueryRequest{IPs: ips}
	var resp common.BatchQueryResponse

	if err := c.post("/batch", &req, &resp); err != nil {
		return err
	}

	printBatchResult(resp)
	return nil
}

func (c *Client) SameProvince(ip1, ip2 string) error {
	req := common.SameProvinceRequest{IP1: ip1, IP2: ip2}
	var resp common.SameProvinceResponse

	if err := c.post("/same-province", &req, &resp); err != nil {
		return err
	}

	printSameProvinceResult(resp)
	return nil
}

func (c *Client) IPInfo(ip string) error {
	req := common.IPInfoRequest{IP: ip}
	var resp common.IPInfoResponse

	if err := c.post("/info", &req, &resp); err != nil {
		return err
	}

	printIPInfoResult(resp)
	return nil
}

func (c *Client) Health() error {
	var resp map[string]interface{}

	if err := c.get("/health", &resp); err != nil {
		return err
	}

	printHealthResult(resp)
	return nil
}

func printQueryResult(resp common.QueryResponse) {
	fmt.Printf("IP: %s\n", resp.IP)
	if resp.Success {
		fmt.Printf("国家: %s\n", resp.Country)
		fmt.Printf("省份: %s\n", resp.Province)
		fmt.Printf("城市: %s\n", resp.City)
	} else {
		fmt.Printf("错误: %s\n", resp.Error)
	}
}

func printBatchResult(resp common.BatchQueryResponse) {
	if !resp.Success {
		fmt.Printf("错误: %s\n", resp.Error)
		return
	}

	for i, r := range resp.Results {
		if i > 0 {
			fmt.Println("---")
		}
		fmt.Printf("IP: %s\n", r.IP)
		if r.Error != "" {
			fmt.Printf("错误: %s\n", r.Error)
		} else {
			fmt.Printf("  国家: %s\n", r.Country)
			fmt.Printf("  省份: %s\n", r.Province)
			fmt.Printf("  城市: %s\n", r.City)
		}
	}
}

func printSameProvinceResult(resp common.SameProvinceResponse) {
	if resp.Success {
		if resp.Same {
			fmt.Println("两个IP属于同一个省份")
		} else {
			fmt.Println("两个IP不属于同一个省份")
		}
	} else {
		fmt.Printf("错误: %s\n", resp.Error)
	}
}

func printIPInfoResult(resp common.IPInfoResponse) {
	fmt.Printf("IP: %s\n", resp.IP)
	if resp.Success {
		fmt.Printf("类型: %s\n", resp.Type)
		fmt.Printf("内网地址: %v\n", resp.IsPrivate)
	} else {
		fmt.Printf("错误: %s\n", resp.Error)
	}
}

func printHealthResult(resp map[string]interface{}) {
	fmt.Printf("状态: %v\n", resp["status"])
	fmt.Printf("已加载IP段数: %v\n", resp["rangeCount"])
}
