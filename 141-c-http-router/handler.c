#include "handler.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int execute_middlewares(middleware_chain_t *chain, const params_t *params) {
    middleware_chain_t *current = chain;
    while (current) {
        if (current->func) {
            int result = current->func(params, current->user_data);
            if (result != 0) {
                return result;
            }
        }
        current = current->next;
    }
    return 0;
}

int dispatch_request(router_t *router, const char *method_str, const char *path) {
    if (!router || !method_str || !path) {
        printf("  [ERROR] Invalid parameters\n");
        return -1;
    }
    
    http_method_t method = method_from_string(method_str);
    if (method == HTTP_METHOD_COUNT) {
        printf("  [ERROR] Unknown HTTP method: %s\n", method_str);
        return -1;
    }
    
    printf("  Matching: %s %s\n", method_str, path);
    
    match_result_t result = router_match(router, method, path);
    
    if (!result.matched) {
        printf("  [404] No route matched\n\n");
        return -1;
    }
    
    printf("  [MATCHED] Pattern: %s\n", result.route->pattern);
    
    if (result.params.count > 0) {
        printf("  Parameters:\n");
        print_params(&result.params);
    }
    
    if (result.route->middlewares) {
        printf("  Executing middlewares...\n");
        int mw_result = execute_middlewares(result.route->middlewares, &result.params);
        if (mw_result != 0) {
            printf("  [FORBIDDEN] Middleware rejected (code: %d)\n\n", mw_result);
            match_result_cleanup(&result);
            return mw_result;
        }
        printf("  Middlewares passed\n");
    }
    
    if (result.route->handler) {
        printf("  Calling handler...\n");
        result.route->handler(&result.params, result.route->user_data);
    }
    
    printf("\n");
    
    match_result_cleanup(&result);
    return 0;
}

void print_params(const params_t *params) {
    if (!params || params->count == 0) {
        printf("    (none)\n");
        return;
    }
    for (size_t i = 0; i < params->count; i++) {
        printf("    %s = %s\n", params->items[i].key, params->items[i].value);
    }
}
