package tagparser

import "strings"

func ParseJSONTag(tag string) *JSONTag {
	if tag == "" {
		return nil
	}

	parts := splitTagOptions(tag)
	if len(parts) == 0 {
		return nil
	}

	result := &JSONTag{}
	name := parts[0]
	if name == "-" {
		result.Name = "-"
	} else {
		result.Name = name
	}

	for _, opt := range parts[1:] {
		switch opt {
		case "omitempty":
			result.OmitEmpty = true
		case "string":
			result.String = true
		}
	}

	return result
}

func splitTagOptions(tag string) []string {
	var result []string
	var current strings.Builder
	inQuote := false

	for i := 0; i < len(tag); i++ {
		c := tag[i]
		if c == ',' && !inQuote {
			result = append(result, current.String())
			current.Reset()
			continue
		}
		current.WriteByte(c)
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}
