package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"go-protocol-adapter/common"
)

const serverURL = "http://localhost:8404"

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

func (c *Client) doRequest(method, path string, body interface{}) (*common.Response, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.Response
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) listAdapters() (*common.Response, error) {
	return c.doRequest(http.MethodGet, "/adapters", nil)
}

func (c *Client) encode(format string, data interface{}) (string, error) {
	resp, err := c.doRequest(http.MethodPost, "/encode?format="+format, data)
	if err != nil {
		return "", err
	}

	if resp.Code != int(common.CodeSuccess) {
		return "", fmt.Errorf("encode failed: %s", resp.Message)
	}

	result, ok := resp.Data.(string)
	if !ok {
		jsonData, _ := json.Marshal(resp.Data)
		return string(jsonData), nil
	}
	return result, nil
}

func (c *Client) decode(format string, data string) (*common.Response, error) {
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/decode?format="+format, strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.Response
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

type ExecuteReq struct {
	InputFormat  string      `json:"input_format"`
	OutputFormat string      `json:"output_format"`
	Operation    string      `json:"operation"`
	Data         interface{} `json:"data"`
}

func (c *Client) execute(inputFormat, outputFormat, operation string, data interface{}) (string, error) {
	reqBody := ExecuteReq{
		InputFormat:  inputFormat,
		OutputFormat: outputFormat,
		Operation:    operation,
		Data:         data,
	}

	resp, err := c.doRequest(http.MethodPost, "/execute", reqBody)
	if err != nil {
		return "", err
	}

	if resp.Code != int(common.CodeSuccess) {
		return "", fmt.Errorf("execute failed: code=%d, message=%s", resp.Code, resp.Message)
	}

	result, ok := resp.Data.(string)
	if !ok {
		jsonData, _ := json.MarshalIndent(resp.Data, "", "  ")
		return string(jsonData), nil
	}
	return result, nil
}

func demoScenario1(client *Client) {
	fmt.Println("==================================================")
	fmt.Println("场景1: JSON转XML和CSV")
	fmt.Println("==================================================")

	jsonData := map[string]interface{}{
		"id":    1,
		"name":  "张三",
		"email": "zhangsan@example.com",
		"age":   25,
	}

	fmt.Println("\n原始JSON数据:")
	jsonStr, _ := json.MarshalIndent(jsonData, "", "  ")
	fmt.Println(string(jsonStr))

	fmt.Println("\n--- 转换为 XML ---")
	xmlResult, err := client.encode("xml", jsonData)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("  结果:\n%s\n", xmlResult)
	}

	fmt.Println("\n--- 转换为 CSV ---")
	csvResult, err := client.encode("csv", jsonData)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("  结果:\n%s\n", csvResult)
	}
}

func demoScenario2(client *Client) {
	fmt.Println("\n==================================================")
	fmt.Println("场景2: CSV创建用户 + JSON/XML查询")
	fmt.Println("==================================================")

	csvData := "id,name,email,age\n,李四,lisi@example.com,30"

	fmt.Println("\nCSV数据 (创建用户):")
	fmt.Println(csvData)

	fmt.Println("\n--- 用CSV格式创建用户 (输入=csv, 输出=json) ---")
	createResult, err := client.execute("csv", "json", "create", csvData)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
		return
	}
	fmt.Printf("  结果 (JSON):\n%s\n", createResult)

	var createdUser map[string]interface{}
	json.Unmarshal([]byte(createResult), &createdUser)
	userID := int64(0)
	if idVal, ok := createdUser["id"]; ok {
		switch v := idVal.(type) {
		case float64:
			userID = int64(v)
		}
	}

	if userID == 0 {
		fmt.Println("\n无法获取用户ID，跳过查询")
		return
	}

	fmt.Printf("\n--- 用JSON格式查询用户ID=%d (输入=json, 输出=json) ---\n", userID)
	getJsonResult, err := client.execute("json", "json", "get", userID)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("  结果:\n%s\n", getJsonResult)
	}

	fmt.Printf("\n--- 用XML格式查询用户ID=%d (输入=json, 输出=xml) ---\n", userID)
	getXmlResult, err := client.execute("json", "xml", "get", userID)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("  结果:\n%s\n", getXmlResult)
	}
}

func demoAdapterList(client *Client) {
	fmt.Println("\n==================================================")
	fmt.Println("已注册的适配器列表")
	fmt.Println("==================================================")

	resp, err := client.listAdapters()
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
		return
	}

	adapters, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(adapters))
}

func printUsage() {
	fmt.Println("Usage: client [command]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  list          List all registered adapters")
	fmt.Println("  demo1         Demo scenario 1: JSON to XML/CSV")
	fmt.Println("  demo2         Demo scenario 2: CSV create + JSON/XML query")
	fmt.Println("  demo          Run all demos")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  client list")
	fmt.Println("  client demo1")
	fmt.Println("  client demo")
}

func main() {
	client := NewClient(serverURL)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := strings.ToLower(os.Args[1])

	switch cmd {
	case "list":
		demoAdapterList(client)
	case "demo1":
		demoScenario1(client)
	case "demo2":
		demoScenario2(client)
	case "demo":
		demoAdapterList(client)
		demoScenario1(client)
		demoScenario2(client)
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
