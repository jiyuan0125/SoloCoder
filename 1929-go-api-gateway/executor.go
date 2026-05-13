package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Executor struct {
	httpClient *http.Client
}

func NewExecutor() *Executor {
	return &Executor{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (e *Executor) ExecutePlan(plan *Plan) ([]ServiceResult, error) {
	results := make([]ServiceResult, len(plan.Services))
	resultsMap := make(map[string]*ServiceResult)

	for i := range plan.Services {
		results[i].Name = plan.Services[i].Name
		resultsMap[plan.Services[i].Name] = &results[i]
	}

	parallelServices := make([]*ServiceCall, 0)
	for _, svc := range plan.Services {
		if svc.Mode == ExecutionModeParallel {
			svcCopy := svc
			parallelServices = append(parallelServices, &svcCopy)
		}
	}

	if len(parallelServices) > 0 {
		var wg sync.WaitGroup
		var errMu sync.Mutex
		var execError *ExecutionError

		for _, svc := range parallelServices {
			wg.Add(1)
			go func(s *ServiceCall) {
				defer wg.Done()
				result := e.executeService(s)
				errMu.Lock()
				resultsMap[s.Name].Success = result.Success
				resultsMap[s.Name].Error = result.Error
				resultsMap[s.Name].Response = result.Response
				resultsMap[s.Name].Status = result.Status
				if !result.Success && s.FallbackPolicy == FallbackFailAll {
					if execError == nil {
						execError = &ExecutionError{
							ServiceName: s.Name,
							Message:     result.Error,
						}
					}
				}
				errMu.Unlock()
			}(svc)
		}
		wg.Wait()

		if execError != nil {
			return results, execError
		}
	}

	for _, svc := range plan.Services {
		if svc.Mode == ExecutionModeSerial {
			result := e.executeService(&svc)
			resultsMap[svc.Name].Success = result.Success
			resultsMap[svc.Name].Error = result.Error
			resultsMap[svc.Name].Response = result.Response
			resultsMap[svc.Name].Status = result.Status

			if !result.Success && svc.FallbackPolicy == FallbackFailAll {
				return results, &ExecutionError{
					ServiceName: svc.Name,
					Message:     result.Error,
				}
			}
		}
	}

	return results, nil
}

func (e *Executor) executeService(svc *ServiceCall) ServiceResult {
	result := ServiceResult{Name: svc.Name}

	var bodyReader io.Reader
	if svc.Body != nil {
		bodyBytes, err := json.Marshal(svc.Body)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to marshal request body: %v", err)
			return applyFallback(svc, result)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	method := svc.Method
	if method == "" {
		method = http.MethodGet
	}

	req, err := http.NewRequest(method, svc.URL, bodyReader)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to create request: %v", err)
		return applyFallback(svc, result)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range svc.Headers {
		req.Header.Set(k, v)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("request failed: %v", err)
		return applyFallback(svc, result)
	}
	defer resp.Body.Close()

	result.Status = resp.StatusCode

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to read response: %v", err)
		return applyFallback(svc, result)
	}

	if resp.StatusCode >= 400 {
		result.Success = false
		result.Error = fmt.Sprintf("backend returned status %d: %s", resp.StatusCode, string(respBody))
		return applyFallback(svc, result)
	}

	var responseData interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &responseData); err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to parse response: %v", err)
			return applyFallback(svc, result)
		}
	}

	result.Success = true
	result.Response = responseData
	return result
}

func applyFallback(svc *ServiceCall, result ServiceResult) ServiceResult {
	switch svc.FallbackPolicy {
	case FallbackReturnDefault:
		result.Success = true
		result.Response = svc.DefaultValue
		result.Error = ""
	case FallbackIgnore:
		result.Success = true
		result.Response = nil
		result.Error = ""
	case FallbackFailAll:
	}
	return result
}
