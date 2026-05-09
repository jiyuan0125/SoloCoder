package leaf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"leaf-segment/pkg/common"
)

type CenterClient interface {
	GetSegment() (*Segment, error)
	ReportProgress(current, max int64) error
}

type HTTPCenterClient struct {
	config      *Config
	httpClient  *http.Client
	businessKey string
}

func NewHTTPCenterClient(config *Config) *HTTPCenterClient {
	return &HTTPCenterClient{
		config:      config,
		httpClient:  &http.Client{Timeout: config.CenterTimeout},
		businessKey: config.BusinessKey,
	}
}

func (c *HTTPCenterClient) GetSegment() (*Segment, error) {
	reqBody := common.CenterGetSegmentRequest{
		BusinessKey: c.businessKey,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/segment/get", c.config.CenterURL)
	var lastErr error
	for i := 0; i <= c.config.MaxRetries; i++ {
		seg, err := c.doGetSegment(url, bodyBytes)
		if err == nil {
			return seg, nil
		}
		lastErr = err
		if i < c.config.MaxRetries {
			time.Sleep(c.config.RetryInterval)
		}
	}
	return nil, lastErr
}

func (c *HTTPCenterClient) doGetSegment(url string, bodyBytes []byte) (*Segment, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("center returned status %d", resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var centerResp common.CenterGetSegmentResponse
	if err := json.Unmarshal(respBody, &centerResp); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	if !centerResp.Success {
		return nil, fmt.Errorf("center error: %s", centerResp.Msg)
	}

	return NewSegment(centerResp.Start, centerResp.End), nil
}

func (c *HTTPCenterClient) ReportProgress(current, max int64) error {
	reqBody := common.CenterReportProgressRequest{
		BusinessKey: c.businessKey,
		Current:     current,
		Max:         max,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/segment/report", c.config.CenterURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, _ = ioutil.ReadAll(resp.Body)
	return nil
}
