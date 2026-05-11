package tagparser

import (
	"strings"

	"go-reflect-tags/pkg/common"
)

func splitTagOptions(tag string) []string {
	if tag == "" {
		return nil
	}
	var parts []string
	var current strings.Builder
	inQuote := false

	for i := 0; i < len(tag); i++ {
		c := tag[i]
		if c == '"' {
			inQuote = !inQuote
			current.WriteByte(c)
			continue
		}
		if c == ',' && !inQuote {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteByte(c)
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

func parseKeyValue(option string) (key, value string) {
	idx := strings.Index(option, "=")
	if idx == -1 {
		return option, ""
	}
	key = strings.TrimSpace(option[:idx])
	value = strings.TrimSpace(option[idx+1:])
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}
	return key, value
}

func isNumericType(fieldType string) bool {
	switch fieldType {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		return true
	}
	return false
}

func isStringType(fieldType string) bool {
	return fieldType == "string"
}

func mergeTags(parent, child map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range child {
		result[k] = v
	}
	for k, v := range parent {
		result[k] = v
	}
	return result
}

type Parser struct {
	structs map[string]*common.FieldDesc
}

func NewParser() *Parser {
	return &Parser{
		structs: make(map[string]*common.FieldDesc),
	}
}

func (p *Parser) RegisterStruct(name string, fields []common.FieldDesc) {
	root := &common.FieldDesc{
		Name:     name,
		Type:     name,
		Fields:   fields,
		Tags:     make(map[string]string),
	}
	p.structs[name] = root
}

func (p *Parser) GetStruct(name string) (*common.FieldDesc, bool) {
	s, ok := p.structs[name]
	return s, ok
}

func (p *Parser) ListStructs() []string {
	result := make([]string, 0, len(p.structs))
	for k := range p.structs {
		result = append(result, k)
	}
	return result
}
