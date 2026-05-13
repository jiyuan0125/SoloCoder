package proxy

import (
	"net/http"
)

func FilterHeaders(headers http.Header) http.Header {
	filtered := make(http.Header)

	for key, values := range headers {
		newValues := make([]string, 0, len(values))
		for _, v := range values {
			filteredValue := filterNonASCII(v)
			if len(filteredValue) > 0 {
				newValues = append(newValues, filteredValue)
			}
		}
		if len(newValues) > 0 {
			filtered[key] = newValues
		}
	}

	return filtered
}

func filterNonASCII(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < 128 {
			result = append(result, b)
		}
	}
	return string(result)
}
