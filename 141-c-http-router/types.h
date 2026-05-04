#ifndef TYPES_H
#define TYPES_H

#include <stddef.h>

typedef enum {
    HTTP_GET,
    HTTP_POST,
    HTTP_PUT,
    HTTP_DELETE,
    HTTP_METHOD_COUNT
} http_method_t;

typedef struct {
    char *key;
    char *value;
} param_t;

typedef struct {
    param_t *items;
    size_t count;
    size_t capacity;
} params_t;

typedef void (*handler_func_t)(const params_t *params, void *user_data);
typedef int (*middleware_func_t)(const params_t *params, void *user_data);

typedef struct middleware_chain {
    middleware_func_t func;
    void *user_data;
    struct middleware_chain *next;
} middleware_chain_t;

typedef struct {
    http_method_t method;
    char *pattern;
    handler_func_t handler;
    void *user_data;
    middleware_chain_t *middlewares;
    int param_count;
    int has_wildcard;
} route_t;

typedef struct trie_node {
    char *segment;
    int is_param;
    int is_wildcard;
    route_t *routes[HTTP_METHOD_COUNT];
    struct trie_node *children;
    size_t child_count;
    size_t child_capacity;
} trie_node_t;

typedef struct {
    trie_node_t *root;
} router_t;

typedef struct {
    route_t *route;
    params_t params;
    int matched;
} match_result_t;

http_method_t method_from_string(const char *method);
const char *method_to_string(http_method_t method);

params_t *params_create(size_t initial_capacity);
void params_destroy(params_t *params);
int params_add(params_t *params, const char *key, const char *value);
const char *params_get(const params_t *params, const char *key);

middleware_chain_t *middleware_chain_create(middleware_func_t func, void *user_data);
void middleware_chain_destroy(middleware_chain_t *chain);
middleware_chain_t *middleware_chain_append(middleware_chain_t *chain, middleware_func_t func, void *user_data);

char *url_decode(const char *str);

char **split_path(const char *path, size_t *count);
void free_segments(char **segments, size_t count);

#endif
