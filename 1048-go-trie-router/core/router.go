package core

import (
	"sort"
	"strings"
	"sync"
	"time"

	"trie-router/common"
)

type Router struct {
	mu        sync.RWMutex
	tree      *node
	routes    []*common.RouteInfo
	config    RouterConfig
}

type RouterConfig struct {
	CaseSensitive bool
}

func NewRouter(config ...RouterConfig) *Router {
	cfg := RouterConfig{CaseSensitive: false}
	if len(config) > 0 {
		cfg = config[0]
	}
	return &Router{
		tree:   newRootNode(),
		routes: make([]*common.RouteInfo, 0),
		config: cfg,
	}
}

func (r *Router) Register(path, method, handlerName string) error {
	if path == "" {
		return ErrInvalidPath
	}
	if method == "" {
		return ErrInvalidMethod
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	normalizedPath := normalizePath(path)
	segments := splitPath(normalizedPath)

	current := r.tree
	for i, seg := range segments {
		isLast := i == len(segments)-1

		switch {
		case isWildcardSegment(seg):
			if !isLast {
				return ErrWildcardNotAtEnd
			}
			paramName := extractParamName(seg)
			if current.wildcard == nil {
				current.wildcard = newNode()
				current.wildcard.paramName = paramName
				current.wildcard.isWildcard = true
			}
			current = current.wildcard
		case isParamSegment(seg):
			paramName, paramType := parseParamSegment(seg)
			if current.param == nil {
				current.param = newNode()
				current.param.paramName = paramName
				current.param.paramType = paramType
				current.param.isParam = true
			} else {
				if current.param.paramName != paramName {
					return ErrParamNameConflict
				}
				if current.param.paramType != paramType {
					return ErrParamTypeConflict
				}
			}
			current = current.param
		default:
			key := normalizeSegment(seg, r.config.CaseSensitive)
			child, exists := current.children[key]
			if !exists {
				child = newNode()
				child.static = seg
				current.children[key] = child
			}
			current = child
		}
	}

	upperMethod := strings.ToUpper(method)
	if current.handlers == nil {
		current.handlers = make(map[string]*common.RouteInfo)
	}

	if _, exists := current.handlers[upperMethod]; exists {
		return ErrRouteExists
	}

	routeInfo := &common.RouteInfo{
		Path:        path,
		Method:      upperMethod,
		HandlerName: handlerName,
		RegisteredAt: time.Now().UnixNano(),
	}
	current.handlers[upperMethod] = routeInfo
	r.routes = append(r.routes, routeInfo)

	return nil
}

func (r *Router) Match(path, method string) *common.MatchResult {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normalizedPath := normalizePath(path)
	segments := splitPath(normalizedPath)
	upperMethod := strings.ToUpper(method)

	result := r.matchNode(r.tree, segments, upperMethod, make(map[string]string))
	if result == nil {
		return &common.MatchResult{
			Found:  false,
			Status: 404,
		}
	}
	return result
}

func (r *Router) matchNode(n *node, segments []string, method string, params map[string]string) *common.MatchResult {
	if len(segments) == 0 {
		if n.handlers != nil {
			if handler, ok := n.handlers[method]; ok {
				return &common.MatchResult{
					Found:  true,
					Route:  handler,
					Params: copyMap(params),
					Status: 200,
				}
			}
			allowed := make([]string, 0, len(n.handlers))
			for m := range n.handlers {
				allowed = append(allowed, m)
			}
			sort.Strings(allowed)
			return &common.MatchResult{
				Found:          false,
				AllowedMethods: allowed,
				Status:         405,
			}
		}
		return nil
	}

	seg := segments[0]
	remaining := segments[1:]
	key := normalizeSegment(seg, r.config.CaseSensitive)

	if child, ok := n.children[key]; ok {
		result := r.matchNode(child, remaining, method, params)
		if result != nil && (result.Found || result.IsMethodNotAllowed()) {
			return result
		}
	}

	if n.param != nil {
		if validateParamType(seg, n.param.paramType) {
			newParams := copyMap(params)
			newParams[n.param.paramName] = seg
			result := r.matchNode(n.param, remaining, method, newParams)
			if result != nil && (result.Found || result.IsMethodNotAllowed()) {
				return result
			}
		}
	}

	if n.wildcard != nil {
		newParams := copyMap(params)
		newParams[n.wildcard.paramName] = strings.Join(segments, "/")
		if n.wildcard.handlers != nil {
			if handler, ok := n.wildcard.handlers[method]; ok {
				return &common.MatchResult{
					Found:  true,
					Route:  handler,
					Params: newParams,
					Status: 200,
				}
			}
			allowed := make([]string, 0, len(n.wildcard.handlers))
			for m := range n.wildcard.handlers {
				allowed = append(allowed, m)
			}
			sort.Strings(allowed)
			return &common.MatchResult{
				Found:          false,
				AllowedMethods: allowed,
				Status:         405,
			}
		}
	}

	return nil
}

func (r *Router) List(sortBy string) []*common.RouteInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*common.RouteInfo, len(r.routes))
	copy(result, r.routes)

	switch sortBy {
	case "priority":
		sort.Slice(result, func(i, j int) bool {
			return getPriority(result[i].Path) < getPriority(result[j].Path)
		})
	case "path":
		sort.Slice(result, func(i, j int) bool {
			return result[i].Path < result[j].Path
		})
	default:
	}

	return result
}

func getPriority(path string) int {
	segments := splitPath(normalizePath(path))
	priority := 0
	for _, seg := range segments {
		switch {
		case isWildcardSegment(seg):
			priority += 100
		case isParamSegment(seg):
			priority += 10
		default:
			priority += 1
		}
	}
	return priority
}
