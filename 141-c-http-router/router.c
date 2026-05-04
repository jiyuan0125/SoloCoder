#include "router.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static trie_node_t *trie_node_create(const char *segment, int is_param, int is_wildcard) {
    trie_node_t *node = malloc(sizeof(trie_node_t));
    if (!node) return NULL;
    node->segment = segment ? strdup(segment) : NULL;
    node->is_param = is_param;
    node->is_wildcard = is_wildcard;
    memset(node->routes, 0, sizeof(node->routes));
    node->children = NULL;
    node->child_count = 0;
    node->child_capacity = 0;
    return node;
}

static void trie_node_destroy(trie_node_t *node) {
    if (!node) return;
    for (int i = 0; i < HTTP_METHOD_COUNT; i++) {
        if (node->routes[i]) {
            free(node->routes[i]->pattern);
            middleware_chain_destroy(node->routes[i]->middlewares);
            free(node->routes[i]);
        }
    }
    for (size_t i = 0; i < node->child_count; i++) {
        trie_node_destroy(&node->children[i]);
    }
    free(node->children);
    free(node->segment);
}

static int trie_node_ensure_children(trie_node_t *node, size_t needed) {
    if (node->child_count + needed <= node->child_capacity) return 0;
    size_t new_capacity = node->child_capacity == 0 ? 4 : node->child_capacity * 2;
    while (new_capacity < node->child_count + needed) new_capacity *= 2;
    trie_node_t *new_children = realloc(node->children, new_capacity * sizeof(trie_node_t));
    if (!new_children) return -1;
    for (size_t i = node->child_count; i < new_capacity; i++) {
        memset(&new_children[i], 0, sizeof(trie_node_t));
    }
    node->children = new_children;
    node->child_capacity = new_capacity;
    return 0;
}

static trie_node_t *trie_node_find_child(trie_node_t *parent, const char *segment, int is_param, int is_wildcard) {
    for (size_t i = 0; i < parent->child_count; i++) {
        trie_node_t *child = &parent->children[i];
        if (child->is_param == is_param && child->is_wildcard == is_wildcard) {
            if (is_param || is_wildcard) {
                return child;
            } else if (child->segment && strcmp(child->segment, segment) == 0) {
                return child;
            }
        }
    }
    return NULL;
}

static trie_node_t *trie_node_add_child(trie_node_t *parent, const char *segment, int is_param, int is_wildcard) {
    trie_node_t *existing = trie_node_find_child(parent, segment, is_param, is_wildcard);
    if (existing) return existing;
    
    if (trie_node_ensure_children(parent, 1) != 0) return NULL;
    
    trie_node_t *new_child = &parent->children[parent->child_count];
    new_child->segment = segment ? strdup(segment) : NULL;
    new_child->is_param = is_param;
    new_child->is_wildcard = is_wildcard;
    memset(new_child->routes, 0, sizeof(new_child->routes));
    new_child->children = NULL;
    new_child->child_count = 0;
    new_child->child_capacity = 0;
    
    parent->child_count++;
    return new_child;
}

static int count_params_in_pattern(const char *pattern) {
    if (!pattern) return 0;
    int count = 0;
    const char *p = pattern;
    while (*p) {
        if (*p == ':') count++;
        p++;
    }
    return count;
}

router_t *router_create(void) {
    router_t *router = malloc(sizeof(router_t));
    if (!router) return NULL;
    router->root = trie_node_create(NULL, 0, 0);
    if (!router->root) {
        free(router);
        return NULL;
    }
    return router;
}

void router_destroy(router_t *router) {
    if (!router) return;
    trie_node_destroy(router->root);
    free(router->root);
    free(router);
}

int router_register(router_t *router, http_method_t method, const char *pattern,
                    handler_func_t handler, void *user_data, middleware_chain_t *middlewares) {
    if (!router || !pattern || !handler) return -1;
    if (method < 0 || method >= HTTP_METHOD_COUNT) return -1;
    
    size_t seg_count;
    char **segments = split_path(pattern, &seg_count);
    if (!segments && seg_count > 0) return -1;
    
    trie_node_t *current = router->root;
    
    for (size_t i = 0; i < seg_count; i++) {
        const char *seg = segments[i];
        int is_param = (seg[0] == ':');
        int is_wildcard = (strcmp(seg, "*") == 0);
        
        if (is_wildcard && i != seg_count - 1) {
            free_segments(segments, seg_count);
            return -1;
        }
        
        const char *node_seg = seg;
        if (is_param) node_seg = seg + 1;
        
        current = trie_node_add_child(current, node_seg, is_param, is_wildcard);
        if (!current) {
            free_segments(segments, seg_count);
            return -1;
        }
        
        if (is_wildcard) break;
    }
    
    if (current->routes[method]) {
        free(current->routes[method]->pattern);
        middleware_chain_destroy(current->routes[method]->middlewares);
        free(current->routes[method]);
    }
    
    route_t *route = malloc(sizeof(route_t));
    if (!route) {
        free_segments(segments, seg_count);
        return -1;
    }
    route->method = method;
    route->pattern = strdup(pattern);
    route->handler = handler;
    route->user_data = user_data;
    route->middlewares = middlewares;
    route->param_count = count_params_in_pattern(pattern);
    route->has_wildcard = (seg_count > 0 && strcmp(segments[seg_count - 1], "*") == 0);
    
    current->routes[method] = route;
    
    free_segments(segments, seg_count);
    return 0;
}

int router_register_with_prefix(router_t *router, const char *prefix, http_method_t method,
                                const char *pattern, handler_func_t handler, void *user_data,
                                middleware_chain_t *middlewares) {
    if (!router || !pattern) return -1;
    
    char full_pattern[1024];
    if (prefix && prefix[0]) {
        size_t prefix_len = strlen(prefix);
        
        if (prefix[prefix_len - 1] == '/' && pattern[0] == '/') {
            snprintf(full_pattern, sizeof(full_pattern), "%.*s%s", 
                     (int)(prefix_len - 1), prefix, pattern);
        } else if (prefix[prefix_len - 1] != '/' && pattern[0] != '/') {
            snprintf(full_pattern, sizeof(full_pattern), "%s/%s", prefix, pattern);
        } else {
            snprintf(full_pattern, sizeof(full_pattern), "%s%s", prefix, pattern);
        }
    } else {
        snprintf(full_pattern, sizeof(full_pattern), "%s", pattern);
    }
    
    return router_register(router, method, full_pattern, handler, user_data, middlewares);
}

route_group_t *route_group_create(router_t *router, const char *prefix, middleware_chain_t *middlewares) {
    if (!router) return NULL;
    
    route_group_t *group = malloc(sizeof(route_group_t));
    if (!group) return NULL;
    
    group->router = router;
    group->prefix = prefix ? strdup(prefix) : NULL;
    
    group->middlewares = NULL;
    middleware_chain_t *curr = middlewares;
    while (curr) {
        group->middlewares = middleware_chain_append(group->middlewares, curr->func, curr->user_data);
        curr = curr->next;
    }
    
    return group;
}

void route_group_destroy(route_group_t *group) {
    if (!group) return;
    free(group->prefix);
    middleware_chain_destroy(group->middlewares);
    free(group);
}

int route_group_register(route_group_t *group, http_method_t method, const char *pattern,
                         handler_func_t handler, void *user_data, middleware_chain_t *middlewares) {
    if (!group || !pattern) return -1;
    
    middleware_chain_t *combined = NULL;
    middleware_chain_t *curr = group->middlewares;
    while (curr) {
        combined = middleware_chain_append(combined, curr->func, curr->user_data);
        curr = curr->next;
    }
    curr = middlewares;
    while (curr) {
        combined = middleware_chain_append(combined, curr->func, curr->user_data);
        curr = curr->next;
    }
    
    int result = router_register_with_prefix(group->router, group->prefix, method, pattern,
                                              handler, user_data, combined);
    
    if (result != 0) {
        middleware_chain_destroy(combined);
    }
    
    return result;
}
