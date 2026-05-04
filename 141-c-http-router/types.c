#include "types.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>

http_method_t method_from_string(const char *method) {
    if (strcmp(method, "GET") == 0) return HTTP_GET;
    if (strcmp(method, "POST") == 0) return HTTP_POST;
    if (strcmp(method, "PUT") == 0) return HTTP_PUT;
    if (strcmp(method, "DELETE") == 0) return HTTP_DELETE;
    return HTTP_METHOD_COUNT;
}

const char *method_to_string(http_method_t method) {
    switch (method) {
        case HTTP_GET: return "GET";
        case HTTP_POST: return "POST";
        case HTTP_PUT: return "PUT";
        case HTTP_DELETE: return "DELETE";
        default: return "UNKNOWN";
    }
}

params_t *params_create(size_t initial_capacity) {
    if (initial_capacity == 0) initial_capacity = 4;
    params_t *params = malloc(sizeof(params_t));
    if (!params) return NULL;
    params->items = malloc(initial_capacity * sizeof(param_t));
    if (!params->items) {
        free(params);
        return NULL;
    }
    params->count = 0;
    params->capacity = initial_capacity;
    return params;
}

void params_destroy(params_t *params) {
    if (!params) return;
    for (size_t i = 0; i < params->count; i++) {
        free(params->items[i].key);
        free(params->items[i].value);
    }
    free(params->items);
    free(params);
}

static int params_ensure_capacity(params_t *params, size_t needed) {
    if (params->count + needed <= params->capacity) return 0;
    size_t new_capacity = params->capacity * 2;
    while (new_capacity < params->count + needed) new_capacity *= 2;
    param_t *new_items = realloc(params->items, new_capacity * sizeof(param_t));
    if (!new_items) return -1;
    params->items = new_items;
    params->capacity = new_capacity;
    return 0;
}

int params_add(params_t *params, const char *key, const char *value) {
    if (!params || !key || !value) return -1;
    if (params_ensure_capacity(params, 1) != 0) return -1;
    params->items[params->count].key = strdup(key);
    params->items[params->count].value = strdup(value);
    if (!params->items[params->count].key || !params->items[params->count].value) {
        free(params->items[params->count].key);
        free(params->items[params->count].value);
        return -1;
    }
    params->count++;
    return 0;
}

const char *params_get(const params_t *params, const char *key) {
    if (!params || !key) return NULL;
    for (size_t i = 0; i < params->count; i++) {
        if (strcmp(params->items[i].key, key) == 0) {
            return params->items[i].value;
        }
    }
    return NULL;
}

middleware_chain_t *middleware_chain_create(middleware_func_t func, void *user_data) {
    middleware_chain_t *chain = malloc(sizeof(middleware_chain_t));
    if (!chain) return NULL;
    chain->func = func;
    chain->user_data = user_data;
    chain->next = NULL;
    return chain;
}

void middleware_chain_destroy(middleware_chain_t *chain) {
    while (chain) {
        middleware_chain_t *next = chain->next;
        free(chain);
        chain = next;
    }
}

middleware_chain_t *middleware_chain_append(middleware_chain_t *chain, middleware_func_t func, void *user_data) {
    middleware_chain_t *new_node = middleware_chain_create(func, user_data);
    if (!new_node) return chain;
    if (!chain) return new_node;
    middleware_chain_t *current = chain;
    while (current->next) current = current->next;
    current->next = new_node;
    return chain;
}

static unsigned char hex_to_byte(char c) {
    if (c >= '0' && c <= '9') return c - '0';
    if (c >= 'a' && c <= 'f') return c - 'a' + 10;
    if (c >= 'A' && c <= 'F') return c - 'A' + 10;
    return 0;
}

char *url_decode(const char *str) {
    if (!str) return NULL;
    size_t len = strlen(str);
    char *decoded = malloc(len + 1);
    if (!decoded) return NULL;
    char *dst = decoded;
    const char *src = str;
    
    while (*src) {
        if (*src == '%' && src[1] && src[2] && isxdigit(src[1]) && isxdigit(src[2])) {
            *dst = (hex_to_byte(src[1]) << 4) | hex_to_byte(src[2]);
            src += 3;
        } else if (*src == '+') {
            *dst = ' ';
            src++;
        } else {
            *dst = *src;
            src++;
        }
        dst++;
    }
    *dst = '\0';
    return decoded;
}

char **split_path(const char *path, size_t *count) {
    if (!path || !count) return NULL;
    *count = 0;
    
    size_t len = strlen(path);
    if (len == 0 || path[0] != '/') return NULL;
    
    char *copy = strdup(path);
    if (!copy) return NULL;
    
    char **segments = NULL;
    size_t seg_count = 0;
    size_t seg_cap = 0;
    
    char *start = copy + 1;
    char *end = start;
    
    while (*end) {
        if (*end == '/') {
            if (end > start) {
                if (seg_count >= seg_cap) {
                    size_t new_cap = seg_cap == 0 ? 4 : seg_cap * 2;
                    char **new_seg = realloc(segments, new_cap * sizeof(char *));
                    if (!new_seg) {
                        free(copy);
                        for (size_t i = 0; i < seg_count; i++) free(segments[i]);
                        free(segments);
                        return NULL;
                    }
                    segments = new_seg;
                    seg_cap = new_cap;
                }
                *end = '\0';
                segments[seg_count++] = strdup(start);
            }
            start = end + 1;
            end = start;
        } else {
            end++;
        }
    }
    
    if (end > start) {
        if (seg_count >= seg_cap) {
            size_t new_cap = seg_cap == 0 ? 4 : seg_cap * 2;
            char **new_seg = realloc(segments, new_cap * sizeof(char *));
            if (!new_seg) {
                free(copy);
                for (size_t i = 0; i < seg_count; i++) free(segments[i]);
                free(segments);
                return NULL;
            }
            segments = new_seg;
            seg_cap = new_cap;
        }
        segments[seg_count++] = strdup(start);
    }
    
    free(copy);
    *count = seg_count;
    return segments;
}

void free_segments(char **segments, size_t count) {
    if (!segments) return;
    for (size_t i = 0; i < count; i++) {
        free(segments[i]);
    }
    free(segments);
}
