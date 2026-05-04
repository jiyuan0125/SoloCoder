#ifndef HANDLER_H
#define HANDLER_H

#include "types.h"
#include "matcher.h"

typedef struct {
    const char *method;
    const char *path;
    match_result_t result;
    int middleware_passed;
} request_context_t;

int execute_middlewares(middleware_chain_t *chain, const params_t *params);

int dispatch_request(router_t *router, const char *method_str, const char *path);

void print_params(const params_t *params);

#endif
