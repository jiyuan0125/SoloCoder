package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type Checker struct {
	manager   *ComponentManager
	tickers   map[string]*time.Ticker
	mu        sync.RWMutex
	stopChans map[string]chan struct{}
}

func NewChecker(manager *ComponentManager) *Checker {
	return &Checker{
		manager:   manager,
		tickers:   make(map[string]*time.Ticker),
		stopChans: make(map[string]chan struct{}),
	}
}

func (c *Checker) CheckHTTP(address string, timeout time.Duration) CheckResult {
	start := time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return CheckResult{
			Success:   false,
			Latency:   time.Since(start),
			Timestamp: time.Now(),
			Error:     fmt.Sprintf("failed to create request: %v", err),
		}
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return CheckResult{
			Success:   false,
			Latency:   time.Since(start),
			Timestamp: time.Now(),
			Error:     fmt.Sprintf("request failed: %v", err),
		}
	}
	defer resp.Body.Close()
	
	success := resp.StatusCode >= 200 && resp.StatusCode < 400
	result := CheckResult{
		Success:   success,
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}
	
	if !success {
		result.Error = fmt.Sprintf("status code %d", resp.StatusCode)
	}
	
	return result
}

func (c *Checker) CheckTCP(address string, timeout time.Duration) CheckResult {
	start := time.Now()
	
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return CheckResult{
			Success:   false,
			Latency:   time.Since(start),
			Timestamp: time.Now(),
			Error:     fmt.Sprintf("connection failed: %v", err),
		}
	}
	defer conn.Close()
	
	return CheckResult{
		Success:   true,
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}
}

func (c *Checker) CheckComponent(name string) {
	comp, exists := c.manager.Get(name)
	if !exists {
		return
	}
	
	var result CheckResult
	switch comp.Config.CheckType {
	case CheckTypeHTTP:
		result = c.CheckHTTP(comp.Config.Address, comp.Config.Timeout)
	case CheckTypeTCP:
		result = c.CheckTCP(comp.Config.Address, comp.Config.Timeout)
	default:
		result = CheckResult{
			Success:   false,
			Latency:   0,
			Timestamp: time.Now(),
			Error:     fmt.Sprintf("unknown check type: %s", comp.Config.CheckType),
		}
	}
	
	stateChanged := c.manager.UpdateResult(name, result)
	
	if stateChanged && comp.Config.CallbackURL != "" {
		go c.sendCallback(comp.Config.CallbackURL, comp.Name, !result.Success)
	}
}

type CallbackPayload struct {
	ComponentName string `json:"componentName"`
	Healthy       bool   `json:"healthy"`
	Timestamp     time.Time `json:"timestamp"`
}

func (c *Checker) sendCallback(url string, componentName string, healthy bool) {
	payload := CallbackPayload{
		ComponentName: componentName,
		Healthy:       healthy,
		Timestamp:     time.Now(),
	}
	
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("callback marshal error for %s: %v", componentName, err)
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		log.Printf("callback request error for %s: %v", componentName, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("callback failed for %s: %v", componentName, err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("callback sent successfully for %s", componentName)
	} else {
		log.Printf("callback returned status %d for %s", resp.StatusCode, componentName)
	}
}

func (c *Checker) StartComponent(name string) {
	comp, exists := c.manager.Get(name)
	if !exists {
		return
	}
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if _, exists := c.tickers[name]; exists {
		return
	}
	
	ticker := time.NewTicker(comp.Config.Interval)
	stopChan := make(chan struct{})
	
	c.tickers[name] = ticker
	c.stopChans[name] = stopChan
	
	go func() {
		c.CheckComponent(name)
		for {
			select {
			case <-ticker.C:
				c.CheckComponent(name)
			case <-stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (c *Checker) StopComponent(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if stopChan, exists := c.stopChans[name]; exists {
		close(stopChan)
		delete(c.stopChans, name)
	}
	
	if ticker, exists := c.tickers[name]; exists {
		ticker.Stop()
		delete(c.tickers, name)
	}
}

func (c *Checker) StopAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for name, stopChan := range c.stopChans {
		close(stopChan)
		delete(c.stopChans, name)
	}
	
	for name, ticker := range c.tickers {
		ticker.Stop()
		delete(c.tickers, name)
	}
}
