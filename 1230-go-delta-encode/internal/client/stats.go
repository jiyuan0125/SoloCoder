package client

import (
	"fmt"
	"strings"

	"delta-encode/pkg/api"
)

func Stats(serverURL, dataType string, values []string) error {
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

	req := api.StatsRequest{
		DataType: dt,
		Data:     data,
	}

	respBody, err := postJSON(serverURL+"/stats", req)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := parseJSON(respBody, &result); err != nil {
		return err
	}

	fmt.Println("Statistics:")
	return prettyPrint(result)
}
