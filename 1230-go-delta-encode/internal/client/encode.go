package client

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"delta-encode/pkg/api"
)

func Encode(serverURL, dataType string, values []string) error {
	dt := api.DataType(strings.ToLower(dataType))
	switch dt {
	case api.DataTypeInt, api.DataTypeFloat:
	default:
		return fmt.Errorf("invalid data type: %s (must be 'int' or 'float')", dataType)
	}

	var data interface{}
	if dt == api.DataTypeInt {
		arr, err := parseIntegers(values)
		if err != nil {
			return fmt.Errorf("failed to parse integers: %w", err)
		}
		data = arr
	} else {
		arr, err := parseFloats(values)
		if err != nil {
			return fmt.Errorf("failed to parse floats: %w", err)
		}
		data = arr
	}

	req := api.EncodeRequest{
		DataType: dt,
		Data:     data,
	}

	respBody, err := postJSON(serverURL+"/encode", req)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := parseJSON(respBody, &result); err != nil {
		return err
	}

	fmt.Println("Encoded result:")
	return prettyPrint(result)
}

func parseIntegers(values []string) ([]int64, error) {
	arr := make([]int64, 0, len(values))
	for i, s := range values {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("value at index %d (%s) is not a valid integer: %w", i, s, err)
		}
		arr = append(arr, v)
	}
	return arr, nil
}

func parseFloats(values []string) ([]float64, error) {
	arr := make([]float64, 0, len(values))
	for i, s := range values {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("value at index %d (%s) is not a valid float: %w", i, s, err)
		}
		arr = append(arr, v)
	}
	return arr, nil
}

func parseJSON(data []byte, v interface{}) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, v)
}
