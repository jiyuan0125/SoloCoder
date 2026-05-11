package api

import (
	"encoding/json"
)

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type AnalyzeRequest struct {
	PackagePath string `json:"package_path"`
	Files       []File `json:"files"`
	Format      string `json:"format"`
}

type AnalyzeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Format  string `json:"format,omitempty"`
	Output  string `json:"output,omitempty"`
}

func EncodeRequest(req AnalyzeRequest) ([]byte, error) {
	return json.Marshal(req)
}

func DecodeRequest(data []byte) (AnalyzeRequest, error) {
	var req AnalyzeRequest
	err := json.Unmarshal(data, &req)
	return req, err
}

func EncodeResponse(resp AnalyzeResponse) ([]byte, error) {
	return json.Marshal(resp)
}

func DecodeResponse(data []byte) (AnalyzeResponse, error) {
	var resp AnalyzeResponse
	err := json.Unmarshal(data, &resp)
	return resp, err
}
