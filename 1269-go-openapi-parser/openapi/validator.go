package openapi

import (
	"fmt"
	"regexp"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Field, e.Message)
}

type ValidationResult struct {
	Valid  bool
	Errors []*ValidationError
}

func Validate(spec *OpenAPI) *ValidationResult {
	result := &ValidationResult{Valid: true}
	errors := make([]*ValidationError, 0)

	if err := validateOpenAPIVersion(spec.OpenAPI); err != nil {
		errors = append(errors, err)
		result.Valid = false
	}

	if err := validateInfo(spec.Info); err != nil {
		errors = append(errors, err...)
		result.Valid = false
	}

	if len(spec.Paths) == 0 {
		errors = append(errors, &ValidationError{
			Field:   "paths",
			Message: "paths is required and must not be empty",
		})
		result.Valid = false
	}

	operationIDs := make(map[string]string)
	pathParamPattern := regexp.MustCompile(`\{([^}]+)\}`)

	for path, pathItem := range spec.Paths {
		if !strings.HasPrefix(path, "/") {
			errors = append(errors, &ValidationError{
				Field:   fmt.Sprintf("paths[%s]", path),
				Message: "path must start with /",
			})
			result.Valid = false
		}

		pathParams := pathParamPattern.FindAllStringSubmatch(path, -1)
		requiredPathParamNames := make(map[string]bool)
		for _, match := range pathParams {
			requiredPathParamNames[match[1]] = true
		}

		pathItemParamNames := make(map[string]bool)
		for _, p := range pathItem.Parameters {
			if p.Ref != "" {
				continue
			}
			if p.Param != nil && p.Param.In == "path" {
				pathItemParamNames[p.Param.Name] = true
			}
		}

		ops := map[string]*Operation{
			"get":     pathItem.Get,
			"post":    pathItem.Post,
			"put":     pathItem.Put,
			"delete":  pathItem.Delete,
			"patch":   pathItem.Patch,
			"options": pathItem.Options,
			"head":    pathItem.Head,
			"trace":   pathItem.Trace,
		}

		for method, op := range ops {
			if op == nil {
				continue
			}

			if op.OperationID != "" {
				if existingPath, exists := operationIDs[op.OperationID]; exists {
					errors = append(errors, &ValidationError{
						Field:   fmt.Sprintf("paths[%s].%s.operationId", path, method),
						Message: fmt.Sprintf("duplicate operationId '%s', already used in %s", op.OperationID, existingPath),
					})
					result.Valid = false
				} else {
					operationIDs[op.OperationID] = fmt.Sprintf("paths[%s].%s", path, method)
				}
			}

			if len(op.Responses) == 0 {
				errors = append(errors, &ValidationError{
					Field:   fmt.Sprintf("paths[%s].%s.responses", path, method),
					Message: "responses must contain at least one status code",
				})
				result.Valid = false
			}

			definedPathParams := make(map[string]bool)
			for name := range pathItemParamNames {
				definedPathParams[name] = true
			}

			for _, p := range op.Parameters {
				if p.Ref != "" {
					continue
				}
				if p.Param != nil && p.Param.In == "path" {
					definedPathParams[p.Param.Name] = true
				}
			}

			for name := range requiredPathParamNames {
				if !definedPathParams[name] {
					errors = append(errors, &ValidationError{
						Field:   fmt.Sprintf("paths[%s].%s.parameters", path, method),
						Message: fmt.Sprintf("path parameter '%s' defined in path but not in operation parameters", name),
					})
					result.Valid = false
				}
			}
		}
	}

	refs := collectAllReferences(spec)
	for ref := range refs {
		if !strings.HasPrefix(ref, "#/") {
			continue
		}
		if err := validateReference(spec, ref); err != nil {
			errors = append(errors, err)
			result.Valid = false
		}
	}

	result.Errors = errors
	return result
}

func validateOpenAPIVersion(version string) *ValidationError {
	validVersions := map[string]bool{
		"3.0.0": true,
		"3.0.1": true,
		"3.0.2": true,
		"3.0.3": true,
	}
	if !validVersions[version] {
		return &ValidationError{
			Field:   "openapi",
			Message: fmt.Sprintf("openapi version must be one of 3.0.0, 3.0.1, 3.0.2, 3.0.3, got '%s'", version),
		}
	}
	return nil
}

func validateInfo(info Info) []*ValidationError {
	errors := make([]*ValidationError, 0)
	if info.Title == "" {
		errors = append(errors, &ValidationError{
			Field:   "info.title",
			Message: "info.title is required",
		})
	}
	if info.Version == "" {
		errors = append(errors, &ValidationError{
			Field:   "info.version",
			Message: "info.version is required",
		})
	}
	return errors
}

func validateReference(spec *OpenAPI, ref string) *ValidationError {
	parts := strings.Split(ref, "/")
	if len(parts) < 3 || parts[0] != "#" {
		return &ValidationError{
			Field:   ref,
			Message: fmt.Sprintf("invalid reference format: %s", ref),
		}
	}

	if spec.Raw == nil {
		return &ValidationError{
			Field:   ref,
			Message: "cannot validate reference: raw data not available",
		}
	}

	current := spec.Raw
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		next, ok := current[part]
		if !ok {
			return &ValidationError{
				Field:   ref,
				Message: fmt.Sprintf("reference target not found: %s (missing segment '%s')", ref, part),
			}
		}
		if i == len(parts)-1 {
			return nil
		}
		current, ok = next.(map[string]interface{})
		if !ok {
			return &ValidationError{
				Field:   ref,
				Message: fmt.Sprintf("reference path '%s' is invalid at segment '%s'", ref, part),
			}
		}
	}
	return nil
}

func collectAllReferences(spec *OpenAPI) map[string]bool {
	refs := make(map[string]bool)

	if spec.Raw == nil {
		return refs
	}

	collectRefsFromMap(spec.Raw, refs)
	return refs
}

func collectRefsFromMap(m map[string]interface{}, refs map[string]bool) {
	for key, value := range m {
		if key == "$ref" {
			if str, ok := value.(string); ok {
				refs[str] = true
			}
			continue
		}
		switch v := value.(type) {
		case map[string]interface{}:
			collectRefsFromMap(v, refs)
		case []interface{}:
			collectRefsFromSlice(v, refs)
		}
	}
}

func collectRefsFromSlice(s []interface{}, refs map[string]bool) {
	for _, item := range s {
		switch v := item.(type) {
		case map[string]interface{}:
			collectRefsFromMap(v, refs)
		case []interface{}:
			collectRefsFromSlice(v, refs)
		}
	}
}
