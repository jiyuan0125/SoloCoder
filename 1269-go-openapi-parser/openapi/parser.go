package openapi

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseFile(filename string) (*OpenAPI, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return Parse(data, filename)
}

func Parse(data []byte, filename string) (*OpenAPI, error) {
	var raw map[string]interface{}
	var err error

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &raw)
	case ".json":
		err = json.Unmarshal(data, &raw)
	default:
		err = yaml.Unmarshal(data, &raw)
		if err != nil {
			err = json.Unmarshal(data, &raw)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	var spec OpenAPI
	if err := convertToStruct(raw, &spec); err != nil {
		return nil, fmt.Errorf("failed to convert to struct: %w", err)
	}
	spec.Raw = raw
	return &spec, nil
}

func convertToStruct(raw map[string]interface{}, spec *OpenAPI) error {
	jsonData, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, spec)
}

func ReadAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
