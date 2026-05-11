package tagparser

import (
	"strconv"
	"strings"
)

func ParseValidateTag(tag string) *ValidateTag {
	if tag == "" {
		return nil
	}

	parts := splitTagOptions(tag)
	if len(parts) == 0 {
		return nil
	}

	result := &ValidateTag{}

	for _, opt := range parts {
		if opt == "required" {
			result.Required = true
			continue
		}
		if opt == "email" {
			result.Email = true
			continue
		}
		if strings.HasPrefix(opt, "min=") {
			val, err := strconv.ParseFloat(strings.TrimPrefix(opt, "min="), 64)
			if err == nil {
				result.Min = &val
			}
			continue
		}
		if strings.HasPrefix(opt, "max=") {
			val, err := strconv.ParseFloat(strings.TrimPrefix(opt, "max="), 64)
			if err == nil {
				result.Max = &val
			}
			continue
		}
		if strings.HasPrefix(opt, "regex=") {
			result.Regex = strings.TrimPrefix(opt, "regex=")
			continue
		}
	}

	return result
}
