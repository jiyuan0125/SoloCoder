package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"data-exporter/pkg/proto"
)

func run(format, fields, fieldMapping, serverURL, outputFile string) error {
	exportFormat, err := parseFormat(format)
	if err != nil {
		return err
	}

	fieldList := parseFields(fields)
	mapping := parseFieldMapping(fieldMapping)

	data, err := readInputData()
	if err != nil {
		return fmt.Errorf("failed to read input data: %v", err)
	}

	req := proto.ExportRequest{
		Data:        data,
		Format:      exportFormat,
		Fields:      fieldList,
		FieldMapping: mapping,
	}

	resp, err := sendRequest(serverURL, req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("export failed: %s", resp.Message)
	}

	if err := writeOutput(resp.Data, outputFile); err != nil {
		return fmt.Errorf("failed to write output: %v", err)
	}

	return nil
}

func parseFormat(format string) (proto.ExportFormat, error) {
	switch strings.ToLower(format) {
	case "csv":
		return proto.FormatCSV, nil
	case "json":
		return proto.FormatJSON, nil
	case "jsonl":
		return proto.FormatJSONLines, nil
	default:
		return "", fmt.Errorf("invalid format: %s (supported: csv, json, jsonl)", format)
	}
}

func parseFields(fields string) []string {
	if fields == "" {
		return nil
	}
	parts := strings.Split(fields, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseFieldMapping(mapping string) map[string]string {
	if mapping == "" {
		return nil
	}
	result := make(map[string]string)
	pairs := strings.Split(mapping, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" {
				result[key] = value
			}
		}
	}
	return result
}

func readInputData() ([]map[string]string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no input data provided")
	}

	var result []map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %v", err)
	}

	return result, nil
}

func sendRequest(serverURL string, req proto.ExportRequest) (*proto.ExportResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := serverURL + "/export"
	httpResp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", httpResp.StatusCode, string(errBody))
	}

	var resp proto.ExportResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func writeOutput(data, outputFile string) error {
	var w io.Writer
	if outputFile == "" {
		w = os.Stdout
	} else {
		f, err := os.Create(outputFile)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	if _, err := io.WriteString(w, data); err != nil {
		return err
	}
	return nil
}
