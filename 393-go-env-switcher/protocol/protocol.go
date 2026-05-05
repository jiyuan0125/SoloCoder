package protocol

import (
	"encoding/json"
	"fmt"
)

type CommandType string

const (
	CmdSwitch CommandType = "switch"
	CmdList   CommandType = "list"
	CmdShow   CommandType = "show"
	CmdExport CommandType = "export"
	CmdAdd    CommandType = "add"
	CmdDel    CommandType = "del"
)

type Request struct {
	Cmd    CommandType       `json:"cmd"`
	Env    string            `json:"env,omitempty"`
	Vars   map[string]string `json:"vars,omitempty"`
}

type Response struct {
	Success bool              `json:"success"`
	Message string            `json:"message,omitempty"`
	Envs    []string          `json:"envs,omitempty"`
	Current string            `json:"current,omitempty"`
	Vars    map[string]string `json:"vars,omitempty"`
	Warnings []string         `json:"warnings,omitempty"`
}

func EncodeRequest(req *Request) ([]byte, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("%s\n", data)), nil
}

func DecodeRequest(data []byte) (*Request, error) {
	var req Request
	err := json.Unmarshal(data, &req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func EncodeResponse(resp *Response) ([]byte, error) {
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("%s\n", data)), nil
}

func DecodeResponse(data []byte) (*Response, error) {
	var resp Response
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
