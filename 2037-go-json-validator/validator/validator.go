package validator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/xeipuuv/gojsonschema"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationResult struct {
	Valid  bool               `json:"valid"`
	Errors []ValidationError  `json:"errors"`
}

type StreamValidator struct {
	schemaValidator *gojsonschema.Schema
	mu              sync.Mutex
}

func NewStreamValidator(schemaJSON []byte) (*StreamValidator, error) {
	schemaLoader := gojsonschema.NewBytesLoader(schemaJSON)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, parseSchemaError(err)
	}
	return &StreamValidator{
		schemaValidator: schema,
	}, nil
}

func parseSchemaError(err error) error {
	return fmt.Errorf("schema 格式不合法: %v", err)
}

func (sv *StreamValidator) ValidateStream(reader io.Reader) (*ValidationResult, error) {
	sv.mu.Lock()
	defer sv.mu.Unlock()

	decoder := json.NewDecoder(reader)
	decoder.UseNumber()

	var data interface{}
	if err := decoder.Decode(&data); err != nil {
		return nil, parseJSONError(err)
	}

	dataLoader := gojsonschema.NewGoLoader(data)
	result, err := sv.schemaValidator.Validate(dataLoader)
	if err != nil {
		return nil, fmt.Errorf("校验过程出错: %v", err)
	}

	validationResult := &ValidationResult{
		Valid:  result.Valid(),
		Errors: make([]ValidationError, 0),
	}

	if !result.Valid() {
		for _, err := range result.Errors() {
			field := convertPointerToDotNotation(err.Field())
			validationResult.Errors = append(validationResult.Errors, ValidationError{
				Field:   field,
				Message: formatErrorMessage(err),
			})
		}
	}

	return validationResult, nil
}

func parseJSONError(err error) error {
	if syntaxErr, ok := err.(*json.SyntaxError); ok {
		return fmt.Errorf("JSON 语法错误，位置 %d: %v", syntaxErr.Offset, syntaxErr)
	}
	if unmarshalErr, ok := err.(*json.UnmarshalTypeError); ok {
		return fmt.Errorf("JSON 类型错误，位置 %d: 期望 %s 但得到 %s", unmarshalErr.Offset, unmarshalErr.Type, unmarshalErr.Value)
	}
	return fmt.Errorf("JSON 解析错误: %v", err)
}

func convertPointerToDotNotation(pointer string) string {
	if pointer == "" {
		return "(root)"
	}
	if !strings.HasPrefix(pointer, "/") {
		return pointer
	}

	parts := strings.Split(pointer[1:], "/")
	var result strings.Builder
	
	for i, part := range parts {
		part = unescapePointerPart(part)
		if isArrayIndex(part) {
			result.WriteString(fmt.Sprintf("[%s]", part))
		} else {
			if i > 0 && result.Len() > 0 {
				result.WriteString(".")
			}
			result.WriteString(part)
		}
	}
	
	return result.String()
}

func unescapePointerPart(part string) string {
	part = strings.ReplaceAll(part, "~1", "/")
	part = strings.ReplaceAll(part, "~0", "~")
	return part
}

func isArrayIndex(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func formatErrorMessage(err gojsonschema.ResultError) string {
	msg := err.Description()
	if strings.Contains(msg, "email") {
		return "不是合法邮箱格式"
	}
	if strings.Contains(msg, "pattern") {
		return "格式不正确"
	}
	if strings.Contains(msg, "required") {
		return "缺少必填字段"
	}
	if strings.Contains(msg, "type") {
		return fmt.Sprintf("类型错误，期望 %s", err.Details()["expected"])
	}
	if strings.Contains(msg, "minimum") {
		return fmt.Sprintf("值小于最小值 %v", err.Details()["minimum"])
	}
	if strings.Contains(msg, "maximum") {
		return fmt.Sprintf("值大于最大值 %v", err.Details()["maximum"])
	}
	return msg
}

func ValidateJSONBytes(schemaJSON []byte, dataJSON []byte) (*ValidationResult, error) {
	validator, err := NewStreamValidator(schemaJSON)
	if err != nil {
		return nil, err
	}
	return validator.ValidateStream(bytes.NewReader(dataJSON))
}
