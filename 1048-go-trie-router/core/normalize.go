package core

import (
	"strings"
)

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}

	parts := strings.Split(path, "/")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}

	if len(result) == 0 {
		return "/"
	}

	return "/" + strings.Join(result, "/")
}

func splitPath(path string) []string {
	if path == "/" {
		return []string{}
	}
	path = strings.TrimPrefix(path, "/")
	return strings.Split(path, "/")
}

func normalizeSegment(seg string, caseSensitive bool) string {
	if caseSensitive {
		return seg
	}
	return strings.ToLower(seg)
}

func isParamSegment(seg string) bool {
	return strings.HasPrefix(seg, ":")
}

func isWildcardSegment(seg string) bool {
	return strings.HasPrefix(seg, "*")
}

func extractParamName(seg string) string {
	if strings.HasPrefix(seg, "*") {
		return strings.TrimPrefix(seg, "*")
	}
	if strings.HasPrefix(seg, ":") {
		seg = strings.TrimPrefix(seg, ":")
		if idx := strings.Index(seg, "("); idx != -1 {
			return seg[:idx]
		}
		return seg
	}
	return seg
}

func parseParamSegment(seg string) (string, ParamType) {
	seg = strings.TrimPrefix(seg, ":")
	paramType := ParamTypeString

	if idx := strings.Index(seg, "("); idx != -1 && strings.HasSuffix(seg, ")") {
		typeStr := seg[idx+1 : len(seg)-1]
		switch strings.ToLower(typeStr) {
		case "int":
			paramType = ParamTypeInt
		case "bool":
			paramType = ParamTypeBool
		case "string":
			paramType = ParamTypeString
		}
		seg = seg[:idx]
	}

	return seg, paramType
}
