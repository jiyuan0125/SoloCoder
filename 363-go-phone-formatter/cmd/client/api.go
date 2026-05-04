package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"phone-formatter/pkg/api"
)

const serverURL = "http://localhost:8080"

func FormatPhone(phone string, formatType string) (string, error) {
	req := api.FormatRequest{
		Phone:      phone,
		FormatType: formatType,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(serverURL+"/format", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result api.FormatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if !result.Success {
		return "", fmt.Errorf(result.Error)
	}

	return result.Result, nil
}

func ValidatePhone(phone string) (bool, error) {
	req := api.ValidateRequest{
		Phone: phone,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return false, err
	}

	resp, err := http.Post(serverURL+"/validate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result api.ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	if !result.Success {
		return false, fmt.Errorf(result.Error)
	}

	return result.Valid, nil
}

func ExtractPhones(text string) ([]string, error) {
	req := api.ExtractRequest{
		Text: text,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/extract", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.ExtractResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return result.Results, nil
}
