package main

import (
	"fmt"
	"net/http"
	"time"
)

const (
	HealthCheckMaxRetries = 10
	HealthCheckInterval    = 3 * time.Second
)

func healthCheck(url string, timeoutSeconds int) error {
	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	for i := 0; i < HealthCheckMaxRetries; i++ {
		resp, err := client.Get(url)

		if err != nil {
			Debug("健康检查第 %d 次尝试失败: %v", i+1, err)
			if i < HealthCheckMaxRetries-1 {
				time.Sleep(HealthCheckInterval)
			}
			continue
		}

		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			Debug("健康检查第 %d 次尝试成功，状态码: %d", i+1, resp.StatusCode)
			return nil
		}

		Debug("健康检查第 %d 次尝试，状态码: %d", i+1, resp.StatusCode)
		if i < HealthCheckMaxRetries-1 {
			time.Sleep(HealthCheckInterval)
		}
	}

	return fmt.Errorf("健康检查失败，尝试了 %d 次，状态码非 200 或请求失败", HealthCheckMaxRetries)
}
