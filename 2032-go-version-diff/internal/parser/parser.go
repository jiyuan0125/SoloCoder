package parser

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseFile(filePath, content string) (map[string]interface{}, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		return parseJSON(content)
	case ".yaml", ".yml":
		return parseYAML(content)
	default:
		return nil, errors.New("不支持的文件格式")
	}
}

func parseJSON(content string) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func parseYAML(content string) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		return nil, err
	}
	return normalizeYAML(data), nil
}

func normalizeYAML(data map[string]interface{}) map[string]interface{} {
	for k, v := range data {
		data[k] = normalizeValue(v)
	}
	return data
}

func normalizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		return normalizeYAML(val)
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for mk, mv := range val {
			result[fmtKey(mk)] = normalizeValue(mv)
		}
		return result
	case []interface{}:
		for i, item := range val {
			val[i] = normalizeValue(item)
		}
		return val
	default:
		return v
	}
}

func fmtKey(k interface{}) string {
	switch key := k.(type) {
	case string:
		return key
	default:
		return ""
	}
}
