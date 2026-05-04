#include "matcher.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

char *normalize_path(const char *path) {
    if (!path) return NULL;
    size_t len = strlen(path);
    if (len == 0) return strdup("/");
    
    char *normalized = malloc(len + 2);
    if (!normalized) return NULL;
    
    const char *src = path;
    char *dst = normalized;
    
    if (*src != '/') {
        *dst++ = '/';
    }
    
    while (*src) {
        if (*src == '/') {
            while (*src == '/') src++;
            *dst++ = '/';
        } else {
            *dst++ = *src++;
        }
    }
    
    *dst = '\0';
    
    size_t new_len = strlen(normalized);
    if (new_len > 1 && normalized[new_len - 1] == '/') {
        normalized[new_len - 1] = '\0';
    }
    
    return normalized;
}

match_candidates_t *match_candidates_create(size_t initial_capacity) {
    if (initial_capacity == 0) initial_capacity = 4;
    match_candidates_t *candidates = malloc(sizeof(match_candidates_t));
    if (!candidates) return NULL;
    candidates->items = malloc(initial_capacity * sizeof(match_candidate_t));
    if (!candidates->items) {
        free(candidates);
        return NULL;
    }
    candidates->count = 0;
    candidates->capacity = initial_capacity;
    return candidates;
}

void match_candidates_destroy(match_candidates_t *candidates) {
    if (!candidates) return;
    for (size_t i = 0; i < candidates->count; i++) {
        params_destroy(candidates->items[i].params);
    }
    free(candidates->items);
    free(candidates);
}

static int match_candidates_ensure_capacity(match_candidates_t *candidates, size_t needed) {
    if (candidates->count + needed <= candidates->capacity) return 0;
    size_t new_capacity = candidates->capacity * 2;
    while (new_capacity < candidates->count + needed) new_capacity *= 2;
    match_candidate_t *new_items = realloc(candidates->items, new_capacity * sizeof(match_candidate_t));
    if (!new_items) return -1;
    candidates->items = new_items;
    candidates->capacity = new_capacity;
    return 0;
}

int match_candidates_add(match_candidates_t *candidates, route_t *route, params_t *params) {
    if (!candidates || !route) return -1;
    if (match_candidates_ensure_capacity(candidates, 1) != 0) return -1;
    candidates->items[candidates->count].route = route;
    candidates->items[candidates->count].params = params;
    candidates->count++;
    return 0;
}

int calculate_specificity(route_t *route, size_t path_segments) {
    (void)path_segments;
    if (!route) return -1;
    
    int static_score = 1000;
    int param_score = 100;
    int wildcard_score = 1;
    
    int specificity = 0;
    
    if (route->has_wildcard) {
        specificity += wildcard_score;
    } else {
        specificity += static_score - (route->param_count * param_score);
    }
    
    return specificity;
}

static void find_matches(trie_node_t *node, char **path_segs, size_t path_count, size_t seg_idx,
                         http_method_t method, params_t *current_params,
                         match_candidates_t *candidates, const char *pattern_segments[],
                         size_t pattern_idx) {
    (void)pattern_segments;
    (void)pattern_idx;
    if (!node) return;
    
    if (seg_idx == path_count) {
        route_t *route = node->routes[method];
        if (route) {
            params_t *params_copy = params_create(current_params->count);
            for (size_t i = 0; i < current_params->count; i++) {
                params_add(params_copy, current_params->items[i].key, current_params->items[i].value);
            }
            match_candidates_add(candidates, route, params_copy);
        }
        return;
    }
    
    const char *current_seg = path_segs[seg_idx];
    
    for (size_t i = 0; i < node->child_count; i++) {
        trie_node_t *child = &node->children[i];
        
        if (!child->is_param && !child->is_wildcard) {
            if (child->segment && strcmp(child->segment, current_seg) == 0) {
                find_matches(child, path_segs, path_count, seg_idx + 1, method,
                             current_params, candidates, NULL, 0);
            }
        }
    }
    
    for (size_t i = 0; i < node->child_count; i++) {
        trie_node_t *child = &node->children[i];
        
        if (child->is_param) {
            char *decoded_value = url_decode(current_seg);
            if (decoded_value) {
                params_add(current_params, child->segment, decoded_value);
                find_matches(child, path_segs, path_count, seg_idx + 1, method,
                             current_params, candidates, NULL, 0);
                
                if (current_params->count > 0) {
                    free(current_params->items[current_params->count - 1].key);
                    free(current_params->items[current_params->count - 1].value);
                    current_params->count--;
                }
                free(decoded_value);
            }
        }
    }
    
    for (size_t i = 0; i < node->child_count; i++) {
        trie_node_t *child = &node->children[i];
        
        if (child->is_wildcard) {
            route_t *route = child->routes[method];
            if (route) {
                char wildcard_value[4096] = "";
                size_t offset = 0;
                for (size_t j = seg_idx; j < path_count; j++) {
                    if (offset > 0) {
                        wildcard_value[offset++] = '/';
                    }
                    char *decoded = url_decode(path_segs[j]);
                    if (decoded) {
                        size_t seg_len = strlen(decoded);
                        if (offset + seg_len < sizeof(wildcard_value) - 1) {
                            memcpy(wildcard_value + offset, decoded, seg_len);
                            offset += seg_len;
                        }
                        free(decoded);
                    }
                }
                wildcard_value[offset] = '\0';
                
                params_t *params_copy = params_create(current_params->count + 1);
                for (size_t k = 0; k < current_params->count; k++) {
                    params_add(params_copy, current_params->items[k].key, current_params->items[k].value);
                }
                params_add(params_copy, "*", wildcard_value);
                match_candidates_add(candidates, route, params_copy);
            }
        }
    }
}

static match_candidate_t *select_best_match(match_candidates_t *candidates, size_t path_segments) {
    if (!candidates || candidates->count == 0) return NULL;
    
    match_candidate_t *best = &candidates->items[0];
    int best_score = calculate_specificity(best->route, path_segments);
    
    for (size_t i = 1; i < candidates->count; i++) {
        match_candidate_t *current = &candidates->items[i];
        int current_score = calculate_specificity(current->route, path_segments);
        
        if (current_score > best_score) {
            best = current;
            best_score = current_score;
        }
    }
    
    return best;
}

match_result_t router_match(router_t *router, http_method_t method, const char *path) {
    match_result_t result = {0};
    
    if (!router || !path || method < 0 || method >= HTTP_METHOD_COUNT) {
        return result;
    }
    
    char *normalized = normalize_path(path);
    if (!normalized) {
        return result;
    }
    
    size_t path_count;
    char **path_segs = split_path(normalized, &path_count);
    free(normalized);
    
    if (!path_segs && path_count > 0) {
        return result;
    }
    
    match_candidates_t *candidates = match_candidates_create(8);
    if (!candidates) {
        free_segments(path_segs, path_count);
        return result;
    }
    
    params_t *temp_params = params_create(4);
    if (!temp_params) {
        match_candidates_destroy(candidates);
        free_segments(path_segs, path_count);
        return result;
    }
    
    if (path_count == 0) {
        route_t *root_route = router->root->routes[method];
        if (root_route) {
            params_t *empty_params = params_create(0);
            match_candidates_add(candidates, root_route, empty_params);
        }
    } else {
        find_matches(router->root, path_segs, path_count, 0, method,
                     temp_params, candidates, NULL, 0);
    }
    
    params_destroy(temp_params);
    free_segments(path_segs, path_count);
    
    if (candidates->count > 0) {
        match_candidate_t *best = select_best_match(candidates, path_count);
        if (best) {
            result.route = best->route;
            result.params = *best->params;
            result.matched = 1;
            
            best->params = NULL;
        }
    }
    
    match_candidates_destroy(candidates);
    return result;
}

void match_result_cleanup(match_result_t *result) {
    if (!result) return;
    if (result->params.items) {
        for (size_t i = 0; i < result->params.count; i++) {
            free(result->params.items[i].key);
            free(result->params.items[i].value);
        }
        free(result->params.items);
        result->params.items = NULL;
        result->params.count = 0;
        result->params.capacity = 0;
    }
}
