package anchor

import (
	"regexp"
	"strings"
)

const (
	aliasMarkerPrefix = "__YAML_ALIAS__:"
	aliasMarkerSuffix = "__"
)

var (
	aliasPattern = regexp.MustCompile(`\*([a-zA-Z_][a-zA-Z0-9_]*)`)
)

func PreprocessYAML(yamlText string) string {
	result := aliasPattern.ReplaceAllStringFunc(yamlText, func(match string) string {
		aliasName := match[1:]
		return `"` + aliasMarkerPrefix + aliasName + aliasMarkerSuffix + `"`
	})
	return result
}

func IsAliasMarker(value string) (bool, string) {
	if strings.HasPrefix(value, aliasMarkerPrefix) && strings.HasSuffix(value, aliasMarkerSuffix) {
		aliasName := value[len(aliasMarkerPrefix) : len(value)-len(aliasMarkerSuffix)]
		return true, aliasName
	}
	return false, ""
}
