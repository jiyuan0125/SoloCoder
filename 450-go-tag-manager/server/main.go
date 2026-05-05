package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
)

type Mux struct {
	routes []route
}

type route struct {
	pattern *regexp.Regexp
	handler http.HandlerFunc
}

func NewMux() *Mux {
	return &Mux{}
}

func (m *Mux) Handle(pattern string, handler http.HandlerFunc) {
	m.routes = append(m.routes, route{
		pattern: regexp.MustCompile(pattern),
		handler: handler,
	})
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range m.routes {
		if route.pattern.MatchString(r.URL.Path) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

func main() {
	store := NewStore()
	handler := NewHandler(store)

	mux := NewMux()

	mux.Handle(`^/tags$`, handler.ListTags)
	mux.Handle(`^/tags/create$`, handler.CreateTag)
	mux.Handle(`^/tags/[^/]+/update$`, handler.UpdateTag)
	mux.Handle(`^/tags/[^/]+/delete$`, handler.DeleteTag)
	mux.Handle(`^/tags/[^/]+/impact$`, handler.GetTagImpact)
	mux.Handle(`^/tags/[^/]+/apply$`, handler.ApplyTagToUser)
	mux.Handle(`^/tags/[^/]+/batch-apply$`, handler.BatchApplyTag)
	mux.Handle(`^/tags/[^/]+/remove$`, handler.RemoveUserTag)
	mux.Handle(`^/tags/[^/]+$`, handler.GetTag)

	mux.Handle(`^/groups$`, handler.ListTagGroups)
	mux.Handle(`^/groups/create$`, handler.CreateTagGroup)
	mux.Handle(`^/groups/[^/]+$`, handler.GetTagGroup)

	mux.Handle(`^/users/[^/]+/tags$`, handler.GetUserTags)
	mux.Handle(`^/users/filter$`, handler.FilterUsers)

	mux.Handle(`^/stats$`, handler.GetTagStats)
	mux.Handle(`^/trend$`, handler.GetTagTrend)
	mux.Handle(`^/audit$`, handler.ListAuditLogs)

	port := ":8080"
	fmt.Printf("Tag Manager Server starting on %s...\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("  GET  /tags                    - List all tags")
	fmt.Println("  POST /tags/create             - Create a tag")
	fmt.Println("  GET  /tags/:id                - Get a tag")
	fmt.Println("  PUT  /tags/:id/update         - Update a tag")
	fmt.Println("  DELETE /tags/:id/delete       - Delete a tag (with cascade)")
	fmt.Println("  GET  /tags/:id/impact         - Get tag deletion impact")
	fmt.Println("  POST /tags/:id/apply          - Apply tag to a user")
	fmt.Println("  POST /tags/:id/batch-apply    - Batch apply tag to users")
	fmt.Println("  DELETE /tags/:id/remove       - Remove tag from a user")
	fmt.Println("  GET  /groups                  - List all tag groups")
	fmt.Println("  POST /groups/create           - Create a tag group")
	fmt.Println("  GET  /groups/:id              - Get a tag group")
	fmt.Println("  GET  /users/:id/tags          - Get user's tags")
	fmt.Println("  POST /users/filter            - Filter users by tags")
	fmt.Println("  GET  /stats                   - Get tag usage stats")
	fmt.Println("  GET  /trend                   - Get tag trend (last 30 days)")
	fmt.Println("  GET  /audit                   - List audit logs")

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
