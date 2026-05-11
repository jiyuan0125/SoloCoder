package core

import (
	"bytes"
	"fmt"
	"go/format"
)

func GenerateTestCode(sourceCode string, fileName string) (*GeneratedTest, error) {
	parsed, err := ParseSource(sourceCode, fileName)
	if err != nil {
		return nil, err
	}

	valueGen := NewValueGenerator(parsed.Structs)
	testCaseGen := NewTestCaseGenerator(valueGen)
	gen := NewGenerator(testCaseGen)

	result, err := gen.GenerateTestFile(parsed, fileName)
	if err != nil {
		return nil, err
	}

	formattedCode, err := formatCode(result.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to format code: %w", err)
	}
	result.Code = formattedCode

	return result, nil
}

func GeneratePreview(sourceCode string, fileName string) (*Preview, error) {
	parsed, err := ParseSource(sourceCode, fileName)
	if err != nil {
		return nil, err
	}

	valueGen := NewValueGenerator(parsed.Structs)
	testCaseGen := NewTestCaseGenerator(valueGen)
	gen := NewGenerator(testCaseGen)

	return gen.GeneratePreview(parsed, fileName)
}

func formatCode(code string) (string, error) {
	formatted, err := format.Source([]byte(code))
	if err != nil {
		return code, err
	}

	var buf bytes.Buffer
	buf.Write(formatted)
	return buf.String(), nil
}
