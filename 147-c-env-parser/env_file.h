#ifndef ENV_FILE_H
#define ENV_FILE_H

#include <stddef.h>

typedef struct {
    char *key;
    char *value;
} env_entry_t;

typedef struct {
    env_entry_t *entries;
    size_t count;
    size_t capacity;
} env_file_t;

typedef enum {
    ENV_FILE_OK = 0,
    ENV_FILE_ERROR_FILE,
    ENV_FILE_ERROR_MEMORY,
    ENV_FILE_ERROR_SYNTAX
} env_file_error_t;

env_file_t *env_file_create(void);
void env_file_destroy(env_file_t *file);

env_file_error_t env_file_parse(env_file_t *file, const char *path);
env_file_error_t env_file_parse_string(env_file_t *file, const char *content);

const char *env_file_get(const env_file_t *file, const char *key);
int env_file_set(env_file_t *file, const char *key, const char *value);

char *env_file_process_value(const char *raw_value);

#endif
