package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hospital-bed/shared/protocol"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type APIClient struct {
	baseURL string
}

func NewAPIClient(serverAddr string) *APIClient {
	baseURL := strings.TrimSuffix(serverAddr, "/")
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	return &APIClient{baseURL: baseURL}
}

func (c *APIClient) doRequest(method, path string, body interface{}, respObj interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("序列化请求失败: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp protocol.Response
		json.Unmarshal(respData, &errResp)
		if errResp.Message != "" {
			return fmt.Errorf("错误: %s", errResp.Message)
		}
		return fmt.Errorf("服务器返回状态码: %d", resp.StatusCode)
	}

	if respObj != nil {
		if err := json.Unmarshal(respData, respObj); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}
	}

	return nil
}

func (c *APIClient) ConfigureDepartment(req protocol.DepartmentConfigRequest) error {
	return c.doRequest(http.MethodPost, "/api/configure", req, nil)
}

func (c *APIClient) AdmitPatient(req protocol.AdmissionRequest) (*protocol.AdmissionResponse, error) {
	var resp protocol.AdmissionResponse
	if err := c.doRequest(http.MethodPost, "/api/admit", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) DischargePatient(req protocol.DischargeRequest) (*protocol.DischargeResponse, error) {
	var resp protocol.DischargeResponse
	if err := c.doRequest(http.MethodPost, "/api/discharge", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) GetDepartmentStatus(deptName string) (*protocol.DepartmentStatusResponse, error) {
	path := fmt.Sprintf("/api/department/status?department=%s", url.QueryEscape(deptName))
	var resp protocol.DepartmentStatusResponse
	if err := c.doRequest(http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) GetPatientBed(patientID string) (*protocol.PatientBedResponse, error) {
	path := fmt.Sprintf("/api/patient/bed?patient_id=%s", url.QueryEscape(patientID))
	var resp protocol.PatientBedResponse
	if err := c.doRequest(http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) GetQueueStatus(deptName string) (*protocol.QueueStatusResponse, error) {
	path := fmt.Sprintf("/api/queue/status?department=%s", url.QueryEscape(deptName))
	var resp protocol.QueueStatusResponse
	if err := c.doRequest(http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

type TransferRequest struct {
	PatientID     string          `json:"patient_id"`
	TargetDept    string          `json:"target_department"`
	TargetWard    *string         `json:"target_ward,omitempty"`
	PreferredType protocol.BedType `json:"preferred_type"`
}

func (c *APIClient) TransferPatient(req TransferRequest) error {
	return c.doRequest(http.MethodPost, "/api/transfer", req, nil)
}
