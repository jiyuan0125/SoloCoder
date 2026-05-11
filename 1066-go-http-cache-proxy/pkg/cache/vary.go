package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"strings"
)

func ParseVary(header http.Header) []string {
	value := header.Get("Vary")
	if value == "" {
		return nil
	}

	var fields []string
	for _, field := range splitVaryFields(value) {
		field = strings.TrimSpace(field)
		if field != "" {
			fields = append(fields, http.CanonicalHeaderKey(field))
		}
	}

	return fields
}

func splitVaryFields(value string) []string {
	var fields []string
	var current strings.Builder

	for i := 0; i < len(value); i++ {
		c := value[i]
		if c == ',' {
			if current.Len() > 0 {
				fields = append(fields, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteByte(c)
	}

	if current.Len() > 0 {
		fields = append(fields, current.String())
	}

	return fields
}

func GenerateVaryKey(requestHeader http.Header, varyFields []string) string {
	if len(varyFields) == 0 {
		return ""
	}

	var values []string
	for _, field := range varyFields {
		value := strings.Join(requestHeader.Values(field), ",")
		values = append(values, field+":"+value)
	}

	sort.Strings(values)

	h := sha256.New()
	h.Write([]byte(strings.Join(values, "|")))
	return hex.EncodeToString(h.Sum(nil))
}

func BuildCacheKey(method, url string) string {
	return method + ":" + url
}

func MatchesVary(cachedVaryHeaders, cachedRespHeader, requestHeader http.Header) bool {
	varyFields := ParseVary(cachedRespHeader)
	if len(varyFields) == 0 {
		return true
	}

	cachedVaryKey := GenerateVaryKey(cachedVaryHeaders, varyFields)
	requestVaryKey := GenerateVaryKey(requestHeader, varyFields)

	return cachedVaryKey == requestVaryKey
}
