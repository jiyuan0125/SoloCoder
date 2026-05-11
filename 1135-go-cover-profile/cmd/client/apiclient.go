package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/solocoder/coverage-analyzer/pkg/api"
)

type APIClient struct {
	BaseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{BaseURL: baseURL}
}

func (c *APIClient) UploadFile(filePath string) (*api.UploadResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", c.BaseURL+"/upload", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp)
	}

	var uploadResp api.UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &uploadResp, nil
}

func (c *APIClient) GetFunctionCoverage(id string) (*api.FunctionCoverageReport, error) {
	var report api.FunctionCoverageReport
	err := c.getJSON("/coverage/function?id="+id, &report)
	return &report, err
}

func (c *APIClient) GetPackageCoverage(id string) (*api.PackageCoverageReport, error) {
	var report api.PackageCoverageReport
	err := c.getJSON("/coverage/package?id="+id, &report)
	return &report, err
}

func (c *APIClient) GetLineCoverage(id string) (*api.LineCoverageReport, error) {
	var report api.LineCoverageReport
	err := c.getJSON("/coverage/line?id="+id, &report)
	return &report, err
}

func (c *APIClient) GetSummary(id string) (*api.SummaryResponse, error) {
	var summary api.SummaryResponse
	err := c.getJSON("/summary?id="+id, &summary)
	return &summary, err
}

func (c *APIClient) GetDiff(oldID, newID string) (*api.DiffReport, error) {
	var diff api.DiffReport
	err := c.getJSON("/diff?old="+oldID+"&new="+newID, &diff)
	return &diff, err
}

func (c *APIClient) getJSON(path string, result interface{}) error {
	resp, err := http.Get(c.BaseURL + path)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return readError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func readError(resp *http.Response) error {
	var errResp api.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}

	if errResp.Line > 0 {
		return fmt.Errorf("error at line %d: %s", errResp.Line, errResp.Error)
	}
	return fmt.Errorf("error: %s", errResp.Error)
}
