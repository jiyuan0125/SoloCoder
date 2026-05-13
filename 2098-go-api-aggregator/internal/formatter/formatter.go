package formatter

import (
	"encoding/json"
	"fmt"
	"strings"
)

type FormatType string

const (
	FormatMerge    FormatType = "merge"
	FormatExtract  FormatType = "extract"
	FormatArray    FormatType = "array"
	FormatResponse FormatType = "response"
)

type FormatConfig struct {
	Type     FormatType              `json:"type"`
	Extracts map[string]ExtractField `json:"extracts,omitempty"`
}

type ExtractField struct {
	EndpointIndex int    `json:"endpoint_index"`
	Path          string `json:"path"`
}

type EndpointResult struct {
	Index      int
	Endpoint   string
	StatusCode int
	Body       []byte
	Success    bool
	Error      string
}

func FormatResults(results []*EndpointResult, formatType FormatType, formatConfig map[string]interface{}) ([]byte, error) {
	switch formatType {
	case FormatMerge:
		return mergeResults(results)
	case FormatExtract:
		return extractFields(results, formatConfig)
	case FormatArray:
		return arrayResults(results)
	case FormatResponse:
		return responseDetails(results)
	default:
		return nil, fmt.Errorf("invalid format type: %s", formatType)
	}
}

func mergeResults(results []*EndpointResult) ([]byte, error) {
	merged := make(map[string]interface{})

	for _, r := range results {
		if !r.Success {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal(r.Body, &data); err != nil {
			continue
		}

		for k, v := range data {
			if existing, ok := merged[k]; ok {
				switch ev := existing.(type) {
				case []interface{}:
					switch vv := v.(type) {
					case []interface{}:
						ev = append(ev, vv...)
					default:
						ev = append(ev, v)
					}
					merged[k] = ev
				case map[string]interface{}:
					if vm, ok := v.(map[string]interface{}); ok {
						for k2, v2 := range vm {
							ev[k2] = v2
						}
						merged[k] = ev
					}
				default:
					merged[k] = v
				}
			} else {
				merged[k] = v
			}
		}
	}

	return json.Marshal(merged)
}

func extractFields(results []*EndpointResult, config map[string]interface{}) ([]byte, error) {
	extracts, ok := config["extracts"]
	if !ok {
		return nil, fmt.Errorf("format config missing 'extracts'")
	}

	extractsMap, ok := extracts.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("'extracts' must be an object")
	}

	resultMap := make(map[string]interface{})

	for fieldKey, extractSpec := range extractsMap {
		spec, ok := extractSpec.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid extract spec for field: %s", fieldKey)
		}

		endpointIdx := 0
		if idx, ok := spec["endpoint_index"].(float64); ok {
			endpointIdx = int(idx)
		}

		path, _ := spec["path"].(string)

		if endpointIdx < 0 || endpointIdx >= len(results) {
			resultMap[fieldKey] = nil
			continue
		}

		r := results[endpointIdx]
		if !r.Success {
			resultMap[fieldKey] = nil
			continue
		}

		var data interface{}
		if err := json.Unmarshal(r.Body, &data); err != nil {
			resultMap[fieldKey] = nil
			continue
		}

		resultMap[fieldKey] = getValueByPath(data, path)
	}

	return json.Marshal(resultMap)
}

func getValueByPath(data interface{}, path string) interface{} {
	if path == "" {
		return data
	}

	current := data
	parts := splitPath(path)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		default:
			return nil
		}

		if current == nil {
			return nil
		}
	}

	return current
}

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	var parts []string
	var current strings.Builder

	for i := 0; i < len(path); i++ {
		c := path[i]
		if c == '.' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else if c == '[' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else if c == ']' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(c)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func arrayResults(results []*EndpointResult) ([]byte, error) {
	var arr []interface{}

	for _, r := range results {
		if !r.Success {
			arr = append(arr, map[string]interface{}{
				"index":     r.Index,
				"endpoint":  r.Endpoint,
				"success":   false,
				"error":     r.Error,
				"status":    r.StatusCode,
			})
			continue
		}

		var data interface{}
		if err := json.Unmarshal(r.Body, &data); err != nil {
			arr = append(arr, map[string]interface{}{
				"index":     r.Index,
				"endpoint":  r.Endpoint,
				"success":   false,
				"error":     err.Error(),
				"raw_body":  string(r.Body),
			})
			continue
		}

		arr = append(arr, map[string]interface{}{
			"index":     r.Index,
			"endpoint":  r.Endpoint,
			"success":   true,
			"status":    r.StatusCode,
			"data":      data,
		})
	}

	return json.Marshal(arr)
}

func responseDetails(results []*EndpointResult) ([]byte, error) {
	details := make([]map[string]interface{}, 0, len(results))

	for _, r := range results {
		detail := map[string]interface{}{
			"index":    r.Index,
			"endpoint": r.Endpoint,
			"success":  r.Success,
			"status":   r.StatusCode,
		}

		if r.Success {
			var data interface{}
			if err := json.Unmarshal(r.Body, &data); err != nil {
				detail["raw_body"] = string(r.Body)
			} else {
				detail["data"] = data
			}
		} else {
			detail["error"] = r.Error
		}

		details = append(details, detail)
	}

	return json.Marshal(map[string]interface{}{
		"results": details,
	})
}
