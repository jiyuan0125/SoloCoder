package schema

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type ValidationError struct {
	Path       string
	Constraint string
	Actual     interface{}
	Expected   interface{}
}

func (e ValidationError) String() string {
	base := fmt.Sprintf("%s: 违反%s约束", e.Path, e.Constraint)
	if e.Actual != nil {
		base += fmt.Sprintf(", 实际值: %v", e.Actual)
	}
	if e.Expected != nil {
		base += fmt.Sprintf(", 期望值: %v", e.Expected)
	}
	return base
}

type Validator struct {
	rootSchema interface{}
	schemaMap  map[string]interface{}
}

func NewValidator(schemaJSON []byte) (*Validator, error) {
	var schema interface{}
	if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		return nil, fmt.Errorf("解析schema失败: %w", err)
	}
	return &Validator{
		rootSchema: schema,
		schemaMap:  map[string]interface{}{},
	}, nil
}

func (v *Validator) Validate(dataJSON []byte) []ValidationError {
	var data interface{}
	if err := json.Unmarshal(dataJSON, &data); err != nil {
		return []ValidationError{
			{
				Path:       "$",
				Constraint: "parse",
				Actual:     string(dataJSON),
				Expected:   "valid JSON",
			},
		}
	}
	ctx := newContext(v.rootSchema)
	return ctx.validate("$", v.rootSchema, data)
}

type context struct {
	root    interface{}
	visited map[string]bool
}

func newContext(root interface{}) *context {
	return &context{
		root:    root,
		visited: map[string]bool{},
	}
}

func (ctx *context) validate(path string, schema interface{}, data interface{}) []ValidationError {
	errors := []ValidationError{}

	if schema == nil {
		return errors
	}

	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		return errors
	}

	if ref, exists := schemaMap["$ref"]; exists {
		refStr, ok := ref.(string)
		if ok {
			resolvedSchema, resolveErr := ctx.resolveRef(refStr)
			if resolveErr != nil {
				return append(errors, ValidationError{
					Path:       path,
					Constraint: "$ref",
					Actual:     refStr,
					Expected:   "valid $ref",
				})
			}
			return ctx.validate(path, resolvedSchema, data)
		}
	}

	if schemaType, exists := schemaMap["type"]; exists {
		typeErrors := ctx.checkType(path, schemaType, data)
		errors = append(errors, typeErrors...)
		if len(typeErrors) > 0 {
			return errors
		}
	}

	if enum, exists := schemaMap["enum"]; exists {
		enumErrors := ctx.checkEnum(path, enum, data)
		errors = append(errors, enumErrors...)
	}

	switch d := data.(type) {
	case string:
		errors = append(errors, ctx.validateString(path, schemaMap, d)...)
	case float64:
		errors = append(errors, ctx.validateNumber(path, schemaMap, d)...)
	case bool:
	case nil:
	case []interface{}:
		errors = append(errors, ctx.validateArray(path, schemaMap, d)...)
	case map[string]interface{}:
		errors = append(errors, ctx.validateObject(path, schemaMap, d)...)
	}

	if allOf, exists := schemaMap["allOf"]; exists {
		errors = append(errors, ctx.validateAllOf(path, allOf, data)...)
	}

	if anyOf, exists := schemaMap["anyOf"]; exists {
		errors = append(errors, ctx.validateAnyOf(path, anyOf, data)...)
	}

	if oneOf, exists := schemaMap["oneOf"]; exists {
		errors = append(errors, ctx.validateOneOf(path, oneOf, data)...)
	}

	if not, exists := schemaMap["not"]; exists {
		errors = append(errors, ctx.validateNot(path, not, data)...)
	}

	return errors
}

func (ctx *context) checkType(path string, schemaType interface{}, data interface{}) []ValidationError {
	errors := []ValidationError{}

	expectedTypes := []string{}
	switch t := schemaType.(type) {
	case string:
		expectedTypes = []string{t}
	case []interface{}:
		for _, v := range t {
			if s, ok := v.(string); ok {
				expectedTypes = append(expectedTypes, s)
			}
		}
	default:
		return errors
	}

	actualType := getJSONType(data)
	for _, et := range expectedTypes {
		if et == actualType {
			return errors
		}
	}

	return append(errors, ValidationError{
		Path:       path,
		Constraint: "type",
		Actual:     actualType,
		Expected:   strings.Join(expectedTypes, "/"),
	})
}

func getJSONType(data interface{}) string {
	switch data.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case nil:
		return "null"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

func (ctx *context) checkEnum(path string, enum interface{}, data interface{}) []ValidationError {
	errors := []ValidationError{}
	enumArr, ok := enum.([]interface{})
	if !ok {
		return errors
	}

	for _, expected := range enumArr {
		if deepEqual(data, expected) {
			return errors
		}
	}

	return append(errors, ValidationError{
		Path:       path,
		Constraint: "enum",
		Actual:     data,
		Expected:   enumArr,
	})
}

func deepEqual(a, b interface{}) bool {
	aJSON, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bJSON, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(aJSON) == string(bJSON)
}

func (ctx *context) validateString(path string, schema map[string]interface{}, data string) []ValidationError {
	errors := []ValidationError{}

	if minLength, exists := schema["minLength"]; exists {
		if min, ok := toInt(minLength); ok {
			if len(data) < min {
				errors = append(errors, ValidationError{
					Path:       path,
					Constraint: "minLength",
					Actual:     len(data),
					Expected:   min,
				})
			}
		}
	}

	if maxLength, exists := schema["maxLength"]; exists {
		if max, ok := toInt(maxLength); ok {
			if len(data) > max {
				errors = append(errors, ValidationError{
					Path:       path,
					Constraint: "maxLength",
					Actual:     len(data),
					Expected:   max,
				})
			}
		}
	}

	if pattern, exists := schema["pattern"]; exists {
		if patternStr, ok := pattern.(string); ok {
			re, err := regexp.Compile(patternStr)
			if err == nil && !re.MatchString(data) {
				errors = append(errors, ValidationError{
					Path:       path,
					Constraint: "pattern",
					Actual:     data,
					Expected:   patternStr,
				})
			}
		}
	}

	return errors
}

func (ctx *context) validateNumber(path string, schema map[string]interface{}, data float64) []ValidationError {
	errors := []ValidationError{}

	if schemaType, exists := schema["type"]; exists {
		if typeStr, ok := schemaType.(string); ok && typeStr == "integer" {
			if data != float64(int64(data)) {
				errors = append(errors, ValidationError{
					Path:       path,
					Constraint: "integer",
					Actual:     data,
					Expected:   "integer value",
				})
				return errors
			}
		}
	}

	if minimum, exists := schema["minimum"]; exists {
		if min, ok := toFloat(minimum); ok {
			if data < min {
				errors = append(errors, ValidationError{
					Path:       path,
					Constraint: "minimum",
					Actual:     data,
					Expected:   min,
				})
			}
		}
	}

	if maximum, exists := schema["maximum"]; exists {
		if max, ok := toFloat(maximum); ok {
			if data > max {
				errors = append(errors, ValidationError{
					Path:       path,
					Constraint: "maximum",
					Actual:     data,
					Expected:   max,
				})
			}
		}
	}

	return errors
}

func (ctx *context) validateArray(path string, schema map[string]interface{}, data []interface{}) []ValidationError {
	errors := []ValidationError{}

	if items, exists := schema["items"]; exists {
		for i, item := range data {
			itemPath := fmt.Sprintf("%s/%d", path, i)
			errors = append(errors, ctx.validate(itemPath, items, item)...)
		}
	}

	if minItems, exists := schema["minItems"]; exists {
		if min, ok := toInt(minItems); ok {
			if len(data) < min {
				errors = append(errors, ValidationError{
					Path:       path,
					Constraint: "minItems",
					Actual:     len(data),
					Expected:   min,
				})
			}
		}
	}

	if maxItems, exists := schema["maxItems"]; exists {
		if max, ok := toInt(maxItems); ok {
			if len(data) > max {
				errors = append(errors, ValidationError{
					Path:       path,
					Constraint: "maxItems",
					Actual:     len(data),
					Expected:   max,
				})
			}
		}
	}

	return errors
}

func (ctx *context) validateObject(path string, schema map[string]interface{}, data map[string]interface{}) []ValidationError {
	errors := []ValidationError{}

	if required, exists := schema["required"]; exists {
		if reqArr, ok := required.([]interface{}); ok {
			for _, r := range reqArr {
				if reqStr, ok := r.(string); ok {
					if _, exists := data[reqStr]; !exists {
						reqPath := fmt.Sprintf("%s/%s", path, escapeJSONPointer(reqStr))
						errors = append(errors, ValidationError{
							Path:       reqPath,
							Constraint: "required",
							Actual:     nil,
							Expected:   reqStr,
						})
					}
				}
			}
		}
	}

	if properties, exists := schema["properties"]; exists {
		if propsMap, ok := properties.(map[string]interface{}); ok {
			for propName, propSchema := range propsMap {
				if propValue, exists := data[propName]; exists {
					propPath := fmt.Sprintf("%s/%s", path, escapeJSONPointer(propName))
					errors = append(errors, ctx.validate(propPath, propSchema, propValue)...)
				}
			}
		}
	}

	additionalProps := true
	if ap, exists := schema["additionalProperties"]; exists {
		if b, ok := ap.(bool); ok {
			additionalProps = b
		}
	}

	if !additionalProps {
		allowedProps := map[string]bool{}
		if properties, exists := schema["properties"]; exists {
			if propsMap, ok := properties.(map[string]interface{}); ok {
				for k := range propsMap {
					allowedProps[k] = true
				}
			}
		}

		for propName := range data {
			if !allowedProps[propName] {
				propPath := fmt.Sprintf("%s/%s", path, escapeJSONPointer(propName))
				errors = append(errors, ValidationError{
					Path:       propPath,
					Constraint: "additionalProperties",
					Actual:     data[propName],
					Expected:   "not allowed",
				})
			}
		}
	}

	return errors
}

func (ctx *context) validateAllOf(path string, allOf interface{}, data interface{}) []ValidationError {
	errors := []ValidationError{}
	schemas, ok := allOf.([]interface{})
	if !ok {
		return errors
	}

	for i, s := range schemas {
		schemaPath := fmt.Sprintf("%s/allOf[%d]", path, i)
		errors = append(errors, ctx.validate(schemaPath, s, data)...)
	}

	return errors
}

func (ctx *context) validateAnyOf(path string, anyOf interface{}, data interface{}) []ValidationError {
	schemas, ok := anyOf.([]interface{})
	if !ok {
		return []ValidationError{}
	}

	for _, s := range schemas {
		subErrors := ctx.validate(path, s, data)
		if len(subErrors) == 0 {
			return []ValidationError{}
		}
	}

	return []ValidationError{
		{
			Path:       path,
			Constraint: "anyOf",
			Actual:     data,
			Expected:   "at least one schema should match",
		},
	}
}

func (ctx *context) validateOneOf(path string, oneOf interface{}, data interface{}) []ValidationError {
	schemas, ok := oneOf.([]interface{})
	if !ok {
		return []ValidationError{}
	}

	validCount := 0
	for _, s := range schemas {
		subErrors := ctx.validate(path, s, data)
		if len(subErrors) == 0 {
			validCount++
		}
	}

	if validCount == 1 {
		return []ValidationError{}
	}

	return []ValidationError{
		{
			Path:       path,
			Constraint: "oneOf",
			Actual:     validCount,
			Expected:   "exactly one schema should match",
		},
	}
}

func (ctx *context) validateNot(path string, not interface{}, data interface{}) []ValidationError {
	subErrors := ctx.validate(path, not, data)
	if len(subErrors) > 0 {
		return []ValidationError{}
	}

	return []ValidationError{
		{
			Path:       path,
			Constraint: "not",
			Actual:     data,
			Expected:   "should not match schema",
		},
	}
}

func (ctx *context) resolveRef(ref string) (interface{}, error) {
	if !strings.HasPrefix(ref, "#") {
		return nil, fmt.Errorf("unsupported $ref: %s", ref)
	}

	refPath := strings.TrimPrefix(ref, "#")

	if ctx.visited[ref] {
		return nil, fmt.Errorf("circular $ref detected: %s", ref)
	}
	ctx.visited[ref] = true
	defer delete(ctx.visited, ref)

	if refPath == "" {
		return ctx.root, nil
	}

	parts := strings.Split(refPath, "/")
	if parts[0] != "" {
		return nil, fmt.Errorf("invalid JSON Pointer: %s", refPath)
	}
	parts = parts[1:]

	current := ctx.root
	for _, part := range parts {
		part = unescapeJSONPointer(part)
		switch c := current.(type) {
		case map[string]interface{}:
			if v, exists := c[part]; exists {
				current = v
			} else {
				return nil, fmt.Errorf("$ref path not found: %s", ref)
			}
		case []interface{}:
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(c) {
				return nil, fmt.Errorf("$ref path not found: %s", ref)
			}
			current = c[idx]
		default:
			return nil, fmt.Errorf("$ref path not found: %s", ref)
		}
	}

	return current, nil
}

func escapeJSONPointer(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	s = strings.ReplaceAll(s, "/", "~1")
	return s
}

func unescapeJSONPointer(s string) string {
	s = strings.ReplaceAll(s, "~1", "/")
	s = strings.ReplaceAll(s, "~0", "~")
	return s
}

func toInt(v interface{}) (int, bool) {
	switch x := v.(type) {
	case float64:
		return int(x), true
	case int:
		return x, true
	case int64:
		return int(x), true
	default:
		return 0, false
	}
}

func toFloat(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	default:
		return 0, false
	}
}
