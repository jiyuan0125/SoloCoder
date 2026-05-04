#ifndef MATCHER_H
#define MATCHER_H

#include "types.h"
#include "router.h"

typedef struct match_candidate {
    route_t *route;
    params_t *params;
    int specificity;
} match_candidate_t;

typedef struct {
    match_candidate_t *items;
    size_t count;
    size_t capacity;
} match_candidates_t;

char *normalize_path(const char *path);

match_candidates_t *match_candidates_create(size_t initial_capacity);
void match_candidates_destroy(match_candidates_t *candidates);
int match_candidates_add(match_candidates_t *candidates, route_t *route, params_t *params);

int calculate_specificity(route_t *route, size_t path_segments);

match_result_t router_match(router_t *router, http_method_t method, const char *path);

void match_result_cleanup(match_result_t *result);

#endif
