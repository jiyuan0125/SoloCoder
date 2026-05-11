package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/drivingschool/common"
)

func makeGetRequest(path string) (common.Response, error) {
	url := serverURL + path
	resp, err := http.Get(url)
	if err != nil {
		return common.Response{}, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	return parseResponse(resp)
}

func makePostRequest(path string, body interface{}) (common.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.Response{}, fmt.Errorf("序列化请求体失败: %v", err)
	}

	url := serverURL + path
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return common.Response{}, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	return parseResponse(resp)
}

func parseResponse(resp *http.Response) (common.Response, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.Response{}, fmt.Errorf("读取响应失败: %v", err)
	}

	var response common.Response
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return common.Response{}, fmt.Errorf("解析响应失败: %v", err)
	}

	return response, nil
}

func prettyPrint(v interface{}) {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("%v\n", v)
		return
	}
	fmt.Println(string(jsonBytes))
}
