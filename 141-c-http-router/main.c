#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "types.h"
#include "router.h"
#include "matcher.h"
#include "handler.h"

static void handle_home(const params_t *params, void *user_data) {
    (void)params;
    (void)user_data;
    printf("    -> Handler: Welcome to Home Page!\n");
}

static void handle_users_list(const params_t *params, void *user_data) {
    (void)params;
    (void)user_data;
    printf("    -> Handler: Listing all users\n");
}

static void handle_user_me(const params_t *params, void *user_data) {
    (void)params;
    (void)user_data;
    printf("    -> Handler: Current user profile (me)\n");
}

static void handle_user_by_id(const params_t *params, void *user_data) {
    (void)user_data;
    const char *id = params_get(params, "id");
    printf("    -> Handler: User profile for ID: %s\n", id ? id : "unknown");
}

static void handle_files(const params_t *params, void *user_data) {
    (void)user_data;
    const char *category = params_get(params, "category");
    const char *name = params_get(params, "name");
    printf("    -> Handler: File - Category: %s, Name: %s\n",
           category ? category : "unknown", name ? name : "unknown");
}

static void handle_static(const params_t *params, void *user_data) {
    (void)user_data;
    const char *path = params_get(params, "*");
    printf("    -> Handler: Static file: %s\n", path ? path : "unknown");
}

static void handle_api_v2_users(const params_t *params, void *user_data) {
    (void)params;
    (void)user_data;
    printf("    -> Handler: API v2 - List users\n");
}

static void handle_api_v2_user(const params_t *params, void *user_data) {
    (void)user_data;
    const char *id = params_get(params, "id");
    printf("    -> Handler: API v2 - User: %s\n", id ? id : "unknown");
}

static int auth_middleware(const params_t *params, void *user_data) {
    (void)params;
    (void)user_data;
    printf("      [Middleware] Checking authentication... PASSED\n");
    return 0;
}

static int admin_middleware(const params_t *params, void *user_data) {
    (void)params;
    (void)user_data;
    printf("      [Middleware] Checking admin privileges... PASSED\n");
    return 0;
}

static void demo_section(const char *title) {
    printf("========================================\n");
    printf("%s\n", title);
    printf("========================================\n\n");
}

int main(void) {
    printf("========================================\n");
    printf("   HTTP Router Module Demo\n");
    printf("========================================\n\n");

    router_t *router = router_create();
    if (!router) {
        fprintf(stderr, "Failed to create router\n");
        return 1;
    }

    demo_section("1. Registering Routes");

    printf("Registering static routes...\n");
    router_register(router, HTTP_GET, "/", handle_home, NULL, NULL);
    router_register(router, HTTP_GET, "/web/users", handle_users_list, NULL, NULL);
    printf("  GET /\n");
    printf("  GET /web/users\n\n");

    printf("Registering static vs parameter routes (to test priority)...\n");
    router_register(router, HTTP_GET, "/web/users/me", handle_user_me, NULL, NULL);
    router_register(router, HTTP_GET, "/web/users/:id", handle_user_by_id, NULL, NULL);
    printf("  GET /web/users/me  (static, higher priority)\n");
    printf("  GET /web/users/:id (parameter)\n\n");

    printf("Registering multi-parameter route...\n");
    router_register(router, HTTP_GET, "/files/:category/:name", handle_files, NULL, NULL);
    printf("  GET /files/:category/:name\n\n");

    printf("Registering wildcard route...\n");
    router_register(router, HTTP_GET, "/static/*", handle_static, NULL, NULL);
    printf("  GET /static/*\n\n");

    printf("Registering routes with middlewares...\n");
    middleware_chain_t *admin_chain = NULL;
    admin_chain = middleware_chain_append(admin_chain, auth_middleware, NULL);
    admin_chain = middleware_chain_append(admin_chain, admin_middleware, NULL);
    router_register(router, HTTP_DELETE, "/web/users/:id", handle_user_by_id, NULL, admin_chain);
    printf("  DELETE /web/users/:id (with auth + admin middlewares)\n\n");

    printf("Creating route group with prefix /web/v2 and auth middleware...\n");
    middleware_chain_t *api_chain = middleware_chain_create(auth_middleware, NULL);
    route_group_t *api_v2 = route_group_create(router, "/web/v2", api_chain);
    middleware_chain_destroy(api_chain);
    
    route_group_register(api_v2, HTTP_GET, "/users", handle_api_v2_users, NULL, NULL);
    route_group_register(api_v2, HTTP_GET, "/users/:id", handle_api_v2_user, NULL, NULL);
    printf("  GET /web/v2/users    (auto-prefixed, with auth middleware)\n");
    printf("  GET /web/v2/users/:id (auto-prefixed, with auth middleware)\n\n");

    demo_section("2. Testing Static Route Matching");
    dispatch_request(router, "GET", "/");
    dispatch_request(router, "GET", "/web/users");

    demo_section("3. Testing Static vs Parameter Priority");
    printf("Note: /web/users/me (static) should match before /web/users/:id\n\n");
    dispatch_request(router, "GET", "/web/users/me");
    dispatch_request(router, "GET", "/web/users/123");
    dispatch_request(router, "GET", "/web/users/abc%20xyz");

    demo_section("4. Testing Multi-Parameter Routes");
    dispatch_request(router, "GET", "/files/images/photo.jpg");
    dispatch_request(router, "GET", "/files/documents/report%2Epdf");

    demo_section("5. Testing Wildcard Routes");
    dispatch_request(router, "GET", "/static/css/style.css");
    dispatch_request(router, "GET", "/static/js/app/lib/utils.js");
    dispatch_request(router, "GET", "/static/index.html");

    demo_section("6. Testing Path Normalization (Trailing Slash)");
    printf("Note: /web/users and /web/users/ should match the same route\n\n");
    dispatch_request(router, "GET", "/web/users");
    dispatch_request(router, "GET", "/web/users/");
    dispatch_request(router, "GET", "//web//users//");

    demo_section("7. Testing Route Groups with Prefix");
    dispatch_request(router, "GET", "/web/v2/users");
    dispatch_request(router, "GET", "/web/v2/users/999");

    demo_section("8. Testing Middlewares");
    dispatch_request(router, "DELETE", "/web/users/5");

    demo_section("9. Testing 404 - No Match");
    dispatch_request(router, "GET", "/nonexistent");
    dispatch_request(router, "POST", "/web/users");
    dispatch_request(router, "GET", "/web/users/123/details");

    demo_section("10. Testing Different HTTP Methods");
    printf("Note: DELETE /web/users/:id has different handler than GET\n\n");
    dispatch_request(router, "GET", "/web/users/42");
    dispatch_request(router, "DELETE", "/web/users/42");

    demo_section("Cleanup");
    route_group_destroy(api_v2);
    router_destroy(router);
    printf("Router and resources cleaned up.\n\n");

    printf("========================================\n");
    printf("   Demo Complete!\n");
    printf("========================================\n");

    return 0;
}
