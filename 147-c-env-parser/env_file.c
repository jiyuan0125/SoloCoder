#include "env_file.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <ctype.h>

#define INITIAL_CAPACITY 8

static char *trim_whitespace(const char *str) {
    if (!str) return NULL;
    
    while (isspace((unsigned char)*str)) str++;
    if (*str == '\0') return strdup("");
    
    const char *end = str + strlen(str) - 1;
    while (end > str && isspace((unsigned char)*end)) end--;
    
    size_t len = end - str + 1;
    char *result = (char *)malloc(len + 1);
    if (!result) return NULL;
    strncpy(result, str, len);
    result[len] = '\0';
    return result;
}

static int is_empty_or_comment(const char *line) {
    while (isspace((unsigned char)*line)) line++;
    return (*line == '\0' || *line == '#');
}

env_file_t *env_file_create(void) {
    env_file_t *file = (env_file_t *)malloc(sizeof(env_file_t));
    if (!file) return NULL;
    
    file->entries = (env_entry_t *)malloc(INITIAL_CAPACITY * sizeof(env_entry_t));
    if (!file->entries) {
        free(file);
        return NULL;
    }
    
    file->count = 0;
    file->capacity = INITIAL_CAPACITY;
    return file;
}

void env_file_destroy(env_file_t *file) {
    if (!file) return;
    
    for (size_t i = 0; i < file->count; i++) {
        free(file->entries[i].key);
        free(file->entries[i].value);
    }
    free(file->entries);
    free(file);
}

static int env_file_ensure_capacity(env_file_t *file) {
    if (file->count >= file->capacity) {
        size_t new_capacity = file->capacity * 2;
        env_entry_t *new_entries = (env_entry_t *)realloc(file->entries, new_capacity * sizeof(env_entry_t));
        if (!new_entries) return 0;
        file->entries = new_entries;
        file->capacity = new_capacity;
    }
    return 1;
}

const char *env_file_get(const env_file_t *file, const char *key) {
    if (!file || !key) return NULL;
    
    for (size_t i = 0; i < file->count; i++) {
        if (strcmp(file->entries[i].key, key) == 0) {
            return file->entries[i].value;
        }
    }
    return NULL;
}

int env_file_set(env_file_t *file, const char *key, const char *value) {
    if (!file || !key) return 0;
    
    for (size_t i = 0; i < file->count; i++) {
        if (strcmp(file->entries[i].key, key) == 0) {
            free(file->entries[i].value);
            file->entries[i].value = value ? strdup(value) : NULL;
            return file->entries[i].value != NULL;
        }
    }
    
    if (!env_file_ensure_capacity(file)) return 0;
    
    file->entries[file->count].key = strdup(key);
    file->entries[file->count].value = value ? strdup(value) : NULL;
    
    if (!file->entries[file->count].key || 
        (value && !file->entries[file->count].value)) {
        free(file->entries[file->count].key);
        free(file->entries[file->count].value);
        return 0;
    }
    
    file->count++;
    return 1;
}

static char *process_unquoted(const char *value) {
    return trim_whitespace(value);
}

static char *process_single_quoted(const char *value) {
    size_t len = strlen(value);
    if (len < 2 || value[0] != '\'' || value[len-1] != '\'') {
        return strdup(value);
    }
    
    size_t inner_len = len - 2;
    char *result = (char *)malloc(inner_len + 1);
    if (!result) return NULL;
    
    strncpy(result, value + 1, inner_len);
    result[inner_len] = '\0';
    return result;
}

static char *process_double_quoted(const char *value) {
    size_t len = strlen(value);
    if (len < 2 || value[0] != '"' || value[len-1] != '"') {
        return strdup(value);
    }
    
    char *result = (char *)malloc(len);
    if (!result) return NULL;
    
    size_t write_idx = 0;
    for (size_t i = 1; i < len - 1; i++) {
        if (value[i] == '\\' && i + 1 < len - 1) {
            switch (value[i+1]) {
                case 'n':
                    result[write_idx++] = '\n';
                    i++;
                    break;
                case 't':
                    result[write_idx++] = '\t';
                    i++;
                    break;
                case '\\':
                    result[write_idx++] = '\\';
                    i++;
                    break;
                case '"':
                    result[write_idx++] = '"';
                    i++;
                    break;
                default:
                    result[write_idx++] = value[i];
            }
        } else {
            result[write_idx++] = value[i];
        }
    }
    
    result[write_idx] = '\0';
    return result;
}

static char *process_triple_quoted(const char *value) {
    size_t len = strlen(value);
    if (len < 6 || 
        (value[0] != '"' || value[1] != '"' || value[2] != '"') ||
        (value[len-3] != '"' || value[len-2] != '"' || value[len-1] != '"')) {
        return strdup(value);
    }
    
    size_t inner_len = len - 6;
    char *result = (char *)malloc(inner_len + 1);
    if (!result) return NULL;
    
    strncpy(result, value + 3, inner_len);
    result[inner_len] = '\0';
    return result;
}

char *env_file_process_value(const char *raw_value) {
    if (!raw_value) return strdup("");
    
    char *trimmed = trim_whitespace(raw_value);
    if (!trimmed) return NULL;
    
    size_t len = strlen(trimmed);
    char *result = NULL;
    
    if (len >= 6 && 
        trimmed[0] == '"' && trimmed[1] == '"' && trimmed[2] == '"' &&
        trimmed[len-3] == '"' && trimmed[len-2] == '"' && trimmed[len-1] == '"') {
        result = process_triple_quoted(trimmed);
    }
    else if (len >= 2 && trimmed[0] == '"' && trimmed[len-1] == '"') {
        result = process_double_quoted(trimmed);
    }
    else if (len >= 2 && trimmed[0] == '\'' && trimmed[len-1] == '\'') {
        result = process_single_quoted(trimmed);
    }
    else {
        result = process_unquoted(trimmed);
    }
    
    free(trimmed);
    return result;
}

static int parse_key_value(const char *line, char **key, char **value) {
    const char *eq_pos = strchr(line, '=');
    if (!eq_pos) return 0;
    
    size_t key_len = eq_pos - line;
    char *raw_key = (char *)malloc(key_len + 1);
    if (!raw_key) return 0;
    strncpy(raw_key, line, key_len);
    raw_key[key_len] = '\0';
    
    *key = trim_whitespace(raw_key);
    free(raw_key);
    
    if (!*key) return 0;
    
    const char *raw_value = eq_pos + 1;
    *value = strdup(raw_value);
    if (!*value) {
        free(*key);
        return 0;
    }
    
    return 1;
}

env_file_error_t env_file_parse_string(env_file_t *file, const char *content) {
    if (!file || !content) return ENV_FILE_ERROR_SYNTAX;
    
    const char *ptr = content;
    char line[4096];
    size_t line_idx = 0;
    
    char *multiline_buffer = NULL;
    size_t multiline_len = 0;
    size_t multiline_cap = 0;
    int in_multiline = 0;
    char multiline_key[4096] = {0};
    
    while (*ptr) {
        if (*ptr == '\n' || *ptr == '\0') {
            line[line_idx] = '\0';
            
            if (in_multiline) {
                size_t new_len = multiline_len + line_idx + 1;
                if (new_len >= multiline_cap) {
                    size_t new_cap = multiline_cap ? multiline_cap * 2 : 4096;
                    char *new_buf = (char *)realloc(multiline_buffer, new_cap);
                    if (!new_buf) {
                        free(multiline_buffer);
                        return ENV_FILE_ERROR_MEMORY;
                    }
                    multiline_buffer = new_buf;
                    multiline_cap = new_cap;
                }
                memcpy(multiline_buffer + multiline_len, line, line_idx);
                multiline_buffer[multiline_len + line_idx] = '\n';
                multiline_len = new_len;
                multiline_buffer[multiline_len] = '\0';
                
                const char *end_triple = "\"\"\"";
                int found_end = 0;
                size_t content_len = 0;
                
                if (multiline_len >= 4 && 
                    multiline_buffer[multiline_len - 1] == '\n' &&
                    strncmp(multiline_buffer + multiline_len - 4, end_triple, 3) == 0) {
                    content_len = multiline_len - 4;
                    found_end = 1;
                }
                else if (multiline_len >= 3 && 
                         strncmp(multiline_buffer + multiline_len - 3, end_triple, 3) == 0) {
                    content_len = multiline_len - 3;
                    found_end = 1;
                }
                
                if (found_end) {
                    multiline_buffer[content_len] = '\0';
                    
                    char *processed = (char *)malloc(content_len + 6 + 1);
                    if (!processed) {
                        free(multiline_buffer);
                        return ENV_FILE_ERROR_MEMORY;
                    }
                    strcpy(processed, "\"\"\"");
                    strncat(processed, multiline_buffer, content_len);
                    strcat(processed, "\"\"\"");
                    
                    char *final_value = env_file_process_value(processed);
                    if (final_value) {
                        env_file_set(file, multiline_key, final_value);
                        free(final_value);
                    }
                    free(processed);
                    free(multiline_buffer);
                    
                    in_multiline = 0;
                    multiline_buffer = NULL;
                    multiline_len = 0;
                    multiline_key[0] = '\0';
                }
            }
            else if (!is_empty_or_comment(line)) {
                char *key = NULL;
                char *value = NULL;
                
                if (parse_key_value(line, &key, &value)) {
                    char *trimmed_value = trim_whitespace(value);
                    if (trimmed_value) {
                        size_t val_len = strlen(trimmed_value);
                        if (val_len >= 3 && 
                            trimmed_value[0] == '"' && trimmed_value[1] == '"' && trimmed_value[2] == '"' &&
                            !(val_len >= 6 && trimmed_value[val_len-3] == '"' && trimmed_value[val_len-2] == '"' && trimmed_value[val_len-1] == '"')) {
                            
                            in_multiline = 1;
                            strncpy(multiline_key, key, sizeof(multiline_key) - 1);
                            
                            multiline_buffer = strdup(trimmed_value + 3);
                            if (!multiline_buffer) {
                                free(key);
                                free(value);
                                free(trimmed_value);
                                return ENV_FILE_ERROR_MEMORY;
                            }
                            multiline_len = strlen(multiline_buffer);
                            if (multiline_len > 0 && multiline_buffer[multiline_len - 1] != '\n') {
                                multiline_buffer = (char *)realloc(multiline_buffer, multiline_len + 2);
                                multiline_buffer[multiline_len] = '\n';
                                multiline_buffer[multiline_len + 1] = '\0';
                                multiline_len++;
                            }
                            multiline_cap = multiline_len + 1;
                        }
                        else {
                            char *processed = env_file_process_value(trimmed_value);
                            if (processed) {
                                env_file_set(file, key, processed);
                                free(processed);
                            }
                        }
                        free(trimmed_value);
                    }
                    free(key);
                    free(value);
                }
                else {
                    return ENV_FILE_ERROR_SYNTAX;
                }
            }
            
            line_idx = 0;
            if (*ptr == '\n') ptr++;
        }
        else if (*ptr == '\r') {
            ptr++;
        }
        else {
            if (line_idx < sizeof(line) - 1) {
                line[line_idx++] = *ptr;
            }
            ptr++;
        }
    }
    
    if (in_multiline) {
        free(multiline_buffer);
        return ENV_FILE_ERROR_SYNTAX;
    }
    
    return ENV_FILE_OK;
}

env_file_error_t env_file_parse(env_file_t *file, const char *path) {
    if (!file || !path) return ENV_FILE_ERROR_FILE;
    
    FILE *f = fopen(path, "r");
    if (!f) return ENV_FILE_ERROR_FILE;
    
    fseek(f, 0, SEEK_END);
    long size = ftell(f);
    fseek(f, 0, SEEK_SET);
    
    if (size < 0) {
        fclose(f);
        return ENV_FILE_ERROR_FILE;
    }
    
    char *content = (char *)malloc(size + 1);
    if (!content) {
        fclose(f);
        return ENV_FILE_ERROR_MEMORY;
    }
    
    size_t read_size = fread(content, 1, size, f);
    content[read_size] = '\0';
    fclose(f);
    
    env_file_error_t result = env_file_parse_string(file, content);
    free(content);
    
    return result;
}
