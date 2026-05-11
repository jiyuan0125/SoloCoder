package codegen

import (
	"fmt"
	"strings"
)

var validGRPCCodes = map[string]bool{
	"OK":                  true,
	"CANCELLED":           true,
	"UNKNOWN":             true,
	"INVALID_ARGUMENT":    true,
	"DEADLINE_EXCEEDED":   true,
	"NOT_FOUND":           true,
	"ALREADY_EXISTS":      true,
	"PERMISSION_DENIED":   true,
	"RESOURCE_EXHAUSTED":  true,
	"FAILED_PRECONDITION": true,
	"ABORTED":             true,
	"OUT_OF_RANGE":        true,
	"UNIMPLEMENTED":       true,
	"INTERNAL":            true,
	"UNAVAILABLE":         true,
	"DATA_LOSS":           true,
	"UNAUTHENTICATED":     true,
}

var commonErrorCodes = []string{"INTERNAL_ERROR", "NOT_FOUND", "INVALID_ARGUMENT", "PERMISSION_DENIED", "UNAUTHENTICATED"}

func Validate(def *YAMLDefinition) ([]ValidatedError, []string, error) {
	var validated []ValidatedError
	var warnings []string

	codes := make(map[int]string)
	names := make(map[string]bool)

	for _, spec := range def.Errors {
		if spec.Name == "" {
			return nil, nil, fmt.Errorf("error name cannot be empty")
		}
		if names[spec.Name] {
			return nil, nil, fmt.Errorf("duplicate error name: %s", spec.Name)
		}
		names[spec.Name] = true

		if existing, ok := codes[spec.Code]; ok {
			if existing != spec.Message {
				return nil, nil, fmt.Errorf("duplicate code %d with different messages: existing=%q, new=%q", spec.Code, existing, spec.Message)
			}
		}
		codes[spec.Code] = spec.Message

		if spec.HTTPStatus < 100 || spec.HTTPStatus > 599 {
			return nil, nil, fmt.Errorf("invalid http_status %d for %s (must be 100-599)", spec.HTTPStatus, spec.Name)
		}

		grpcUpper := strings.ToUpper(spec.GRPCCode)
		if !validGRPCCodes[grpcUpper] {
			return nil, nil, fmt.Errorf("invalid grpc_code %q for %s", spec.GRPCCode, spec.Name)
		}

		category := strings.ToUpper(spec.Category)
		if category == "" {
			return nil, nil, fmt.Errorf("category cannot be empty for %s", spec.Name)
		}

		validated = append(validated, ValidatedError{
			Name:       spec.Name,
			Code:       spec.Code,
			Message:    spec.Message,
			HTTPStatus: spec.HTTPStatus,
			GRPCCode:   grpcUpper,
			Category:   category,
			Deprecated: spec.Deprecated,
		})
	}

	definedNames := make(map[string]bool)
	for _, v := range validated {
		definedNames[v.Name] = true
	}

	for _, common := range commonErrorCodes {
		if !definedNames[common] {
			warnings = append(warnings, fmt.Sprintf("suggestion: consider defining common error code %q", common))
		}
	}

	return validated, warnings, nil
}
