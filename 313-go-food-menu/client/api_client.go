package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/example/food-menu/common"
)

type APIClient struct {
	BaseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{BaseURL: baseURL}
}

func (c *APIClient) doRequest(method, path string, body interface{}) (*common.APIResponse, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求失败: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	var apiResp common.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &apiResp, nil
}

func (c *APIClient) CreateDish(req *common.CreateDishRequest) (*common.Dish, error) {
	resp, err := c.doRequest(http.MethodPost, "/api/dishes", req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	var dish common.Dish
	if err := json.Unmarshal(dataBytes, &dish); err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	return &dish, nil
}

func (c *APIClient) UpdateDish(id string, req *common.UpdateDishRequest) (*common.Dish, error) {
	resp, err := c.doRequest(http.MethodPut, "/api/dishes/"+id, req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	var dish common.Dish
	if err := json.Unmarshal(dataBytes, &dish); err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	return &dish, nil
}

func (c *APIClient) SetDishOnSale(id string) error {
	resp, err := c.doRequest(http.MethodPut, "/api/dishes/"+id+"/on_sale", nil)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func (c *APIClient) SetDishOffSale(id string) error {
	resp, err := c.doRequest(http.MethodPut, "/api/dishes/"+id+"/off_sale", nil)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func (c *APIClient) GetDish(id string) (*common.Dish, error) {
	resp, err := c.doRequest(http.MethodGet, "/api/dishes/"+id, nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	var dish common.Dish
	if err := json.Unmarshal(dataBytes, &dish); err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	return &dish, nil
}

func (c *APIClient) ListDishes(category string, onlyOnSale bool) (*common.DishListResponse, error) {
	path := "/api/dishes"
	if category != "" {
		path += "?category=" + category
	}
	if onlyOnSale {
		if category != "" {
			path += "&"
		} else {
			path += "?"
		}
		path += "status=on_sale"
	}

	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	var listResp common.DishListResponse
	if err := json.Unmarshal(dataBytes, &listResp); err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	return &listResp, nil
}

func (c *APIClient) SearchDishes(category, keyword string) (*common.DishListResponse, error) {
	path := "/api/dishes/search"
	first := true
	if category != "" {
		path += "?category=" + category
		first = false
	}
	if keyword != "" {
		if first {
			path += "?"
		} else {
			path += "&"
		}
		path += "keyword=" + keyword
	}

	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	var listResp common.DishListResponse
	if err := json.Unmarshal(dataBytes, &listResp); err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	return &listResp, nil
}

func (c *APIClient) GetCategorySummary() (*common.CategorySummaryResponse, error) {
	resp, err := c.doRequest(http.MethodGet, "/api/categories/summary", nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	var summaryResp common.CategorySummaryResponse
	if err := json.Unmarshal(dataBytes, &summaryResp); err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	return &summaryResp, nil
}

func (c *APIClient) SetTodayRecommend(dishIDs []string) (*common.RecommendResponse, error) {
	req := common.SetRecommendRequest{DishIDs: dishIDs}
	resp, err := c.doRequest(http.MethodPost, "/api/recommend", &req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	var recommendResp common.RecommendResponse
	if err := json.Unmarshal(dataBytes, &recommendResp); err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	return &recommendResp, nil
}

func (c *APIClient) GetTodayRecommend() (*common.RecommendResponse, error) {
	resp, err := c.doRequest(http.MethodGet, "/api/recommend", nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	var recommendResp common.RecommendResponse
	if err := json.Unmarshal(dataBytes, &recommendResp); err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	return &recommendResp, nil
}

func (c *APIClient) GetCategories() ([]string, error) {
	resp, err := c.doRequest(http.MethodGet, "/api/categories", nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	var categories []string
	if err := json.Unmarshal(dataBytes, &categories); err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	return categories, nil
}
