package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"metric-aggregator/pkg/common"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	serverURL  string
	httpClient *http.Client
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL:  serverURL,
		httpClient: &http.Client{},
	}
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("请求体序列化失败: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}
	
	req, err := http.NewRequest(method, c.serverURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	
	if resp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return nil, fmt.Errorf("服务器错误 [%s]: %s", errResp.Code, errResp.Message)
		}
		return nil, fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}
	
	return respBody, nil
}

func (c *Client) CreateMetric(name string) error {
	req := common.CreateMetricRequest{Metric: name}
	_, err := c.doRequest(http.MethodPost, "/metrics", req)
	return err
}

func (c *Client) DeleteMetric(name string) error {
	_, err := c.doRequest(http.MethodDelete, "/metrics/"+url.PathEscape(name), nil)
	return err
}

func (c *Client) ListMetrics() ([]string, error) {
	respBody, err := c.doRequest(http.MethodGet, "/metrics", nil)
	if err != nil {
		return nil, err
	}
	
	var result struct {
		Metrics []string `json:"metrics"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	
	return result.Metrics, nil
}

func (c *Client) GetMetricInfo(name string) (map[string]interface{}, error) {
	respBody, err := c.doRequest(http.MethodGet, "/metrics/"+url.PathEscape(name), nil)
	if err != nil {
		return nil, err
	}
	
	var info map[string]interface{}
	if err := json.Unmarshal(respBody, &info); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	
	return info, nil
}

func (c *Client) Report(points []common.DataPoint) (*common.ReportResponse, error) {
	req := common.ReportRequest{Points: points}
	respBody, err := c.doRequest(http.MethodPost, "/report", req)
	if err != nil {
		return nil, err
	}
	
	var resp common.ReportResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	
	return &resp, nil
}

func (c *Client) Query(
	metrics []string,
	start, end int64,
	granularity common.Granularity,
	aggregation common.AggregationType,
) (*common.QueryResponse, error) {
	req := common.QueryRequest{
		Metrics:     metrics,
		Start:       start,
		End:         end,
		Granularity: granularity,
		Aggregation: aggregation,
	}
	
	respBody, err := c.doRequest(http.MethodPost, "/query", req)
	if err != nil {
		return nil, err
	}
	
	var resp common.QueryResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	
	return &resp, nil
}

func (c *Client) Health() error {
	_, err := c.doRequest(http.MethodGet, "/health", nil)
	return err
}

func ParseTimestamp(ts string) (int64, error) {
	if ts == "now" {
		return time.Now().Unix(), nil
	}
	
	if val, err := strconv.ParseInt(ts, 10, 64); err == nil {
		return val, nil
	}
	
	return 0, fmt.Errorf("无效的时间戳: %s", ts)
}

func ParseValue(v string) (float64, error) {
	return strconv.ParseFloat(v, 64)
}
