package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"parking-system/common"
	"net/url"
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
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return err
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var apiResp common.Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Code != http.StatusOK {
		return fmt.Errorf("api error: %s", apiResp.Message)
	}

	if result != nil && apiResp.Data != nil {
		dataBytes, err := json.Marshal(apiResp.Data)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(dataBytes, result); err != nil {
			return err
		}
	}

	return nil
}

func (c *APIClient) CreateSpot(req common.CreateParkingSpotRequest) (*common.ParkingSpotDTO, error) {
	var spot common.ParkingSpotDTO
	if err := c.doRequest(http.MethodPost, "/api/spots", req, &spot); err != nil {
		return nil, err
	}
	return &spot, nil
}

func (c *APIClient) ListSpots(area, status string) ([]common.ParkingSpotDTO, error) {
	query := url.Values{}
	if area != "" {
		query.Set("area", area)
	}
	if status != "" {
		query.Set("status", status)
	}
	path := "/api/spots"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var spots []common.ParkingSpotDTO
	if err := c.doRequest(http.MethodGet, path, nil, &spots); err != nil {
		return nil, err
	}
	return spots, nil
}

func (c *APIClient) UpdateSpotStatus(id string, status common.ParkingSpotStatus) (*common.ParkingSpotDTO, error) {
	req := common.UpdateParkingSpotStatusRequest{Status: status}
	var spot common.ParkingSpotDTO
	if err := c.doRequest(http.MethodPut, "/api/spots/"+id+"/status", req, &spot); err != nil {
		return nil, err
	}
	return &spot, nil
}

func (c *APIClient) GetGuidance() (*common.ParkingGuidanceDTO, error) {
	var guidance common.ParkingGuidanceDTO
	if err := c.doRequest(http.MethodGet, "/api/guidance", nil, &guidance); err != nil {
		return nil, err
	}
	return &guidance, nil
}

func (c *APIClient) CheckIn(plateNumber string) (*common.CheckInResponse, error) {
	req := common.CheckInRequest{PlateNumber: plateNumber}
	var resp common.CheckInResponse
	if err := c.doRequest(http.MethodPost, "/api/parking/checkin", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) CheckOut(plateNumber string) (*common.CheckOutResponse, error) {
	req := common.CheckOutRequest{PlateNumber: plateNumber}
	var resp common.CheckOutResponse
	if err := c.doRequest(http.MethodPost, "/api/parking/checkout", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) QueryFee(plateNumber string) (*common.QueryFeeResponse, error) {
	query := url.Values{}
	query.Set("plate_number", plateNumber)
	path := "/api/parking/fee?" + query.Encode()

	var resp common.QueryFeeResponse
	if err := c.doRequest(http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) CreateMonthlyCard(req common.CreateMonthlyCardRequest) (*common.MonthlyCardDTO, error) {
	var card common.MonthlyCardDTO
	if err := c.doRequest(http.MethodPost, "/api/cards", req, &card); err != nil {
		return nil, err
	}
	return &card, nil
}

func (c *APIClient) RenewMonthlyCard(cardID string) (*common.MonthlyCardDTO, error) {
	req := common.RenewMonthlyCardRequest{CardID: cardID}
	var card common.MonthlyCardDTO
	if err := c.doRequest(http.MethodPost, "/api/cards/renew", req, &card); err != nil {
		return nil, err
	}
	return &card, nil
}

func (c *APIClient) ListMonthlyCards(plateNumber string, activeOnly bool) ([]common.MonthlyCardDTO, error) {
	query := url.Values{}
	if plateNumber != "" {
		query.Set("plate_number", plateNumber)
	}
	if activeOnly {
		query.Set("active_only", "true")
	}
	path := "/api/cards"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var cards []common.MonthlyCardDTO
	if err := c.doRequest(http.MethodGet, path, nil, &cards); err != nil {
		return nil, err
	}
	return cards, nil
}

func (c *APIClient) ListParkingRecords(plateNumber string, activeOnly bool) ([]common.ParkingRecordDTO, error) {
	query := url.Values{}
	if plateNumber != "" {
		query.Set("plate_number", plateNumber)
	}
	if activeOnly {
		query.Set("active_only", "true")
	}
	path := "/api/parking/records"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var records []common.ParkingRecordDTO
	if err := c.doRequest(http.MethodGet, path, nil, &records); err != nil {
		return nil, err
	}
	return records, nil
}
