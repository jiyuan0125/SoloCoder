package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"hash-checkpoint/common"
)

type APIClient struct {
	baseURL string
	http    *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return &APIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *APIClient) postJSON(path string, req interface{}, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpResp, err := c.http.Post(c.baseURL+path, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	if httpResp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil {
			return fmt.Errorf("api error: %s", errResp.Error)
		}
		return fmt.Errorf("api error: http %d", httpResp.StatusCode)
	}

	return json.Unmarshal(respBody, resp)
}

func (c *APIClient) StartCheck(
	filePath string,
	modTime time.Time,
	fileSize int64,
	algorithm common.HashAlgorithm,
	chunkSize int64,
) (*common.StartCheckResponse, error) {
	req := common.StartCheckRequest{
		FilePath:  filePath,
		ModTime:   modTime,
		FileSize:  fileSize,
		Algorithm: algorithm,
		ChunkSize: chunkSize,
	}
	var resp common.StartCheckResponse
	err := c.postJSON("/check/start", &req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) SubmitChunk(
	filePath string,
	modTime time.Time,
	chunkIndex int,
	chunkHash string,
	chunkLen int64,
) (*common.SubmitChunkResponse, error) {
	req := common.SubmitChunkRequest{
		FilePath:   filePath,
		ModTime:    modTime,
		ChunkIndex: chunkIndex,
		ChunkHash:  chunkHash,
		ChunkLen:   chunkLen,
	}
	var resp common.SubmitChunkResponse
	err := c.postJSON("/check/chunk", &req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) GetProgress(
	filePath string,
	modTime time.Time,
) (*common.GetProgressResponse, error) {
	query := url.Values{}
	query.Set("file_path", filePath)
	query.Set("mod_time", modTime.UTC().Format(time.RFC3339Nano))

	httpResp, err := c.http.Get(c.baseURL + "/check/progress?" + query.Encode())
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil {
			return nil, fmt.Errorf("api error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("api error: http %d", httpResp.StatusCode)
	}

	var resp common.GetProgressResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) CompleteCheck(
	filePath string,
	modTime time.Time,
) (*common.CompleteCheckResponse, error) {
	req := common.CompleteCheckRequest{
		FilePath: filePath,
		ModTime:  modTime,
	}
	var resp common.CompleteCheckResponse
	err := c.postJSON("/check/complete", &req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
