package middleware

import (
	"context"
	"net/http"
	"sort"
	"strings"

	"traffictag/internal/matcher"
	"traffictag/internal/tagstore"
)

type contextKey string

const (
	ContextKeyTags  contextKey = "traffic-tags"
	HeaderTrafficTags = "X-Traffic-Tags"
)

func GetTagsFromContext(ctx context.Context) map[string]string {
	if v := ctx.Value(ContextKeyTags); v != nil {
		if tags, ok := v.(map[string]string); ok {
			return tags
		}
	}
	return nil
}

func WithTags(ctx context.Context, tags map[string]string) context.Context {
	if len(tags) == 0 {
		return ctx
	}
	return context.WithValue(ctx, ContextKeyTags, tags)
}

func ParseTagsFromHeader(header string) map[string]string {
	if header == "" {
		return nil
	}
	parts := strings.Split(header, ",")
	tags := make(map[string]string, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		idx := strings.Index(part, "=")
		if idx > 0 {
			k := strings.TrimSpace(part[:idx])
			v := strings.TrimSpace(part[idx+1:])
			if k != "" {
				tags[k] = v
			}
		}
	}
	if len(tags) == 0 {
		return nil
	}
	return tags
}

func TagsToHeader(tags map[string]string) string {
	if len(tags) == 0 {
		return ""
	}
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+tags[k])
	}
	return strings.Join(parts, ",")
}

func mergeTags(a, b map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}

func Tagging(ruleStore *matcher.RuleStore, tagStore *tagstore.TagStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			var tags map[string]string

			if existingHeader := r.Header.Get(HeaderTrafficTags); existingHeader != "" {
				existing := ParseTagsFromHeader(existingHeader)
				if existing != nil {
					tags = existing
				}
			}

			matched := ruleStore.MatchAll(r)
			if len(matched) > 0 {
				if tags == nil {
					tags = matched
				} else {
					tags = mergeTags(tags, matched)
				}
			}

			if tags != nil && len(tags) > 0 {
				r = r.WithContext(WithTags(ctx, tags))
				w.Header().Set(HeaderTrafficTags, TagsToHeader(tags))
				tagStore.Increment(tags)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func PropagateClient(cli *http.Client, ruleStore *matcher.RuleStore) *http.Client {
	if cli == nil {
		cli = http.DefaultClient
	}

	transport := cli.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	cliCopy := *cli
	cliCopy.Transport = &propagatingTransport{
		base:      transport,
		ruleStore: ruleStore,
	}
	return &cliCopy
}

type propagatingTransport struct {
	base      http.RoundTripper
	ruleStore *matcher.RuleStore
}

func (t *propagatingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqCopy := req.Clone(req.Context())

	if tags := GetTagsFromContext(req.Context()); len(tags) > 0 {
		if reqCopy.Header.Get(HeaderTrafficTags) == "" {
			reqCopy.Header.Set(HeaderTrafficTags, TagsToHeader(tags))
		}
	}

	if t.ruleStore != nil && reqCopy.Header.Get(HeaderTrafficTags) == "" {
		matched := t.ruleStore.MatchAll(reqCopy)
		if len(matched) > 0 {
			reqCopy.Header.Set(HeaderTrafficTags, TagsToHeader(matched))
		}
	}

	return t.base.RoundTrip(reqCopy)
}

func DefaultClient(ruleStore *matcher.RuleStore) *http.Client {
	return PropagateClient(nil, ruleStore)
}
