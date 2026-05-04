#include "env_expand.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <ctype.h>

env_expand_config_t env_expand_default_config(void) {
    env_expand_config_t config;
    config.allow_system_env = 1;
    config.system_override = 1;
    config.error_on_undefined = 0;
    return config;
}

env_expand_ctx_t *env_expand_ctx_create(env_file_t *file_vars, env_expand_config_t config) {
    env_expand_ctx_t *ctx = (env_expand_ctx_t *)malloc(sizeof(env_expand_ctx_t));
    if (!ctx) return NULL;
    
    ctx->file_vars = file_vars;
    ctx->config = config;
    return ctx;
}

void env_expand_ctx_destroy(env_expand_ctx_t *ctx) {
    if (ctx) {
        free(ctx);
    }
}

const char *env_expand_get_var(env_expand_ctx_t *ctx, const char *name) {
    if (!ctx || !name) return NULL;
    
    if (ctx->config.allow_system_env) {
        if (ctx->config.system_override) {
            const char *system_val = getenv(name);
            if (system_val) return system_val;
            
            if (ctx->file_vars) {
                return env_file_get(ctx->file_vars, name);
            }
        } else {
            if (ctx->file_vars) {
                const char *file_val = env_file_get(ctx->file_vars, name);
                if (file_val) return file_val;
            }
            
            return getenv(name);
        }
    } else {
        if (ctx->file_vars) {
            return env_file_get(ctx->file_vars, name);
        }
    }
    
    return NULL;
}

static int is_valid_var_char(char c) {
    return isalnum((unsigned char)c) || c == '_';
}

static char *extract_var_name(const char **ptr, int *is_braced) {
    const char *start = *ptr;
    *is_braced = 0;
    
    if (*start == '{') {
        *is_braced = 1;
        start++;
        *ptr = start;
    }
    
    const char *var_start = start;
    while (*start && is_valid_var_char(*start)) {
        start++;
    }
    
    if (var_start == start) {
        return NULL;
    }
    
    size_t var_len = start - var_start;
    char *var_name = (char *)malloc(var_len + 1);
    if (!var_name) return NULL;
    
    strncpy(var_name, var_start, var_len);
    var_name[var_len] = '\0';
    
    if (*is_braced) {
        if (*start == '}') {
            start++;
        }
    }
    
    *ptr = start;
    return var_name;
}

typedef struct cycle_node {
    char *name;
    struct cycle_node *next;
} cycle_node_t;

static cycle_node_t *cycle_stack_push(cycle_node_t *stack, const char *name) {
    cycle_node_t *node = (cycle_node_t *)malloc(sizeof(cycle_node_t));
    if (!node) return NULL;
    
    node->name = strdup(name);
    if (!node->name) {
        free(node);
        return NULL;
    }
    
    node->next = stack;
    return node;
}

static cycle_node_t *cycle_stack_pop(cycle_node_t *stack) {
    if (!stack) return NULL;
    
    cycle_node_t *next = stack->next;
    free(stack->name);
    free(stack);
    return next;
}

static int cycle_stack_contains(cycle_node_t *stack, const char *name) {
    cycle_node_t *current = stack;
    while (current) {
        if (strcmp(current->name, name) == 0) {
            return 1;
        }
        current = current->next;
    }
    return 0;
}

static env_expand_error_t expand_value_recursive(
    env_expand_ctx_t *ctx, 
    const char *value, 
    char **result,
    cycle_node_t *cycle_stack
);

static env_expand_error_t expand_single_var(
    env_expand_ctx_t *ctx,
    const char *var_name,
    char **expanded,
    cycle_node_t *cycle_stack
) {
    if (cycle_stack_contains(cycle_stack, var_name)) {
        return ENV_EXPAND_ERROR_CYCLE;
    }
    
    const char *var_value = env_expand_get_var(ctx, var_name);
    
    if (!var_value) {
        if (ctx->config.error_on_undefined) {
            return ENV_EXPAND_ERROR_UNDEFINED_VAR;
        }
        *expanded = strdup("");
        return *expanded ? ENV_EXPAND_OK : ENV_EXPAND_ERROR_MEMORY;
    }
    
    cycle_node_t *new_stack = cycle_stack_push(cycle_stack, var_name);
    if (!new_stack) {
        return ENV_EXPAND_ERROR_MEMORY;
    }
    
    env_expand_error_t err = expand_value_recursive(ctx, var_value, expanded, new_stack);
    
    while (new_stack != cycle_stack) {
        new_stack = cycle_stack_pop(new_stack);
    }
    
    return err;
}

static env_expand_error_t expand_value_recursive(
    env_expand_ctx_t *ctx, 
    const char *value, 
    char **result,
    cycle_node_t *cycle_stack
) {
    if (!value || !result) {
        return ENV_EXPAND_OK;
    }
    
    size_t result_cap = strlen(value) + 1;
    char *buf = (char *)malloc(result_cap);
    if (!buf) return ENV_EXPAND_ERROR_MEMORY;
    
    size_t buf_len = 0;
    const char *ptr = value;
    
    while (*ptr) {
        if (*ptr == '$') {
            ptr++;
            if (*ptr == '\0') {
                if (buf_len + 1 >= result_cap) {
                    result_cap *= 2;
                    char *new_buf = (char *)realloc(buf, result_cap);
                    if (!new_buf) {
                        free(buf);
                        return ENV_EXPAND_ERROR_MEMORY;
                    }
                    buf = new_buf;
                }
                buf[buf_len++] = '$';
                break;
            }
            
            int is_braced = 0;
            char *var_name = extract_var_name(&ptr, &is_braced);
            
            if (var_name) {
                char *expanded = NULL;
                env_expand_error_t err = expand_single_var(ctx, var_name, &expanded, cycle_stack);
                free(var_name);
                
                if (err != ENV_EXPAND_OK) {
                    free(buf);
                    return err;
                }
                
                if (expanded) {
                    size_t expanded_len = strlen(expanded);
                    size_t needed = buf_len + expanded_len + 1;
                    if (needed > result_cap) {
                        while (result_cap < needed) result_cap *= 2;
                        char *new_buf = (char *)realloc(buf, result_cap);
                        if (!new_buf) {
                            free(buf);
                            free(expanded);
                            return ENV_EXPAND_ERROR_MEMORY;
                        }
                        buf = new_buf;
                    }
                    memcpy(buf + buf_len, expanded, expanded_len);
                    buf_len += expanded_len;
                    free(expanded);
                }
            } else {
                if (buf_len + 2 >= result_cap) {
                    result_cap *= 2;
                    char *new_buf = (char *)realloc(buf, result_cap);
                    if (!new_buf) {
                        free(buf);
                        return ENV_EXPAND_ERROR_MEMORY;
                    }
                    buf = new_buf;
                }
                buf[buf_len++] = '$';
                if (*ptr) {
                    buf[buf_len++] = *ptr++;
                }
            }
        } else {
            if (buf_len + 1 >= result_cap) {
                result_cap *= 2;
                char *new_buf = (char *)realloc(buf, result_cap);
                if (!new_buf) {
                    free(buf);
                    return ENV_EXPAND_ERROR_MEMORY;
                }
                buf = new_buf;
            }
            buf[buf_len++] = *ptr++;
        }
    }
    
    buf[buf_len] = '\0';
    *result = buf;
    return ENV_EXPAND_OK;
}

env_expand_error_t env_expand_value(env_expand_ctx_t *ctx, const char *value, char **result) {
    if (!ctx || !result) {
        return ENV_EXPAND_ERROR_MEMORY;
    }
    
    if (!value) {
        *result = strdup("");
        return *result ? ENV_EXPAND_OK : ENV_EXPAND_ERROR_MEMORY;
    }
    
    return expand_value_recursive(ctx, value, result, NULL);
}
