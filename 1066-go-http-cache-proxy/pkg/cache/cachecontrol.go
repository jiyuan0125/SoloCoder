package cache

import (
	"net/http"
	"strconv"
	"strings"
)

type CacheDirectives struct {
	NoStore         bool
	NoCache         bool
	Private         bool
	Public          bool
	MaxAge          int
	MaxStale        int
	MinFresh        int
	MustRevalidate  bool
	ProxyRevalidate bool
	SMaxAge         int
}

func ParseCacheControl(header http.Header) *CacheDirectives {
	cc := &CacheDirectives{
		MaxAge:   -1,
		SMaxAge:  -1,
		MaxStale: -1,
		MinFresh: -1,
	}

	values := header.Values("Cache-Control")
	for _, value := range values {
		directives := parseDirectives(value)
		for _, dir := range directives {
			key := strings.TrimSpace(strings.ToLower(dir))
			switch {
			case key == "no-store":
				cc.NoStore = true
			case key == "no-cache":
				cc.NoCache = true
			case key == "private":
				cc.Private = true
			case key == "public":
				cc.Public = true
			case key == "must-revalidate":
				cc.MustRevalidate = true
			case key == "proxy-revalidate":
				cc.ProxyRevalidate = true
			case strings.HasPrefix(key, "max-age="):
				if age, err := strconv.Atoi(key[8:]); err == nil {
					cc.MaxAge = age
				}
			case strings.HasPrefix(key, "s-maxage="):
				if age, err := strconv.Atoi(key[8:]); err == nil {
					cc.SMaxAge = age
				}
			case strings.HasPrefix(key, "max-stale"):
				if age, err := strconv.Atoi(key[9:]); err == nil {
					cc.MaxStale = age
				} else {
					cc.MaxStale = 0
				}
			case strings.HasPrefix(key, "min-fresh="):
				if age, err := strconv.Atoi(key[10:]); err == nil {
					cc.MinFresh = age
				}
			}
		}
	}

	return cc
}

func parseDirectives(value string) []string {
	var directives []string
	var current strings.Builder
	inQuote := false

	for i := 0; i < len(value); i++ {
		c := value[i]
		if c == '"' {
			inQuote = !inQuote
			continue
		}
		if c == ',' && !inQuote {
			directives = append(directives, current.String())
			current.Reset()
			continue
		}
		current.WriteByte(c)
	}

	if current.Len() > 0 {
		directives = append(directives, current.String())
	}

	return directives
}

func (cc *CacheDirectives) ShouldNotCache() bool {
	return cc.NoStore || cc.Private || (cc.MaxAge == 0 && !cc.NoCache)
}

func (cc *CacheDirectives) MustRevalidateOnUse() bool {
	return cc.NoCache || cc.MustRevalidate || cc.ProxyRevalidate
}

func (cc *CacheDirectives) GetMaxAge() int {
	if cc.SMaxAge >= 0 {
		return cc.SMaxAge
	}
	return cc.MaxAge
}
