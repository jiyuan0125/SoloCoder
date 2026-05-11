package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"marketplace/common"
	"net/http"
	"net/url"
)

type APIClient struct {
	baseURL string
	userID  string
}

func NewAPIClient(baseURL, userID string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		userID:  userID,
	}
}

func (c *APIClient) SetUserID(userID string) {
	c.userID = userID
}

func (c *APIClient) doRequest(method, path string, body interface{}, queryParams url.Values) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	reqURL := c.baseURL + path
	if queryParams != nil && len(queryParams) > 0 {
		reqURL += "?" + queryParams.Encode()
	}

	req, err := http.NewRequest(method, reqURL, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.userID != "" {
		req.Header.Set("X-User-ID", c.userID)
	}

	return http.DefaultClient.Do(req)
}

func (c *APIClient) Register(username string, role common.UserRole) (*common.User, error) {
	reqBody := common.RegisterUserRequest{
		Username: username,
		Role:     role,
	}

	resp, err := c.doRequest(http.MethodPost, "/users/register", reqBody, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}

	return result.Data, nil
}

func (c *APIClient) CreateItem(req common.CreateItemRequest) (*common.Item, error) {
	resp, err := c.doRequest(http.MethodPost, "/items", req, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.ItemResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}

	return result.Data, nil
}

func (c *APIClient) ApproveItem(itemID string, approved bool, reason string) (*common.Item, error) {
	reqBody := common.ApproveItemRequest{
		Approved: approved,
		Reason:   reason,
	}

	resp, err := c.doRequest(http.MethodPost, "/items/"+itemID+"/approve", reqBody, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.ItemResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}

	return result.Data, nil
}

func (c *APIClient) SearchItems(req common.SearchItemsRequest) ([]common.Item, error) {
	query := url.Values{}
	if req.Keyword != "" {
		query.Set("keyword", req.Keyword)
	}
	if req.Category != "" {
		query.Set("category", string(req.Category))
	}
	if req.Condition != "" {
		query.Set("condition", string(req.Condition))
	}
	if req.MinPrice > 0 {
		query.Set("min_price", fmt.Sprintf("%d", req.MinPrice))
	}
	if req.MaxPrice > 0 {
		query.Set("max_price", fmt.Sprintf("%d", req.MaxPrice))
	}

	resp, err := c.doRequest(http.MethodGet, "/items", nil, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.ItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}

	return result.Data, nil
}

func (c *APIClient) GetItem(itemID string) (*common.Item, error) {
	resp, err := c.doRequest(http.MethodGet, "/items/"+itemID, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.ItemResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}

	return result.Data, nil
}

func (c *APIClient) RemoveItem(itemID string) error {
	resp, err := c.doRequest(http.MethodDelete, "/items/"+itemID, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result common.BaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf(result.Message)
	}

	return nil
}

func (c *APIClient) MakeOffer(itemID string, price int64) (*common.Negotiation, error) {
	reqBody := common.MakeOfferRequest{
		Price: price,
	}

	resp, err := c.doRequest(http.MethodPost, "/items/"+itemID+"/offer", reqBody, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.NegotiationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}

	return result.Data, nil
}

func (c *APIClient) SellerRespond(negotiationID string, action string, counterPrice int64) (*common.Negotiation, error) {
	reqBody := common.RespondOfferRequest{
		Action:       action,
		CounterPrice: counterPrice,
	}

	resp, err := c.doRequest(http.MethodPost, "/negotiations/"+negotiationID+"/seller-respond", reqBody, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.NegotiationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}

	return result.Data, nil
}

func (c *APIClient) BuyerRespond(negotiationID string, action string, counterPrice int64) (*common.Negotiation, error) {
	reqBody := common.RespondOfferRequest{
		Action:       action,
		CounterPrice: counterPrice,
	}

	resp, err := c.doRequest(http.MethodPost, "/negotiations/"+negotiationID+"/buyer-respond", reqBody, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.NegotiationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}

	return result.Data, nil
}

func (c *APIClient) PayOrder(orderID string) (*common.Order, error) {
	resp, err := c.doRequest(http.MethodPost, "/orders/"+orderID+"/pay", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}

	return result.Data, nil
}

func (c *APIClient) ShipOrder(orderID, logisticsNo string) (*common.Order, error) {
	reqBody := common.ShipOrderRequest{
		LogisticsNo: logisticsNo,
	}

	resp, err := c.doRequest(http.MethodPost, "/orders/"+orderID+"/ship", reqBody, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}

	return result.Data, nil
}

func (c *APIClient) ConfirmReceiveOrder(orderID string) (*common.Order, error) {
	resp, err := c.doRequest(http.MethodPost, "/orders/"+orderID+"/confirm", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}

	return result.Data, nil
}

func (c *APIClient) GetOrder(orderID string) (*common.Order, error) {
	resp, err := c.doRequest(http.MethodGet, "/orders/"+orderID, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}

	return result.Data, nil
}
