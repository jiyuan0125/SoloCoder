#ifndef ROUTER_H
#define ROUTER_H

#include "types.h"

typedef struct route_group {
    char *prefix;
    middleware_chain_t *middlewares;
    router_t *router;
} route_group_t;

router_t *router_create(void);
void router_destroy(router_t *router);

int router_register(router_t *router, http_method_t method, const char *pattern,
                    handler_func_t handler, void *user_data, middleware_chain_t *middlewares);

int router_register_with_prefix(router_t *router, const char *prefix, http_method_t method,
                                const char *pattern, handler_func_t handler, void *user_data,
                                middleware_chain_t *middlewares);

route_group_t *route_group_create(router_t *router, const char *prefix, middleware_chain_t *middlewares);
void route_group_destroy(route_group_t *group);

int route_group_register(route_group_t *group, http_method_t method, const char *pattern,
                         handler_func_t handler, void *user_data, middleware_chain_t *middlewares);

#endif
