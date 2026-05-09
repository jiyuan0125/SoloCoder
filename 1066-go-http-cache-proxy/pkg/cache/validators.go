package cache

import (
	"net/http"
	"time"
)

type ETag struct {
	Value string
	Weak  bool
}

func ParseETag(value string) *ETag {
	if value == "" {
		return nil
	}

	etag := &ETag{}
	value = http.CanonicalHeaderKey(value)

	if len(value) >= 3 && value[:2] == "W/" {
		etag.Weak = true
		value = value[2:]
	}

	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		etag.Value = value[1 : len(value)-1]
	} else {
		etag.Value = value
	}

	return etag
}

func (e *ETag) String() string {
	if e.Weak {
		return "W/\"" + e.Value + "\""
	}
	return "\"" + e.Value + "\""
}

func (e *ETag) Match(other *ETag, useStrong bool) bool {
	if e == nil || other == nil {
		return false
	}

	if e.Value != other.Value {
		return false
	}

	if useStrong {
		return !e.Weak && !other.Weak
	}

	return true
}

func ParseIfNoneMatch(header http.Header) []*ETag {
	value := header.Get("If-None-Match")
	if value == "" {
		return nil
	}

	var etags []*ETag
	for _, part := range splitHeaderValue(value) {
		if part == "*" {
			return []*ETag{}
		}
		if etag := ParseETag(part); etag != nil {
			etags = append(etags, etag)
		}
	}

	return etags
}

func ParseIfModifiedSince(header http.Header) (time.Time, bool) {
	value := header.Get("If-Modified-Since")
	if value == "" {
		return time.Time{}, false
	}

	t, err := http.ParseTime(value)
	if err != nil {
		return time.Time{}, false
	}

	return t, true
}

func ParseLastModified(header http.Header) (time.Time, bool) {
	value := header.Get("Last-Modified")
	if value == "" {
		return time.Time{}, false
	}

	t, err := http.ParseTime(value)
	if err != nil {
		return time.Time{}, false
	}

	return t, true
}

func ParseDate(header http.Header) (time.Time, bool) {
	value := header.Get("Date")
	if value == "" {
		return time.Time{}, false
	}

	t, err := http.ParseTime(value)
	if err != nil {
		return time.Time{}, false
	}

	return t, true
}

func ETagsMatch(cachedETag, requestETags []*ETag, useStrong bool) bool {
	if len(requestETags) == 0 {
		return true
	}

	for _, cached := range cachedETag {
		for _, req := range requestETags {
			if cached.Match(req, useStrong) {
				return true
			}
		}
	}

	return false
}

func splitHeaderValue(value string) []string {
	var parts []string
	var current []rune
	inQuote := false

	for _, r := range value {
		if r == '"' {
			inQuote = !inQuote
			current = append(current, r)
			continue
		}
		if r == ',' && !inQuote {
			if len(current) > 0 {
				parts = append(parts, trimQuotes(string(current)))
				current = nil
			}
			continue
		}
		current = append(current, r)
	}

	if len(current) > 0 {
		parts = append(parts, trimQuotes(string(current)))
	}

	return parts
}

func trimQuotes(s string) string {
	s = http.CanonicalHeaderKey(s)
	s = trimSpace(s)
	return s
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
