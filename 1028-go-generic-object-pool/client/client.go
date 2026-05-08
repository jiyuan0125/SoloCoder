package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"generic-object-pool/common"
	"net/http"
)

type Client struct {
	BaseURL string
}

func (c *Client) url(path string) string {
	return c.BaseURL + path
}

func (c *Client) CreatePool(poolID string, maxSize int, idleTimeout string) error {
	req := common.CreatePoolRequest{
		PoolID:      poolID,
		MaxSize:     maxSize,
		IdleTimeout: idleTimeout,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.url("/pool/create"), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result common.CreatePoolResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Message)
	}

	fmt.Printf("Created pool '%s'\n", poolID)
	return nil
}

func (c *Client) GetObject(poolID string) error {
	req := common.GetRequest{PoolID: poolID}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.url("/pool/get"), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result common.GetResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Message)
	}

	objJSON, _ := json.MarshalIndent(result.Object, "", "  ")
	fmt.Printf("Got object: %s\n", string(objJSON))
	return nil
}

func (c *Client) PutObject(poolID string, data string) error {
	req := common.PutRequest{
		PoolID: poolID,
		Object: &common.PoolObject{Data: data},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.url("/pool/put"), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result common.PutResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Message)
	}

	fmt.Printf("Returned object to pool '%s'\n", poolID)
	return nil
}

func (c *Client) GetStats(poolID string) error {
	req := common.StatsRequest{PoolID: poolID}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.url("/pool/stats"), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result common.StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Message)
	}

	fmt.Printf("Pool '%s' stats:\n", poolID)
	fmt.Printf("  Free objects:   %d\n", result.Free)
	fmt.Printf("  Created total:  %d\n", result.Created)
	fmt.Printf("  Discarded:      %d\n", result.Discarded)
	fmt.Printf("  Hits:           %d\n", result.Hits)
	fmt.Printf("  Misses:         %d\n", result.Misses)
	fmt.Printf("  Hit rate:       %.2f%%\n", result.HitRate*100)
	return nil
}

func (c *Client) ClosePool(poolID string) error {
	req := common.ClosePoolRequest{PoolID: poolID}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.url("/pool/close"), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result common.ClosePoolResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Message)
	}

	fmt.Printf("Closed pool '%s'\n", poolID)
	return nil
}
