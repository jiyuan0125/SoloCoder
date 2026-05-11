package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"quality-trace/pkg/common"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *Client) request(method, path string, body interface{}, queryParams map[string]string) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonData)
	}

	u := c.baseURL + path
	if len(queryParams) > 0 {
		q := url.Values{}
		for k, v := range queryParams {
			q.Add(k, v)
		}
		u += "?" + q.Encode()
	}

	req, err := http.NewRequest(method, u, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

func (c *Client) CreateBatch(req common.CreateBatchRequest) (*common.Batch, error) {
	data, err := c.request(http.MethodPost, "/api/batches", req, nil)
	if err != nil {
		return nil, err
	}

	var resp common.Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	batchJSON, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var batch common.Batch
	if err := json.Unmarshal(batchJSON, &batch); err != nil {
		return nil, err
	}

	return &batch, nil
}

func (c *Client) GetBatch(batchID string) (*common.Batch, error) {
	data, err := c.request(http.MethodGet, "/api/batches/get", nil, map[string]string{
		"batch_id": batchID,
	})
	if err != nil {
		return nil, err
	}

	var resp common.Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	batchJSON, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var batch common.Batch
	if err := json.Unmarshal(batchJSON, &batch); err != nil {
		return nil, err
	}

	return &batch, nil
}

func (c *Client) CompleteBatch(req common.CompleteBatchRequest) (*common.Batch, error) {
	data, err := c.request(http.MethodPost, "/api/batches/complete", req, nil)
	if err != nil {
		return nil, err
	}

	var resp common.Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	batchJSON, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var batch common.Batch
	if err := json.Unmarshal(batchJSON, &batch); err != nil {
		return nil, err
	}

	return &batch, nil
}

func (c *Client) ListBatches(status, productName string) ([]*common.Batch, error) {
	queryParams := map[string]string{}
	if status != "" {
		queryParams["status"] = status
	}
	if productName != "" {
		queryParams["product_name"] = productName
	}

	data, err := c.request(http.MethodGet, "/api/batches", nil, queryParams)
	if err != nil {
		return nil, err
	}

	var resp common.Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	batchesJSON, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var batches []*common.Batch
	if err := json.Unmarshal(batchesJSON, &batches); err != nil {
		return nil, err
	}

	return batches, nil
}

func (c *Client) AddProcessFlow(req common.AddProcessFlowRequest) (*common.ProcessFlow, error) {
	data, err := c.request(http.MethodPost, "/api/process-flows", req, nil)
	if err != nil {
		return nil, err
	}

	var resp common.Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	flowJSON, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var flow common.ProcessFlow
	if err := json.Unmarshal(flowJSON, &flow); err != nil {
		return nil, err
	}

	return &flow, nil
}

func (c *Client) StartProcess(req common.StartProcessRequest) (*common.ProcessRecord, error) {
	data, err := c.request(http.MethodPost, "/api/processes/start", req, nil)
	if err != nil {
		return nil, err
	}

	var resp common.Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	recordJSON, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var record common.ProcessRecord
	if err := json.Unmarshal(recordJSON, &record); err != nil {
		return nil, err
	}

	return &record, nil
}

func (c *Client) CompleteProcess(req common.CompleteProcessRequest) (*common.ProcessRecord, error) {
	data, err := c.request(http.MethodPost, "/api/processes/complete", req, nil)
	if err != nil {
		return nil, err
	}

	var resp common.Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	recordJSON, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var record common.ProcessRecord
	if err := json.Unmarshal(recordJSON, &record); err != nil {
		return nil, err
	}

	return &record, nil
}

func (c *Client) MakeReworkDecision(req common.ReworkDecisionRequest) (*common.Batch, error) {
	data, err := c.request(http.MethodPost, "/api/processes/decision", req, nil)
	if err != nil {
		return nil, err
	}

	var resp common.Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	batchJSON, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var batch common.Batch
	if err := json.Unmarshal(batchJSON, &batch); err != nil {
		return nil, err
	}

	return &batch, nil
}

func (c *Client) ListProcessRecords(batchID string) ([]*common.ProcessRecord, error) {
	queryParams := map[string]string{}
	if batchID != "" {
		queryParams["batch_id"] = batchID
	}

	data, err := c.request(http.MethodGet, "/api/processes", nil, queryParams)
	if err != nil {
		return nil, err
	}

	var resp common.Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	recordsJSON, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var records []*common.ProcessRecord
	if err := json.Unmarshal(recordsJSON, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (c *Client) AddInspectionSpec(req common.AddInspectionSpecRequest) (*common.InspectionSpec, error) {
	data, err := c.request(http.MethodPost, "/api/inspection-specs", req, nil)
	if err != nil {
		return nil, err
	}

	var resp common.Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	specJSON, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var spec common.InspectionSpec
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

func (c *Client) AddInspectionRecord(req common.AddInspectionRequest) (*common.InspectionRecord, error) {
	data, err := c.request(http.MethodPost, "/api/inspections", req, nil)
	if err != nil {
		return nil, err
	}

	var resp common.Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	recordJSON, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var record common.InspectionRecord
	if err := json.Unmarshal(recordJSON, &record); err != nil {
		return nil, err
	}

	return &record, nil
}

func (c *Client) ListInspectionRecords(batchID string, inspectionType string) ([]*common.InspectionRecord, error) {
	queryParams := map[string]string{}
	if batchID != "" {
		queryParams["batch_id"] = batchID
	}
	if inspectionType != "" {
		queryParams["inspection_type"] = inspectionType
	}

	data, err := c.request(http.MethodGet, "/api/inspections", nil, queryParams)
	if err != nil {
		return nil, err
	}

	var resp common.Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	recordsJSON, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var records []*common.InspectionRecord
	if err := json.Unmarshal(recordsJSON, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (c *Client) GetMetrics(days int) (*common.MetricsSummaryResponse, error) {
	queryParams := map[string]string{}
	if days > 0 {
		queryParams["days"] = fmt.Sprintf("%d", days)
	}

	data, err := c.request(http.MethodGet, "/api/metrics", nil, queryParams)
	if err != nil {
		return nil, err
	}

	var resp common.Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	metricsJSON, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var metrics common.MetricsSummaryResponse
	if err := json.Unmarshal(metricsJSON, &metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}

func (c *Client) ExportBatches(status, productName, outputFile string) error {
	queryParams := map[string]string{}
	if status != "" {
		queryParams["status"] = status
	}
	if productName != "" {
		queryParams["product_name"] = productName
	}

	data, err := c.request(http.MethodGet, "/api/export/batches", nil, queryParams)
	if err != nil {
		return err
	}

	return os.WriteFile(outputFile, data, 0644)
}

func (c *Client) ExportProcessRecords(batchID, outputFile string) error {
	queryParams := map[string]string{}
	if batchID != "" {
		queryParams["batch_id"] = batchID
	}

	data, err := c.request(http.MethodGet, "/api/export/processes", nil, queryParams)
	if err != nil {
		return err
	}

	return os.WriteFile(outputFile, data, 0644)
}

func (c *Client) ExportInspectionRecords(batchID, inspectionType, outputFile string) error {
	queryParams := map[string]string{}
	if batchID != "" {
		queryParams["batch_id"] = batchID
	}
	if inspectionType != "" {
		queryParams["inspection_type"] = inspectionType
	}

	data, err := c.request(http.MethodGet, "/api/export/inspections", nil, queryParams)
	if err != nil {
		return err
	}

	return os.WriteFile(outputFile, data, 0644)
}

func (c *Client) Health() (string, error) {
	data, err := c.request(http.MethodGet, "/health", nil, nil)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
