package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"supplier-portal/common"
)

type APIClient struct {
	BaseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (c *APIClient) request(method, path string, body, result interface{}) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reader)
	if err != nil {
		return err
	}

	if reader != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		if len(respBody) > 0 {
			json.Unmarshal(respBody, &errResp)
		}
		if errResp.Error != "" {
			return fmt.Errorf("API error [%d]: %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("API error [%d]: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	if result != nil && len(respBody) > 0 {
		return json.Unmarshal(respBody, result)
	}

	return nil
}

func (c *APIClient) CreateInquiry(req common.CreateInquiryRequest) (*common.InquiryResponse, error) {
	var result common.InquiryResponse
	err := c.request("POST", "/api/inquiries", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) GetInquiry(id string) (*common.InquiryResponse, error) {
	var result common.InquiryResponse
	err := c.request("GET", "/api/inquiry?id="+url.QueryEscape(id), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) ListInquiries() ([]common.InquiryResponse, error) {
	var result []common.InquiryResponse
	err := c.request("GET", "/api/inquiries", nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *APIClient) UpdateInquiryMaterials(req common.UpdateInquiryMaterialsRequest) (*common.InquiryResponse, error) {
	var result common.InquiryResponse
	err := c.request("PUT", "/api/inquiry/materials", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) GetTodoReminders(supplierID string) ([]common.TodoReminderResponse, error) {
	var result []common.TodoReminderResponse
	err := c.request("GET", "/api/todos?supplier_id="+url.QueryEscape(supplierID), nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *APIClient) SubmitQuotation(req common.SubmitQuotationRequest) (*common.QuotationResponse, error) {
	var result common.QuotationResponse
	err := c.request("POST", "/api/quotations", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) GetQuotations(inquiryID string) ([]common.QuotationResponse, error) {
	var result []common.QuotationResponse
	err := c.request("GET", "/api/quotations?inquiry_id="+url.QueryEscape(inquiryID), nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *APIClient) GenerateComparison(inquiryID string) (*common.ComparisonResponse, error) {
	var result common.ComparisonResponse
	err := c.request("POST", "/api/comparison", common.ComparisonRequest{InquiryID: inquiryID}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) AwardAndCreateOrders(req common.AwardRequest) ([]common.PurchaseOrderResponse, error) {
	var result []common.PurchaseOrderResponse
	err := c.request("POST", "/api/award", req, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *APIClient) GetOrdersByInquiry(inquiryID string) ([]common.PurchaseOrderResponse, error) {
	var result []common.PurchaseOrderResponse
	err := c.request("GET", "/api/orders?inquiry_id="+url.QueryEscape(inquiryID), nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *APIClient) GetOrdersBySupplier(supplierID string) ([]common.PurchaseOrderResponse, error) {
	var result []common.PurchaseOrderResponse
	err := c.request("GET", "/api/orders?supplier_id="+url.QueryEscape(supplierID), nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
