package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const defaultServerURL = "http://localhost:8080"

func GetServerURL() string {
	if url := os.Getenv("SERVER_URL"); url != "" {
		return url
	}
	return defaultServerURL
}

func SendRequest(method string, path string, data interface{}) ([]byte, error) {
	url := GetServerURL() + path

	var body bytes.Buffer
	if data != nil {
		if err := json.NewEncoder(&body).Encode(data); err != nil {
			return nil, fmt.Errorf("编码请求数据失败: %v", err)
		}
	}

	req, err := http.NewRequest(method, url, &body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("服务器返回错误状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func ParseResponse(respBytes []byte, v interface{}) error {
	return json.Unmarshal(respBytes, v)
}

func PrintJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Println(string(data))
}
