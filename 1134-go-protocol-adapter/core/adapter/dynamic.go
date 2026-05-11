package adapter

import (
	"encoding/json"
	"fmt"
	"strings"
)

type DynamicAdapterConfig struct {
	Name        string      `json:"name"`
	Formats     []string    `json:"formats"`
	DecodeSample interface{} `json:"decode_sample"`
	EncodeSample interface{} `json:"encode_sample"`
	Priority    int         `json:"priority"`
}

type DynamicAdapter struct {
	name      string
	formats   []string
	priority  int
	converter *TypeConverter
}

func NewDynamicAdapter(config DynamicAdapterConfig) (*DynamicAdapter, error) {
	if config.Name == "" {
		return nil, fmt.Errorf("adapter name is required")
	}
	if len(config.Formats) == 0 {
		return nil, fmt.Errorf("at least one format is required")
	}

	if config.DecodeSample != nil {
		sampleJSON, err := json.Marshal(config.DecodeSample)
		if err != nil {
			return nil, fmt.Errorf("invalid decode sample: %w", err)
		}
		var testObj map[string]interface{}
		if err := json.Unmarshal(sampleJSON, &testObj); err != nil {
			return nil, fmt.Errorf("decode sample validation failed: %w", err)
		}
	}

	if config.EncodeSample != nil {
		if _, err := json.Marshal(config.EncodeSample); err != nil {
			return nil, fmt.Errorf("invalid encode sample: %w", err)
		}
	}

	priority := config.Priority
	if priority == 0 {
		priority = 50
	}

	return &DynamicAdapter{
		name:      config.Name,
		formats:   config.Formats,
		priority:  priority,
		converter: NewTypeConverter(),
	}, nil
}

func (a *DynamicAdapter) Name() string {
	return a.name
}

func (a *DynamicAdapter) CanHandle(format string) bool {
	for _, f := range a.formats {
		if strings.EqualFold(f, format) {
			return true
		}
	}
	return false
}

func (a *DynamicAdapter) Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (a *DynamicAdapter) Encode(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func (a *DynamicAdapter) Priority() int {
	return a.priority
}
