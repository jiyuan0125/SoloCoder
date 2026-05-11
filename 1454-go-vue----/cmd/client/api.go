package main

import (
	"bytes"
	"canteen/internal/common"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{},
	}
}

func (c *Client) do(method, path string, body interface{}, result interface{}) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var apiResp common.Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return err
	}

	if !apiResp.Success {
		return fmt.Errorf("%s", apiResp.Error)
	}

	if result != nil && apiResp.Data != nil {
		data, err := json.Marshal(apiResp.Data)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, result)
	}
	return nil
}

func (c *Client) CreateEmployee(req common.CreateEmployeeRequest) (*common.Employee, error) {
	var emp common.Employee
	if err := c.do(http.MethodPost, "/api/employees", req, &emp); err != nil {
		return nil, err
	}
	return &emp, nil
}

func (c *Client) ListEmployees() ([]common.Employee, error) {
	var employees []common.Employee
	if err := c.do(http.MethodGet, "/api/employees", nil, &employees); err != nil {
		return nil, err
	}
	return employees, nil
}

func (c *Client) GetEmployee(id string) (*common.Employee, error) {
	var emp common.Employee
	if err := c.do(http.MethodGet, "/api/employees/"+id, nil, &emp); err != nil {
		return nil, err
	}
	return &emp, nil
}

func (c *Client) Recharge(req common.RechargeRequest) (*common.RechargeRecord, error) {
	var record common.RechargeRecord
	if err := c.do(http.MethodPost, "/api/recharge", req, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (c *Client) ListRechargeRecords(employeeID string) ([]common.RechargeRecord, error) {
	var records []common.RechargeRecord
	path := "/api/recharges?employee_id=" + url.QueryEscape(employeeID)
	if err := c.do(http.MethodGet, path, nil, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (c *Client) CreateDish(req common.CreateDishRequest) (*common.Dish, error) {
	var dish common.Dish
	if err := c.do(http.MethodPost, "/api/dishes", req, &dish); err != nil {
		return nil, err
	}
	return &dish, nil
}

func (c *Client) ListDishes(onlyOnSale bool) ([]common.Dish, error) {
	var dishes []common.Dish
	path := "/api/dishes"
	if onlyOnSale {
		path += "?only_on_sale=true"
	}
	if err := c.do(http.MethodGet, path, nil, &dishes); err != nil {
		return nil, err
	}
	return dishes, nil
}

func (c *Client) GetDish(id string) (*common.Dish, error) {
	var dish common.Dish
	if err := c.do(http.MethodGet, "/api/dishes/"+id, nil, &dish); err != nil {
		return nil, err
	}
	return &dish, nil
}

func (c *Client) UpdateDish(id string, req common.UpdateDishRequest) (*common.Dish, error) {
	var dish common.Dish
	if err := c.do(http.MethodPut, "/api/dishes/"+id, req, &dish); err != nil {
		return nil, err
	}
	return &dish, nil
}

func (c *Client) SetDishOnSale(id string, onSale bool) error {
	path := fmt.Sprintf("/api/dishes/%s/onsale?on_sale=%s", id, strconv.FormatBool(onSale))
	return c.do(http.MethodPost, path, nil, nil)
}

func (c *Client) Checkout(req common.CheckoutRequest) (*common.ConsumptionRecord, error) {
	var record common.ConsumptionRecord
	if err := c.do(http.MethodPost, "/api/checkout", req, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (c *Client) ListConsumptionRecords(employeeID string) ([]common.ConsumptionRecord, error) {
	var records []common.ConsumptionRecord
	path := "/api/consumptions?employee_id=" + url.QueryEscape(employeeID)
	if err := c.do(http.MethodGet, path, nil, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (c *Client) GetNutritionSummary(employeeID string, days int) (*common.NutritionSummary, error) {
	var summary common.NutritionSummary
	path := fmt.Sprintf("/api/nutrition?employee_id=%s&days=%d", url.QueryEscape(employeeID), days)
	if err := c.do(http.MethodGet, path, nil, &summary); err != nil {
		return nil, err
	}
	return &summary, nil
}
