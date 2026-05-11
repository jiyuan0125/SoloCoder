package codegen

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseYAML(data []byte) (*YAMLDefinition, error) {
	var def YAMLDefinition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, err
	}
	return &def, nil
}

func InferPackageName(yamlPath string) (string, error) {
	absPath, err := filepath.Abs(yamlPath)
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(absPath)
	base := filepath.Base(dir)
	return sanitizePackageName(base), nil
}

func sanitizePackageName(name string) string {
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}

func GenerateFromYAMLFile(yamlPath, packageName string) (*GenerateResult, error) {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, err
	}
	return GenerateFromYAML(data, packageName)
}

func GenerateFromYAML(data []byte, packageName string) (*GenerateResult, error) {
	def, err := ParseYAML(data)
	if err != nil {
		return nil, err
	}
	validated, warnings, err := Validate(def)
	if err != nil {
		return nil, err
	}
	code, err := Generate(validated, packageName)
	if err != nil {
		return &GenerateResult{Code: code, Warnings: warnings}, err
	}
	return &GenerateResult{Code: code, Warnings: warnings}, nil
}
