package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"repair-platform/common"
)

type Client struct {
	baseURL string
}

func newClient() *Client {
	base := os.Getenv("SERVER_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	return &Client{baseURL: base}
}

func (c *Client) do(method, path string, body interface{}, out interface{}) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *Client) CreateRepair(userID, address, area string, appliance common.ApplianceInfo, faultDesc string) (*common.CreateRepairResponse, error) {
	req := common.CreateRepairRequest{
		UserID:    userID,
		Address:   address,
		Area:      area,
		Appliance: appliance,
		FaultDesc: faultDesc,
	}
	var out common.CreateRepairResponse
	err := c.do(http.MethodPost, "/api/repairs", req, &out)
	return &out, err
}

func (c *Client) ListRepairs() (*common.ListRepairRequestsResponse, error) {
	var out common.ListRepairRequestsResponse
	err := c.do(http.MethodGet, "/api/repairs/list", nil, &out)
	return &out, err
}

func (c *Client) AssignTech(reqID string) (*common.AssignTechnicianResponse, error) {
	req := common.AssignTechnicianRequest{RequestID: reqID}
	var out common.AssignTechnicianResponse
	err := c.do(http.MethodPost, "/api/repairs/assign", req, &out)
	return &out, err
}

func (c *Client) ApplySpare(reqID, techID string, items []common.ApplySpareItem) (*common.ApplySpareResponse, error) {
	req := common.ApplySpareRequest{RequestID: reqID, TechID: techID, SpareParts: items}
	var out common.ApplySpareResponse
	err := c.do(http.MethodPost, "/api/repairs/apply-spare", req, &out)
	return &out, err
}

func (c *Client) CompleteRepair(reqID, techID, faultCause string, used []common.UsedPartItem, ret []common.ReturnPartItem) (*common.CompleteRepairResponse, error) {
	req := common.CompleteRepairRequest{
		RequestID:   reqID,
		TechID:      techID,
		FaultCause:  faultCause,
		PartsUsed:   used,
		PartsReturn: ret,
	}
	var out common.CompleteRepairResponse
	err := c.do(http.MethodPost, "/api/repairs/complete", req, &out)
	return &out, err
}

func (c *Client) AddSpare(part common.SparePart) error {
	req := common.AddSparePartRequest{Part: part}
	return c.do(http.MethodPost, "/api/spares", req, nil)
}

func (c *Client) ListSpares() (*common.GetSparePartsResponse, error) {
	var out common.GetSparePartsResponse
	err := c.do(http.MethodGet, "/api/spares/list", nil, &out)
	return &out, err
}

func (c *Client) UpdateSparePrice(code string, newPrice float64) error {
	req := common.UpdateSparePriceRequest{Code: code, NewPrice: newPrice}
	return c.do(http.MethodPost, "/api/spares/update-price", req, nil)
}

func (c *Client) ListPOs() (*common.GetPurchaseOrdersResponse, error) {
	var out common.GetPurchaseOrdersResponse
	err := c.do(http.MethodGet, "/api/spares/purchase-orders", nil, &out)
	return &out, err
}

func (c *Client) AddTech(tech common.Technician) error {
	req := common.AddTechnicianRequest{Tech: tech}
	return c.do(http.MethodPost, "/api/techs", req, nil)
}

func (c *Client) ListTechs() (*common.GetTechniciansResponse, error) {
	var out common.GetTechniciansResponse
	err := c.do(http.MethodGet, "/api/techs/list", nil, &out)
	return &out, err
}
