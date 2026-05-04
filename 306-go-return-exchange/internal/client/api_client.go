package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"return-exchange/pkg/common"
	"return-exchange/pkg/models"
)

type APIClient struct {
	baseURL string
	client  *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *APIClient) doRequest(method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *APIClient) CreateMockOrder() (string, error) {
	var result map[string]interface{}
	err := c.doRequest("POST", "/api/orders/mock", nil, &result)
	if err != nil {
		return "", err
	}
	if success, ok := result["success"].(bool); !ok || !success {
		msg, _ := result["message"].(string)
		return "", fmt.Errorf("failed to create mock order: %s", msg)
	}
	orderID, _ := result["order_id"].(string)
	return orderID, nil
}

func (c *APIClient) SubmitReturn(req common.SubmitReturnRequest) (string, error) {
	var result common.SubmitResponse
	err := c.doRequest("POST", "/api/applications/return", req, &result)
	if err != nil {
		return "", err
	}
	if !result.Success {
		return "", fmt.Errorf(result.Message)
	}
	return result.ApplicationID, nil
}

func (c *APIClient) SubmitExchange(req common.SubmitExchangeRequest) (string, error) {
	var result common.SubmitResponse
	err := c.doRequest("POST", "/api/applications/exchange", req, &result)
	if err != nil {
		return "", err
	}
	if !result.Success {
		return "", fmt.Errorf(result.Message)
	}
	return result.ApplicationID, nil
}

func (c *APIClient) ReviewApplication(req common.ReviewRequest) error {
	var result common.CommonResponse
	err := c.doRequest("POST", "/api/applications/review", req, &result)
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf(result.Message)
	}
	return nil
}

func (c *APIClient) ProcessRefund(req common.ProcessRefundRequest) (string, float64, error) {
	var result common.ProcessRefundResponse
	err := c.doRequest("POST", "/api/applications/refund", req, &result)
	if err != nil {
		return "", 0, err
	}
	if !result.Success {
		return "", 0, fmt.Errorf(result.Message)
	}
	return result.RefundID, result.TotalRefund, nil
}

func (c *APIClient) HandlePriceDifference(req common.HandlePriceDifferenceRequest) (*common.HandlePriceDifferenceResponse, error) {
	var result common.HandlePriceDifferenceResponse
	err := c.doRequest("POST", "/api/applications/price-difference", req, &result)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}
	return &result, nil
}

func (c *APIClient) CreateShippingOrder(req common.CreateShippingOrderRequest) (string, error) {
	var result common.CreateShippingOrderResponse
	err := c.doRequest("POST", "/api/applications/shipping", req, &result)
	if err != nil {
		return "", err
	}
	if !result.Success {
		return "", fmt.Errorf(result.Message)
	}
	return result.ShippingOrderID, nil
}

func (c *APIClient) CompleteApplication(applicationID string) error {
	var result common.CommonResponse
	err := c.doRequest("POST", "/api/applications/"+applicationID+"/complete", nil, &result)
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf(result.Message)
	}
	return nil
}

func (c *APIClient) GetApplication(applicationID string) (*models.ReturnExchangeApplication, error) {
	var result common.ApplicationDetailResponse
	err := c.doRequest("GET", "/api/applications/"+applicationID, nil, &result)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}
	return result.Application, nil
}

func (c *APIClient) ListAllApplications() ([]models.ReturnExchangeApplication, error) {
	var result common.ApplicationListResponse
	err := c.doRequest("GET", "/api/applications", nil, &result)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}
	return result.Applications, nil
}

func (c *APIClient) ListUserApplications(userID string) ([]models.ReturnExchangeApplication, error) {
	var result common.ApplicationListResponse
	err := c.doRequest("GET", "/api/users/"+userID+"/applications", nil, &result)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}
	return result.Applications, nil
}
