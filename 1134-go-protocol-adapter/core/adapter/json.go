package adapter

import (
	"encoding/json"
	"strings"
)

type JSONAdapter struct{}

func NewJSONAdapter() *JSONAdapter {
	return &JSONAdapter{}
}

func (a *JSONAdapter) Name() string {
	return "json"
}

func (a *JSONAdapter) CanHandle(format string) bool {
	return strings.EqualFold(format, "json")
}

func (a *JSONAdapter) Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (a *JSONAdapter) Encode(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func (a *JSONAdapter) Priority() int {
	return 100
}
