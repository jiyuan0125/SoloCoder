#include "env_loader.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <ctype.h>

env_loader_t *env_loader_create(env_expand_config_t config) {
    env_loader_t *loader = (env_loader_t *)malloc(sizeof(env_loader_t));
    if (!loader) return NULL;
    
    loader->file_vars = env_file_create();
    if (!loader->file_vars) {
        free(loader);
        return NULL;
    }
    
    loader->config = config;
    loader->expand_ctx = env_expand_ctx_create(loader->file_vars, config);
    if (!loader->expand_ctx) {
        env_file_destroy(loader->file_vars);
        free(loader);
        return NULL;
    }
    
    return loader;
}

void env_loader_destroy(env_loader_t *loader) {
    if (!loader) return;
    
    env_expand_ctx_destroy(loader->expand_ctx);
    env_file_destroy(loader->file_vars);
    free(loader);
}

env_loader_error_t env_loader_load(env_loader_t *loader, const char *path) {
    if (!loader || !path) {
        return ENV_LOADER_ERROR_FILE;
    }
    
    env_file_error_t file_err = env_file_parse(loader->file_vars, path);
    
    switch (file_err) {
        case ENV_FILE_OK:
            return ENV_LOADER_OK;
        case ENV_FILE_ERROR_FILE:
            return ENV_LOADER_ERROR_FILE;
        case ENV_FILE_ERROR_MEMORY:
            return ENV_LOADER_ERROR_MEMORY;
        case ENV_FILE_ERROR_SYNTAX:
            return ENV_LOADER_ERROR_SYNTAX;
        default:
            return ENV_LOADER_ERROR_SYNTAX;
    }
}

const char *env_loader_get_string(env_loader_t *loader, const char *key, const char *default_val) {
    if (!loader || !key) {
        return default_val;
    }
    
    const char *raw_value = env_expand_get_var(loader->expand_ctx, key);
    if (!raw_value) {
        return default_val;
    }
    
    char *expanded = NULL;
    env_expand_error_t err = env_expand_value(loader->expand_ctx, raw_value, &expanded);
    
    if (err != ENV_EXPAND_OK || !expanded) {
        free(expanded);
        return default_val;
    }
    
    static char *last_result = NULL;
    free(last_result);
    last_result = expanded;
    
    return last_result;
}

int env_loader_get_int(env_loader_t *loader, const char *key, int default_val) {
    if (!loader || !key) {
        return default_val;
    }
    
    const char *str_val = env_loader_get_string(loader, key, NULL);
    if (!str_val) {
        return default_val;
    }
    
    char *endptr;
    long result = strtol(str_val, &endptr, 10);
    
    if (endptr == str_val || *endptr != '\0') {
        return default_val;
    }
    
    return (int)result;
}

double env_loader_get_float(env_loader_t *loader, const char *key, double default_val) {
    if (!loader || !key) {
        return default_val;
    }
    
    const char *str_val = env_loader_get_string(loader, key, NULL);
    if (!str_val) {
        return default_val;
    }
    
    char *endptr;
    double result = strtod(str_val, &endptr);
    
    if (endptr == str_val || *endptr != '\0') {
        return default_val;
    }
    
    return result;
}

static int str_to_bool(const char *str) {
    if (!str) return 0;
    
    char lower[64];
    size_t i;
    for (i = 0; i < sizeof(lower) - 1 && str[i]; i++) {
        lower[i] = tolower((unsigned char)str[i]);
    }
    lower[i] = '\0';
    
    if (strcmp(lower, "true") == 0 || 
        strcmp(lower, "1") == 0 || 
        strcmp(lower, "yes") == 0 ||
        strcmp(lower, "on") == 0) {
        return 1;
    }
    
    if (strcmp(lower, "false") == 0 || 
        strcmp(lower, "0") == 0 || 
        strcmp(lower, "no") == 0 ||
        strcmp(lower, "off") == 0) {
        return 0;
    }
    
    return -1;
}

int env_loader_get_bool(env_loader_t *loader, const char *key, int default_val) {
    if (!loader || !key) {
        return default_val;
    }
    
    const char *str_val = env_loader_get_string(loader, key, NULL);
    if (!str_val) {
        return default_val;
    }
    
    int result = str_to_bool(str_val);
    if (result == -1) {
        return default_val;
    }
    
    return result;
}
