package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
)

type MessageType int

const (
	MsgTypeRequest MessageType = iota
	MsgTypeResponse
	MsgTypeError
)

type Request struct {
	ProjectPath  string         `json:"project_path"`
	FilterOpts   FilterOptions  `json:"filter_opts,omitempty"`
}

type Response struct {
	Success  bool         `json:"success"`
	Project  *ProjectInfo `json:"project,omitempty"`
	ErrorMsg string       `json:"error_msg,omitempty"`
}

func EncodeRequest(req *Request) ([]byte, error) {
	return json.Marshal(req)
}

func DecodeRequest(data []byte) (*Request, error) {
	var req Request
	err := json.Unmarshal(data, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request: %w", err)
	}
	if req.ProjectPath == "" {
		return nil, errors.New("project_path is required")
	}
	return &req, nil
}

func EncodeResponse(resp *Response) ([]byte, error) {
	return json.Marshal(resp)
}

func DecodeResponse(data []byte) (*Response, error) {
	var resp Response
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}
