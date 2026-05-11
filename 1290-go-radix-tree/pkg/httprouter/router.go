package httprouter

import (
	"strings"

	"github.com/example/radix-router/pkg/radix"
)

type HTTPRoute struct {
	Path    string
	Handler string
	Method  string
}

type MatchResult struct {
	Route  HTTPRoute
	Params map[string]string
	Found  bool
}

type Router struct {
	trees map[string]*radix.Tree
}

func NewRouter() *Router {
	return &Router{
		trees: make(map[string]*radix.Tree),
	}
}

func (r *Router) AddRoute(method, path, handler string) error {
	method = strings.ToUpper(method)
	if r.trees[method] == nil {
		r.trees[method] = radix.NewTree()
	}

	route := HTTPRoute{
		Path:    path,
		Handler: handler,
		Method:  method,
	}

	r.trees[method].Insert(path, route)
	return nil
}

func (r *Router) DeleteRoute(method, path string) bool {
	method = strings.ToUpper(method)
	tree, exists := r.trees[method]
	if !exists {
		return false
	}

	return tree.Delete(path)
}

func (r *Router) Match(method, path string) MatchResult {
	method = strings.ToUpper(method)
	tree, exists := r.trees[method]
	if !exists {
		return MatchResult{Found: false}
	}

	value, params, found := tree.Search(path)
	if !found {
		return MatchResult{Found: false}
	}

	route, ok := value.(HTTPRoute)
	if !ok {
		return MatchResult{Found: false}
	}

	return MatchResult{
		Route:  route,
		Params: params,
		Found:  true,
	}
}

func (r *Router) List() []HTTPRoute {
	var routes []HTTPRoute
	for method, tree := range r.trees {
		for _, path := range tree.List() {
			value, _, found := tree.Search(path)
			if found {
				if route, ok := value.(HTTPRoute); ok {
					routes = append(routes, HTTPRoute{
						Path:    route.Path,
						Handler: route.Handler,
						Method:  method,
					})
				}
			}
		}
	}
	return routes
}

func (r *Router) Stats() map[string]radix.Stats {
	stats := make(map[string]radix.Stats)
	for method, tree := range r.trees {
		stats[method] = tree.Stats()
	}
	return stats
}
