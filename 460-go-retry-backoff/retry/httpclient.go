package retry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// HTTPClient 是一个支持重试的HTTP客户端
type HTTPClient struct {
	// 底层HTTP客户端
	client *http.Client
	// 默认重试配置
	defaultConfig *Config
	// 基础URL
	baseURL string
}

// NewHTTPClient 创建一个新的支持重试的HTTP客户端
func NewHTTPClient(baseURL string, opts ...Option) *HTTPClient {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	return &HTTPClient{
		client:        &http.Client{},
		defaultConfig: config,
		baseURL:       baseURL,
	}
}

// SetClient 设置底层HTTP客户端
func (c *HTTPClient) SetClient(client *http.Client) {
	c.client = client
}

// Get 执行带重试的GET请求
func (c *HTTPClient) Get(ctx context.Context, path string, opts ...Option) (*http.Response, error) {
	return c.Do(ctx, http.MethodGet, path, nil, opts...)
}

// Post 执行带重试的POST请求
func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}, opts ...Option) (*http.Response, error) {
	return c.Do(ctx, http.MethodPost, path, body, opts...)
}

// Put 执行带重试的PUT请求
func (c *HTTPClient) Put(ctx context.Context, path string, body interface{}, opts ...Option) (*http.Response, error) {
	return c.Do(ctx, http.MethodPut, path, body, opts...)
}

// Delete 执行带重试的DELETE请求
func (c *HTTPClient) Delete(ctx context.Context, path string, opts ...Option) (*http.Response, error) {
	return c.Do(ctx, http.MethodDelete, path, nil, opts...)
}

// Do 执行带重试的HTTP请求
func (c *HTTPClient) Do(ctx context.Context, method, path string, body interface{}, opts ...Option) (*http.Response, error) {
	// 合并配置
	config := *c.defaultConfig
	for _, opt := range opts {
		opt(&config)
	}

	// 构建完整URL
	fullURL := c.buildURL(path)

	// 定义重试函数
	retryFn := func(ctx context.Context) (interface{}, error) {
		var reqBody io.Reader
		if body != nil {
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, NewNonRetryableError(err)
			}
			reqBody = bytes.NewReader(jsonBody)
		}

		// 创建请求
		req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
		if err != nil {
			return nil, NewNonRetryableError(err)
		}

		// 设置Content-Type
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		// 执行请求
		resp, err := c.client.Do(req)
		if err != nil {
			// 网络错误，可能可以重试
			return nil, err
		}

		// 检查HTTP状态码
		if resp.StatusCode >= 400 {
			// 读取响应体以获取错误信息
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			httpErr := &HTTPError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(bodyBytes)),
			}

			// 根据状态码决定是否重试
			if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
				// 5xx和429可以重试
				return nil, httpErr
			}
			// 其他4xx错误不重试
			return nil, NewNonRetryableError(httpErr)
		}

		return resp, nil
	}

	// 执行带重试的请求
	result, err := DoRetry(ctx, retryFn, &config)
	if err != nil {
		return nil, err
	}

	if resp, ok := result.(*http.Response); ok {
		return resp, nil
	}

	return nil, fmt.Errorf("unexpected result type")
}

// buildURL 构建完整的URL
func (c *HTTPClient) buildURL(path string) string {
	if c.baseURL == "" {
		return path
	}

	base, err := url.Parse(c.baseURL)
	if err != nil {
		return path
	}

	rel, err := url.Parse(path)
	if err != nil {
		return path
	}

	return base.ResolveReference(rel).String()
}

// DoWithResponse 执行带重试的HTTP请求并解析JSON响应
func (c *HTTPClient) DoWithResponse(ctx context.Context, method, path string, body interface{}, result interface{}, opts ...Option) error {
	resp, err := c.Do(ctx, method, path, body, opts...)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 解析响应
	if result != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, result); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetWithResponse 执行带重试的GET请求并解析JSON响应
func (c *HTTPClient) GetWithResponse(ctx context.Context, path string, result interface{}, opts ...Option) error {
	return c.DoWithResponse(ctx, http.MethodGet, path, nil, result, opts...)
}

// PostWithResponse 执行带重试的POST请求并解析JSON响应
func (c *HTTPClient) PostWithResponse(ctx context.Context, path string, body interface{}, result interface{}, opts ...Option) error {
	return c.DoWithResponse(ctx, http.MethodPost, path, body, result, opts...)
}

// PutWithResponse 执行带重试的PUT请求并解析JSON响应
func (c *HTTPClient) PutWithResponse(ctx context.Context, path string, body interface{}, result interface{}, opts ...Option) error {
	return c.DoWithResponse(ctx, http.MethodPut, path, body, result, opts...)
}

// DeleteWithResponse 执行带重试的DELETE请求并解析JSON响应
func (c *HTTPClient) DeleteWithResponse(ctx context.Context, path string, result interface{}, opts ...Option) error {
	return c.DoWithResponse(ctx, http.MethodDelete, path, nil, result, opts...)
}
