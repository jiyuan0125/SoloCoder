package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ftp-client/common"
)

type APIClient struct {
	serverURL string
}

func NewAPIClient(serverURL string) *APIClient {
	return &APIClient{serverURL: serverURL}
}

func (c *APIClient) postJSON(path string, req interface{}, resp interface{}) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpResp, err := http.Post(c.serverURL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	if resp != nil {
		return json.Unmarshal(body, resp)
	}
	return nil
}

func (c *APIClient) getJSON(path string, resp interface{}) error {
	httpResp, err := http.Get(c.serverURL + path)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, resp)
}

func (c *APIClient) deleteJSON(path string, req interface{}, resp interface{}) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest(http.MethodDelete, c.serverURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	if resp != nil {
		return json.Unmarshal(body, resp)
	}
	return nil
}

func (c *APIClient) TestConnection(config common.FTPConfig) error {
	var resp common.SimpleResponse
	if err := c.postJSON("/api/config", config, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *APIClient) ListDir(config common.FTPConfig, path string, useMLSD bool) (*common.ListDirResponse, error) {
	req := common.FileOperationRequest{
		Config:     config,
		Operation:  common.OpListDir,
		RemotePath: path,
		UseMLSD:    useMLSD,
	}
	var resp common.ListDirResponse
	if err := c.postJSON("/api/operation", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) ChangeDir(config common.FTPConfig, path string) error {
	req := common.FileOperationRequest{
		Config:     config,
		Operation:  common.OpChangeDir,
		RemotePath: path,
	}
	var resp common.SimpleResponse
	if err := c.postJSON("/api/operation", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *APIClient) MakeDir(config common.FTPConfig, path string) error {
	req := common.FileOperationRequest{
		Config:     config,
		Operation:  common.OpMakeDir,
		RemotePath: path,
	}
	var resp common.SimpleResponse
	if err := c.postJSON("/api/operation", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *APIClient) RemoveDir(config common.FTPConfig, path string) error {
	req := common.FileOperationRequest{
		Config:     config,
		Operation:  common.OpRemoveDir,
		RemotePath: path,
	}
	var resp common.SimpleResponse
	if err := c.postJSON("/api/operation", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *APIClient) GetCurrentDir(config common.FTPConfig) (*common.ListDirResponse, error) {
	req := common.FileOperationRequest{
		Config:    config,
		Operation: common.OpGetCurrentDir,
	}
	var resp common.ListDirResponse
	if err := c.postJSON("/api/operation", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) Download(config common.FTPConfig, remotePath, localPath string, recursive bool) (string, error) {
	req := common.FileOperationRequest{
		Config:     config,
		Operation:  common.OpDownload,
		RemotePath: remotePath,
		LocalPath:  localPath,
		Recursive:  recursive,
	}
	var resp struct {
		Success    bool   `json:"success"`
		TransferID string `json:"transfer_id"`
		Error      string `json:"error"`
	}
	if err := c.postJSON("/api/operation", req, &resp); err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf("%s", resp.Error)
	}
	return resp.TransferID, nil
}

func (c *APIClient) Upload(config common.FTPConfig, localPath, remotePath string, recursive bool) (string, error) {
	req := common.FileOperationRequest{
		Config:     config,
		Operation:  common.OpUpload,
		LocalPath:  localPath,
		RemotePath: remotePath,
		Recursive:  recursive,
	}
	var resp struct {
		Success    bool   `json:"success"`
		TransferID string `json:"transfer_id"`
		Error      string `json:"error"`
	}
	if err := c.postJSON("/api/operation", req, &resp); err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf("%s", resp.Error)
	}
	return resp.TransferID, nil
}

func (c *APIClient) GetQueue() (*common.QueueResponse, error) {
	var resp common.QueueResponse
	if err := c.getJSON("/api/queue", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) CancelTransfer(transferID string) error {
	req := common.TransferCancelRequest{
		TransferID: transferID,
	}
	var resp common.TransferCancelResponse
	if err := c.deleteJSON("/api/cancel", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}
