package client

import (
	"fmt"
	"strconv"
	"strings"

	"delta-encode/pkg/api"
)

func Decode(serverURL, dataType, baseStr string, deltaStrs []string) error {
	dt := api.DataType(strings.ToLower(dataType))
	switch dt {
	case api.DataTypeInt, api.DataTypeFloat:
	default:
		return fmt.Errorf("invalid data type: %s (must be 'int' or 'float')", dataType)
	}

	var encoded interface{}

	if dt == api.DataTypeInt {
		base, err := strconv.ParseInt(strings.TrimSpace(baseStr), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid base value '%s': %w", baseStr, err)
		}

		deltas, err := parseIntegers(deltaStrs)
		if err != nil {
			return fmt.Errorf("failed to parse deltas: %w", err)
		}

		encoded = api.IntEncodeResponse{
			Base:   base,
			Deltas: deltas,
		}
	} else {
		base, err := strconv.ParseFloat(strings.TrimSpace(baseStr), 64)
		if err != nil {
			return fmt.Errorf("invalid base value '%s': %w", baseStr, err)
		}

		deltas, err := parseFloats(deltaStrs)
		if err != nil {
			return fmt.Errorf("failed to parse deltas: %w", err)
		}

		encoded = api.FloatEncodeResponse{
			Base:   base,
			Deltas: deltas,
		}
	}

	req := api.DecodeRequest{
		DataType: dt,
		Encoded:  encoded,
	}

	respBody, err := postJSON(serverURL+"/decode", req)
	if err != nil {
		return err
	}

	var result interface{}
	if dt == api.DataTypeInt {
		result = []int64{}
	} else {
		result = []float64{}
	}

	if err := parseJSON(respBody, &result); err != nil {
		return err
	}

	fmt.Println("Decoded result:")
	return prettyPrint(result)
}
