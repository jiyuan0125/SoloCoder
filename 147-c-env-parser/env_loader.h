#ifndef ENV_LOADER_H
#define ENV_LOADER_H

#include "env_file.h"
#include "env_expand.h"

#define ENV_LOADER_INITIAL_CACHE_SIZE 16

typedef struct {
    env_file_t *file_vars;
    env_expand_ctx_t *expand_ctx;
    env_expand_config_t config;
    char **result_cache;
    size_t cache_count;
    size_t cache_capacity;
} env_loader_t;

typedef enum {
    ENV_LOADER_OK = 0,
    ENV_LOADER_ERROR_FILE,
    ENV_LOADER_ERROR_MEMORY,
    ENV_LOADER_ERROR_SYNTAX,
    ENV_LOADER_ERROR_CYCLE,
    ENV_LOADER_ERROR_UNDEFINED_VAR
} env_loader_error_t;

env_loader_t *env_loader_create(env_expand_config_t config);
void env_loader_destroy(env_loader_t *loader);

env_loader_error_t env_loader_load(env_loader_t *loader, const char *path);

const char *env_loader_get_string(env_loader_t *loader, const char *key, const char *default_val);
int env_loader_get_int(env_loader_t *loader, const char *key, int default_val);
double env_loader_get_float(env_loader_t *loader, const char *key, double default_val);
int env_loader_get_bool(env_loader_t *loader, const char *key, int default_val);

#endif
