#ifndef ENV_EXPAND_H
#define ENV_EXPAND_H

#include "env_file.h"

typedef enum {
    ENV_EXPAND_OK = 0,
    ENV_EXPAND_ERROR_CYCLE,
    ENV_EXPAND_ERROR_MEMORY,
    ENV_EXPAND_ERROR_UNDEFINED_VAR
} env_expand_error_t;

typedef struct {
    int allow_system_env;
    int system_override;
    int error_on_undefined;
} env_expand_config_t;

env_expand_config_t env_expand_default_config(void);

typedef struct {
    env_file_t *file_vars;
    env_expand_config_t config;
} env_expand_ctx_t;

env_expand_ctx_t *env_expand_ctx_create(env_file_t *file_vars, env_expand_config_t config);
void env_expand_ctx_destroy(env_expand_ctx_t *ctx);

env_expand_error_t env_expand_value(env_expand_ctx_t *ctx, const char *value, char **result);

const char *env_expand_get_var(env_expand_ctx_t *ctx, const char *name);

#endif
