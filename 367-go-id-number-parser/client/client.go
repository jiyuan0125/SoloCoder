package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"id-number-parser/common"
	"net/http"
	"os"
)

const serverURL = "http://localhost:8080"

func handleParse(idNumber string) {
	req := common.ParseRequest{
		IDNumber: idNumber,
	}

	resp, err := sendRequest("/api/parse", req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result common.ParseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
		os.Exit(1)
	}

	printParseResult(result)
}

func handleValidate(idNumber string) {
	req := common.ValidateRequest{
		IDNumber: idNumber,
	}

	resp, err := sendRequest("/api/validate", req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result common.ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
		os.Exit(1)
	}

	printValidateResult(result)
}

func sendRequest(endpoint string, request interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	url := serverURL + endpoint
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	return resp, nil
}
