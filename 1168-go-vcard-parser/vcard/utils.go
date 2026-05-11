package vcard

import "strings"

func unescapeValue(s string) string {
	var builder strings.Builder
	i := 0
	for i < len(s) {
		if i+1 < len(s) && s[i] == '\\' {
			switch s[i+1] {
			case ';':
				builder.WriteByte(';')
				i += 2
			case ',':
				builder.WriteByte(',')
				i += 2
			case ':':
				builder.WriteByte(':')
				i += 2
			case '\\':
				builder.WriteByte('\\')
				i += 2
			case 'n', 'N':
				builder.WriteByte('\n')
				i += 2
			case 'r', 'R':
				builder.WriteByte('\r')
				i += 2
			default:
				builder.WriteByte(s[i])
				i++
			}
		} else {
			builder.WriteByte(s[i])
			i++
		}
	}
	return builder.String()
}

func splitByChar(s string, sep byte, preserveEscaped bool) []string {
	var result []string
	var builder strings.Builder
	i := 0
	for i < len(s) {
		if i+1 < len(s) && s[i] == '\\' {
			if preserveEscaped {
				builder.WriteByte(s[i])
				builder.WriteByte(s[i+1])
			} else {
				builder.WriteByte(s[i])
				builder.WriteByte(s[i+1])
			}
			i += 2
			continue
		}
		if s[i] == sep {
			result = append(result, builder.String())
			builder.Reset()
			i++
			continue
		}
		builder.WriteByte(s[i])
		i++
	}
	result = append(result, builder.String())
	return result
}

func splitPropertyValueWithEscaping(s string, sep byte) []string {
	var result []string
	var builder strings.Builder
	i := 0
	for i < len(s) {
		if i+1 < len(s) && s[i] == '\\' {
			if s[i+1] == ';' || s[i+1] == ',' || s[i+1] == ':' || s[i+1] == '\\' || s[i+1] == 'n' || s[i+1] == 'N' {
				builder.WriteByte(s[i])
				builder.WriteByte(s[i+1])
				i += 2
				continue
			}
		}
		if s[i] == sep {
			result = append(result, builder.String())
			builder.Reset()
			i++
			continue
		}
		builder.WriteByte(s[i])
		i++
	}
	result = append(result, builder.String())
	return result
}
