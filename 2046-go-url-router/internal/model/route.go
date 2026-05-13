package model

import (
	"strings"
	"time"
)

type Route struct {
	ID        int64     `json:"id"`
	Path      string    `json:"path"`
	Method    string    `json:"method"`
	Target    string    `json:"target"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *Route) IsWildcard() bool {
	return strings.HasSuffix(r.Path, "/*")
}

func (r *Route) WildcardPrefix() string {
	if !r.IsWildcard() {
		return r.Path
	}
	return strings.TrimSuffix(r.Path, "*")
}

func (r *Route) Priority() int {
	if r.IsWildcard() {
		return len(r.WildcardPrefix())
	}
	return len(r.Path) + 1000
}
