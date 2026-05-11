package openapi

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ResolveReference(spec *OpenAPI, ref string) (interface{}, error) {
	if !strings.HasPrefix(ref, "#/") {
		return nil, fmt.Errorf("only internal references are supported: %s", ref)
	}

	parts := strings.Split(ref, "/")
	if len(parts) < 3 || parts[0] != "#" {
		return nil, fmt.Errorf("invalid reference format: %s", ref)
	}

	if spec.Raw == nil {
		return nil, fmt.Errorf("raw data not available")
	}

	current := spec.Raw
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		next, ok := current[part]
		if !ok {
			return nil, fmt.Errorf("reference target not found: %s", ref)
		}
		if i == len(parts)-1 {
			return resolveAllRefs(spec, next), nil
		}
		current, ok = next.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("reference path is invalid: %s", ref)
		}
	}
	return nil, fmt.Errorf("reference not found: %s", ref)
}

func ResolveAll(spec *OpenAPI) map[string]interface{} {
	if spec.Raw == nil {
		return nil
	}
	return resolveAllRefs(spec, spec.Raw).(map[string]interface{})
}

func resolveAllRefs(spec *OpenAPI, value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		if ref, ok := v["$ref"]; ok {
			if refStr, ok := ref.(string); ok && strings.HasPrefix(refStr, "#/") {
				resolved, err := ResolveReference(spec, refStr)
				if err == nil {
					return resolved
				}
			}
		}
		result := make(map[string]interface{})
		for key, val := range v {
			result[key] = resolveAllRefs(spec, val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = resolveAllRefs(spec, item)
		}
		return result
	default:
		return value
	}
}

func GetAllPaths(spec *OpenAPI) []PathInfo {
	paths := make([]PathInfo, 0)
	for path, pathItem := range spec.Paths {
		methods := make([]string, 0)
		if pathItem.Get != nil {
			methods = append(methods, "GET")
		}
		if pathItem.Post != nil {
			methods = append(methods, "POST")
		}
		if pathItem.Put != nil {
			methods = append(methods, "PUT")
		}
		if pathItem.Delete != nil {
			methods = append(methods, "DELETE")
		}
		if pathItem.Patch != nil {
			methods = append(methods, "PATCH")
		}
		if pathItem.Options != nil {
			methods = append(methods, "OPTIONS")
		}
		if pathItem.Head != nil {
			methods = append(methods, "HEAD")
		}
		if pathItem.Trace != nil {
			methods = append(methods, "TRACE")
		}
		paths = append(paths, PathInfo{
			Path:    path,
			Methods: methods,
		})
	}
	return paths
}

type PathInfo struct {
	Path    string   `json:"path"`
	Methods []string `json:"methods"`
}

func ToJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
