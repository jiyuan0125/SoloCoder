package common

import "encoding/base64"

type InspectRequest struct {
	Buffer string            `json:"buffer"`
	Schema *SchemaDefinition `json:"schema"`
}

type SchemaDefinition struct {
	Name   string               `json:"name"`
	Fields []FieldDefinition    `json:"fields"`
}

type FieldDefinition struct {
	Name     string             `json:"name"`
	Type     string             `json:"type"`
	ElemType string             `json:"elem_type,omitempty"`
	Nested   *SchemaDefinition  `json:"nested,omitempty"`
}

type InspectResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message,omitempty"`
	Result  map[string]interface{} `json:"result,omitempty"`
}

func (req *InspectRequest) DecodeBuffer() ([]byte, error) {
	return base64.StdEncoding.DecodeString(req.Buffer)
}

func EncodeBuffer(buf []byte) string {
	return base64.StdEncoding.EncodeToString(buf)
}
