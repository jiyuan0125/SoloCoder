package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"pipeline-system/common"
)

const serverURL = "http://localhost:8080"

func main() {
	fmt.Println("========================================")
	fmt.Println("Pipeline System Demo")
	fmt.Println("========================================")
	fmt.Println()

	fmt.Println("[Step 1] Creating pipeline 'demo_pipeline'...")
	if err := createPipeline("demo_pipeline"); err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Println("  ✓ Pipeline created successfully")
	fmt.Println()

	fmt.Println("[Step 2] Creating 3 handlers...")
	handler1, err := createHandler("handler_validate", "Validate Handler", "validate")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  ✓ Handler 1 created: ID=%s\n", handler1.ID)

	handler2, err := createHandler("", "Transform Handler", "transform")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  ✓ Handler 2 created: ID=%s (auto-generated)\n", handler2.ID)

	handler3, err := createHandler("handler_echo", "Echo Handler", "echo")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  ✓ Handler 3 created: ID=%s\n", handler3.ID)
	fmt.Println()

	fmt.Println("[Step 3] Creating 1 async middleware (timeout: 2000ms)...")
	mw, err := createMiddleware("middleware_auth", "Async Auth Middleware", true, 2000)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  ✓ Async middleware created: ID=%s\n", mw.ID)
	fmt.Println()

	fmt.Println("[Step 4] Adding middleware and handlers to pipeline...")
	if err := addHandlerToPipeline("demo_pipeline", mw.ID, 0); err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  ✓ Added middleware at position 0\n")

	if err := addHandlerToPipeline("demo_pipeline", handler1.ID, -1); err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  ✓ Added handler 1 at end\n")

	if err := addHandlerToPipeline("demo_pipeline", handler2.ID, -1); err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  ✓ Added handler 2 at end\n")

	if err := addHandlerToPipeline("demo_pipeline", handler3.ID, -1); err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  ✓ Added handler 3 at end\n")
	fmt.Println()

	fmt.Println("[Step 5] Checking pipeline configuration...")
	info, err := getPipeline("demo_pipeline")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  Pipeline: %s\n", info.Name)
	fmt.Printf("  Handlers in order:\n")
	for i, id := range info.Handlers {
		fmt.Printf("    [%d] %s\n", i, id)
	}
	fmt.Println()

	fmt.Println("[Step 6] Sending test request...")
	testBody := "Hello, Pipeline System!"
	result, err := executeRequest("demo_pipeline", testBody)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  ✓ Request executed successfully\n")
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("Execution Results")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Printf("Request ID: %s\n", result.RequestID)
	fmt.Printf("Pipeline:   %s\n", result.Pipeline)
	fmt.Printf("Status:     %s\n", result.Status)
	fmt.Printf("Total Time: %d ms\n", result.DurationMs)
	fmt.Println()

	fmt.Println("Handler Execution Details:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-4s | %-30s | %-10s | %-8s | %s\n", "#", "Handler Name", "Status", "Time(ms)", "ID")
	fmt.Println(strings.Repeat("-", 80))
	for i, hr := range result.HandlerResults {
		status := hr.Status
		if hr.Panic {
			status = "PANIC"
		}
		fmt.Printf("%-4d | %-30s | %-10s | %-8d | %s\n", i+1, hr.HandlerName, status, hr.DurationMs, hr.HandlerID)
		if hr.Error != "" {
			fmt.Printf("       Error: %s\n", hr.Error)
		}
	}
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println()

	fmt.Println("Response:")
	fmt.Printf("  Status Code: %d\n", result.Response.StatusCode)
	fmt.Printf("  Body: %s\n", result.Response.Body)
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("Demo Complete!")
	fmt.Println("========================================")
}

func httpPost(url string, body interface{}, resp interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return err
	}
	res, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		data, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("HTTP %d: %s", res.StatusCode, string(data))
	}
	if resp != nil {
		return json.NewDecoder(res.Body).Decode(resp)
	}
	return nil
}

func httpPut(url string, body interface{}, resp interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		data, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("HTTP %d: %s", res.StatusCode, string(data))
	}
	if resp != nil {
		return json.NewDecoder(res.Body).Decode(resp)
	}
	return nil
}

func httpGet(url string, resp interface{}) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		data, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("HTTP %d: %s", res.StatusCode, string(data))
	}
	if resp != nil {
		return json.NewDecoder(res.Body).Decode(resp)
	}
	return nil
}

func createPipeline(name string) error {
	return httpPost(serverURL+"/pipelines", common.CreatePipelineRequest{Name: name}, nil)
}

func getPipeline(name string) (*common.PipelineInfo, error) {
	var info common.PipelineInfo
	err := httpGet(serverURL+"/pipelines/"+name, &info)
	return &info, err
}

func createHandler(id, name, handlerType string) (*common.HandlerInfo, error) {
	var info common.HandlerInfo
	err := httpPost(serverURL+"/handlers", common.CreateHandlerRequest{
		ID:   id,
		Name: name,
		Type: handlerType,
	}, &info)
	return &info, err
}

func createMiddleware(id, name string, async bool, timeoutMs int64) (*common.MiddlewareInfo, error) {
	var info common.MiddlewareInfo
	err := httpPost(serverURL+"/middlewares", common.CreateMiddlewareRequest{
		ID:        id,
		Name:      name,
		Async:     async,
		TimeoutMs: timeoutMs,
	}, &info)
	return &info, err
}

func addHandlerToPipeline(pipeline, handlerID string, position int) error {
	return httpPut(serverURL+"/pipelines/"+pipeline, map[string]interface{}{
		"action":     "add_handler",
		"handler_id": handlerID,
		"position":   position,
	}, nil)
}

func executeRequest(pipeline, body string) (*common.ExecutionResult, error) {
	var result common.ExecutionResult
	err := httpPost(serverURL+"/execute", common.Request{
		Pipeline: pipeline,
		Header: map[string]string{
			"Content-Type": "text/plain",
			"X-Request-Id": "demo-request-001",
		},
		Body: body,
	}, &result)
	return &result, err
}
