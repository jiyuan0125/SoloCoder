#define _GNU_SOURCE
#include "ini_parser.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include <errno.h>
#include <stddef.h>

static char* ini_strdup(const char* s) {
    if (!s) return NULL;
    size_t len = strlen(s);
    char* dup = (char*)malloc(len + 1);
    if (!dup) return NULL;
    strcpy(dup, s);
    return dup;
}

static int ini_strcasecmp(const char* a, const char* b) {
    if (!a && !b) return 0;
    if (!a || !b) return 1;
    while (*a && *b) {
        int ca = tolower((unsigned char)*a);
        int cb = tolower((unsigned char)*b);
        if (ca != cb) return ca - cb;
        a++;
        b++;
    }
    return tolower((unsigned char)*a) - tolower((unsigned char)*b);
}

static char* ini_trim(char* s) {
    if (!s) return NULL;
    
    char* original = s;
    
    while (isspace((unsigned char)*s)) s++;
    
    if (*s == '\0') {
        *original = '\0';
        return original;
    }
    
    size_t len = strlen(s);
    if (s != original) {
        memmove(original, s, len + 1);
    }
    
    char* end = original + len - 1;
    while (end > original && isspace((unsigned char)*end)) end--;
    *(end + 1) = '\0';
    
    return original;
}

static char* ini_expand_env(const char* value) {
    if (!value) return NULL;
    size_t result_cap = strlen(value) * 2 + 256;
    char* result = (char*)malloc(result_cap);
    if (!result) return NULL;
    result[0] = '\0';
    
    const char* p = value;
    size_t result_len = 0;
    
    while (*p) {
        if (p[0] == '$' && p[1] == '{') {
            const char* start = p + 2;
            const char* end = start;
            while (*end && *end != '}') end++;
            
            if (*end == '}') {
                size_t var_len = end - start;
                char* var_name = (char*)malloc(var_len + 1);
                if (!var_name) {
                    free(result);
                    return NULL;
                }
                strncpy(var_name, start, var_len);
                var_name[var_len] = '\0';
                
                char* var_value = getenv(var_name);
                free(var_name);
                
                if (var_value) {
                    size_t var_value_len = strlen(var_value);
                    if (result_len + var_value_len + 1 > result_cap) {
                        result_cap = (result_len + var_value_len) * 2;
                        char* new_result = (char*)realloc(result, result_cap);
                        if (!new_result) {
                            free(result);
                            return NULL;
                        }
                        result = new_result;
                    }
                    strcpy(result + result_len, var_value);
                    result_len += var_value_len;
                } else {
                    size_t literal_len = end - p + 1;
                    if (result_len + literal_len + 1 > result_cap) {
                        result_cap = (result_len + literal_len) * 2;
                        char* new_result = (char*)realloc(result, result_cap);
                        if (!new_result) {
                            free(result);
                            return NULL;
                        }
                        result = new_result;
                    }
                    strncpy(result + result_len, p, literal_len);
                    result_len += literal_len;
                    result[result_len] = '\0';
                }
                p = end + 1;
            } else {
                if (result_len + 2 > result_cap) {
                    result_cap = (result_len + 2) * 2;
                    char* new_result = (char*)realloc(result, result_cap);
                    if (!new_result) {
                        free(result);
                        return NULL;
                    }
                    result = new_result;
                }
                result[result_len++] = *p++;
                result[result_len] = '\0';
            }
        } else {
            if (result_len + 2 > result_cap) {
                result_cap = (result_len + 2) * 2;
                char* new_result = (char*)realloc(result, result_cap);
                if (!new_result) {
                    free(result);
                    return NULL;
                }
                result = new_result;
            }
            result[result_len++] = *p++;
            result[result_len] = '\0';
        }
    }
    return result;
}

static void ini_set_error(ini_parser_t* parser, int line, const char* msg) {
    if (!parser) return;
    parser->last_error.line_number = line;
    snprintf(parser->last_error.message, sizeof(parser->last_error.message), "%s", msg);
}

static ini_parser_t* ini_parser_create(void) {
    ini_parser_t* parser = (ini_parser_t*)calloc(1, sizeof(ini_parser_t));
    if (!parser) return NULL;
    parser->section_capacity = 8;
    parser->sections = (ini_section_t*)calloc(parser->section_capacity, sizeof(ini_section_t));
    if (!parser->sections) {
        free(parser);
        return NULL;
    }
    parser->leading_capacity = 8;
    parser->leading_lines = (char**)calloc(parser->leading_capacity, sizeof(char*));
    if (!parser->leading_lines) {
        free(parser->sections);
        free(parser);
        return NULL;
    }
    return parser;
}

static void ini_kv_free(ini_kv_t* kv) {
    if (!kv) return;
    free(kv->key);
    free(kv->value);
    free(kv->raw_line);
}

static void ini_section_free(ini_section_t* sec) {
    if (!sec) return;
    free(sec->name);
    free(sec->raw_header);
    for (int i = 0; i < sec->entry_count; i++) {
        ini_kv_free(&sec->entries[i]);
    }
    free(sec->entries);
    for (int i = 0; i < sec->trailing_count; i++) {
        free(sec->trailing_lines[i]);
    }
    free(sec->trailing_lines);
}

void ini_free(ini_parser_t* parser) {
    if (!parser) return;
    for (int i = 0; i < parser->section_count; i++) {
        ini_section_free(&parser->sections[i]);
    }
    free(parser->sections);
    for (int i = 0; i < parser->leading_count; i++) {
        free(parser->leading_lines[i]);
    }
    free(parser->leading_lines);
    free(parser);
}

static ini_section_t* ini_find_section(ini_parser_t* parser, const char* name) {
    if (!parser || !name) return NULL;
    for (int i = 0; i < parser->section_count; i++) {
        if (ini_strcasecmp(parser->sections[i].name, name) == 0) {
            return &parser->sections[i];
        }
    }
    return NULL;
}

static ini_kv_t* ini_find_kv(ini_section_t* sec, const char* key) {
    if (!sec || !key) return NULL;
    for (int i = 0; i < sec->entry_count; i++) {
        if (ini_strcasecmp(sec->entries[i].key, key) == 0) {
            return &sec->entries[i];
        }
    }
    return NULL;
}

static ini_section_t* ini_add_section(ini_parser_t* parser, const char* name, const char* raw_header) {
    if (!parser || !name) return NULL;
    
    ini_section_t* existing = ini_find_section(parser, name);
    if (existing) return existing;
    
    if (parser->section_count >= parser->section_capacity) {
        int new_cap = parser->section_capacity * 2;
        ini_section_t* new_sections = (ini_section_t*)realloc(parser->sections, new_cap * sizeof(ini_section_t));
        if (!new_sections) return NULL;
        parser->sections = new_sections;
        parser->section_capacity = new_cap;
    }
    
    ini_section_t* sec = &parser->sections[parser->section_count++];
    memset(sec, 0, sizeof(ini_section_t));
    sec->name = ini_strdup(name);
    if (!sec->name) {
        parser->section_count--;
        return NULL;
    }
    if (raw_header) {
        sec->raw_header = ini_strdup(raw_header);
    }
    sec->entry_capacity = 8;
    sec->entries = (ini_kv_t*)calloc(sec->entry_capacity, sizeof(ini_kv_t));
    if (!sec->entries) {
        free(sec->name);
        free(sec->raw_header);
        parser->section_count--;
        return NULL;
    }
    sec->trailing_capacity = 8;
    sec->trailing_lines = (char**)calloc(sec->trailing_capacity, sizeof(char*));
    if (!sec->trailing_lines) {
        free(sec->entries);
        free(sec->name);
        free(sec->raw_header);
        parser->section_count--;
        return NULL;
    }
    return sec;
}

static int ini_add_kv(ini_section_t* sec, const char* key, const char* value, const char* raw_line) {
    if (!sec || !key) return -1;
    
    ini_kv_t* existing = ini_find_kv(sec, key);
    if (existing) {
        free(existing->value);
        existing->value = value ? ini_strdup(value) : NULL;
        if (raw_line) {
            free(existing->raw_line);
            existing->raw_line = ini_strdup(raw_line);
        }
        return 0;
    }
    
    if (sec->entry_count >= sec->entry_capacity) {
        int new_cap = sec->entry_capacity * 2;
        ini_kv_t* new_entries = (ini_kv_t*)realloc(sec->entries, new_cap * sizeof(ini_kv_t));
        if (!new_entries) return -1;
        sec->entries = new_entries;
        sec->entry_capacity = new_cap;
    }
    
    ini_kv_t* kv = &sec->entries[sec->entry_count++];
    memset(kv, 0, sizeof(ini_kv_t));
    kv->key = ini_strdup(key);
    if (!kv->key) {
        sec->entry_count--;
        return -1;
    }
    kv->value = value ? ini_strdup(value) : NULL;
    kv->raw_line = raw_line ? ini_strdup(raw_line) : NULL;
    kv->modified = 0;
    return 0;
}

static int ini_add_leading_line(ini_parser_t* parser, const char* line) {
    if (!parser || !line) return -1;
    
    if (parser->leading_count >= parser->leading_capacity) {
        int new_cap = parser->leading_capacity * 2;
        char** new_lines = (char**)realloc(parser->leading_lines, new_cap * sizeof(char*));
        if (!new_lines) return -1;
        parser->leading_lines = new_lines;
        parser->leading_capacity = new_cap;
    }
    
    parser->leading_lines[parser->leading_count++] = ini_strdup(line);
    return parser->leading_lines[parser->leading_count - 1] ? 0 : -1;
}

static int ini_add_trailing_line(ini_section_t* sec, const char* line) {
    if (!sec || !line) return -1;
    
    if (sec->trailing_count >= sec->trailing_capacity) {
        int new_cap = sec->trailing_capacity * 2;
        char** new_lines = (char**)realloc(sec->trailing_lines, new_cap * sizeof(char*));
        if (!new_lines) return -1;
        sec->trailing_lines = new_lines;
        sec->trailing_capacity = new_cap;
    }
    
    sec->trailing_lines[sec->trailing_count++] = ini_strdup(line);
    return sec->trailing_lines[sec->trailing_count - 1] ? 0 : -1;
}

static int ini_is_comment(const char* line) {
    if (!line || !*line) return 0;
    while (isspace((unsigned char)*line)) line++;
    return (*line == ';' || *line == '#');
}

static int ini_is_empty(const char* line) {
    if (!line) return 1;
    while (isspace((unsigned char)*line)) line++;
    return *line == '\0';
}

static int ini_parse_file_internal(ini_parser_t* parser, const char* path, int depth);

static int ini_handle_include(ini_parser_t* parser, const char* value, int depth, int line_num) {
    if (depth >= INI_MAX_INCLUDE_DEPTH) {
        char err[256];
        snprintf(err, sizeof(err), "include recursion depth exceeded (%d)", depth);
        ini_set_error(parser, line_num, err);
        return -1;
    }
    
    char* include_path = ini_trim(ini_strdup(value));
    if (!include_path) {
        ini_set_error(parser, line_num, "memory allocation failed");
        return -1;
    }
    
    int result = ini_parse_file_internal(parser, include_path, depth + 1);
    free(include_path);
    return result;
}

static int ini_parse_file_internal(ini_parser_t* parser, const char* path, int depth) {
    FILE* f = fopen(path, "r");
    if (!f) {
        char err[256];
        snprintf(err, sizeof(err), "cannot open file '%s': %s", path, strerror(errno));
        ini_set_error(parser, 0, err);
        return -1;
    }
    
    char* line = NULL;
    size_t line_cap = 0;
    ssize_t line_len;
    int line_num = 0;
    ini_section_t* current_sec = NULL;
    
    while ((line_len = getline(&line, &line_cap, f)) != -1) {
        line_num++;
        
        if (line_len > 0 && line[line_len - 1] == '\n') {
            line[--line_len] = '\0';
            if (line_len > 0 && line[line_len - 1] == '\r') {
                line[--line_len] = '\0';
            }
        }
        
        char* line_dup = ini_strdup(line);
        if (!line_dup) {
            ini_set_error(parser, line_num, "memory allocation failed");
            free(line);
            fclose(f);
            return -1;
        }
        
        if (ini_is_empty(line)) {
            if (current_sec) {
                ini_add_trailing_line(current_sec, line_dup);
            } else {
                ini_add_leading_line(parser, line_dup);
            }
            free(line_dup);
            continue;
        }
        
        if (ini_is_comment(line)) {
            if (current_sec) {
                ini_add_trailing_line(current_sec, line_dup);
            } else {
                ini_add_leading_line(parser, line_dup);
            }
            free(line_dup);
            continue;
        }
        
        char* p = line;
        while (isspace((unsigned char)*p)) p++;
        
        if (*p == '[') {
            p++;
            char* sec_start = p;
            while (*p && *p != ']') p++;
            
            if (*p != ']') {
                ini_set_error(parser, line_num, "unterminated section header");
                free(line_dup);
                free(line);
                fclose(f);
                return -1;
            }
            
            *p = '\0';
            char* sec_name = ini_trim(sec_start);
            if (!*sec_name) {
                ini_set_error(parser, line_num, "empty section name");
                free(line_dup);
                free(line);
                fclose(f);
                return -1;
            }
            
            p++;
            while (*p && isspace((unsigned char)*p)) p++;
            if (*p && *p != ';' && *p != '#') {
                ini_set_error(parser, line_num, "unexpected characters after section header");
                free(line_dup);
                free(line);
                fclose(f);
                return -1;
            }
            
            current_sec = ini_add_section(parser, sec_name, line_dup);
            if (!current_sec) {
                ini_set_error(parser, line_num, "failed to add section");
                free(line_dup);
                free(line);
                fclose(f);
                return -1;
            }
            free(line_dup);
            continue;
        }
        
        char* eq = strchr(p, '=');
        if (eq) {
            *eq = '\0';
            char* key = ini_trim(p);
            char* value = eq + 1;
            
            if (!*key) {
                ini_set_error(parser, line_num, "empty key");
                free(line_dup);
                free(line);
                fclose(f);
                return -1;
            }
            
            if (ini_strcasecmp(key, "include") == 0 && !current_sec) {
                if (ini_handle_include(parser, value, depth, line_num) != 0) {
                    free(line_dup);
                    free(line);
                    fclose(f);
                    return -1;
                }
                free(line_dup);
                continue;
            }
            
            char* trimmed_value = ini_trim(value);
            char* expanded_value = ini_expand_env(trimmed_value);
            if (!expanded_value) {
                ini_set_error(parser, line_num, "memory allocation failed");
                free(line_dup);
                free(line);
                fclose(f);
                return -1;
            }
            char* final_value = ini_trim(expanded_value);
            
            if (!current_sec) {
                current_sec = ini_add_section(parser, "", NULL);
                if (!current_sec) {
                    free(expanded_value);
                    free(line_dup);
                    free(line);
                    fclose(f);
                    ini_set_error(parser, line_num, "failed to create default section");
                    return -1;
                }
            }
            
            if (ini_add_kv(current_sec, key, final_value, line_dup) != 0) {
                free(expanded_value);
                free(line_dup);
                free(line);
                fclose(f);
                ini_set_error(parser, line_num, "failed to add key-value");
                return -1;
            }
            
            free(expanded_value);
            free(line_dup);
            continue;
        }
        
        ini_set_error(parser, line_num, "syntax error");
        free(line_dup);
        free(line);
        fclose(f);
        return -1;
    }
    
    free(line);
    fclose(f);
    return 0;
}

ini_parser_t* ini_parse(const char* path) {
    ini_parser_t* parser = ini_parser_create();
    if (!parser) return NULL;
    
    if (ini_parse_file_internal(parser, path, 0) != 0) {
        ini_free(parser);
        return NULL;
    }
    
    return parser;
}

const char* ini_get(const ini_parser_t* parser, const char* section, const char* key, const char* default_value) {
    if (!parser || !section || !key) return default_value;
    
    ini_section_t* sec = ini_find_section((ini_parser_t*)parser, section);
    if (!sec) return default_value;
    
    ini_kv_t* kv = ini_find_kv(sec, key);
    if (!kv || !kv->value) return default_value;
    
    return kv->value;
}

const char* ini_get_string(const ini_parser_t* parser, const char* section, const char* key, const char* default_value) {
    return ini_get(parser, section, key, default_value);
}

int ini_get_int(const ini_parser_t* parser, const char* section, const char* key, int default_value) {
    const char* val = ini_get(parser, section, key, NULL);
    if (!val) return default_value;
    
    char* p = (char*)val;
    while (isspace((unsigned char)*p)) p++;
    
    if (*p != '+' && *p != '-' && !isdigit((unsigned char)*p)) {
        return default_value;
    }
    
    char sign = *p;
    if (sign == '+' || sign == '-') p++;
    
    if (!isdigit((unsigned char)*p)) {
        return default_value;
    }
    
    while (isdigit((unsigned char)*p)) p++;
    while (isspace((unsigned char)*p)) p++;
    
    if (*p != '\0') {
        return default_value;
    }
    
    errno = 0;
    long lval = strtol(val, NULL, 10);
    if (errno != 0) {
        return default_value;
    }
    
    return (int)lval;
}

int ini_get_bool(const ini_parser_t* parser, const char* section, const char* key, int default_value) {
    const char* val = ini_get(parser, section, key, NULL);
    if (!val) return default_value;
    
    char* dup_val = ini_strdup(val);
    if (!dup_val) return default_value;
    
    char* trimmed = ini_trim(dup_val);
    
    int result = default_value;
    
    if (ini_strcasecmp(trimmed, "true") == 0 ||
        ini_strcasecmp(trimmed, "yes") == 0 ||
        strcmp(trimmed, "1") == 0) {
        result = 1;
    } else if (ini_strcasecmp(trimmed, "false") == 0 ||
               ini_strcasecmp(trimmed, "no") == 0 ||
               strcmp(trimmed, "0") == 0) {
        result = 0;
    }
    
    free(dup_val);
    return result;
}

int ini_set(ini_parser_t* parser, const char* section, const char* key, const char* value) {
    if (!parser || !section || !key) return -1;
    
    ini_section_t* sec = ini_find_section(parser, section);
    if (!sec) {
        sec = ini_add_section(parser, section, NULL);
        if (!sec) return -1;
    }
    
    ini_kv_t* kv = ini_find_kv(sec, key);
    if (kv) {
        free(kv->value);
        kv->value = value ? ini_strdup(value) : NULL;
        kv->modified = 1;
    } else {
        if (ini_add_kv(sec, key, value, NULL) != 0) {
            return -1;
        }
        sec->entries[sec->entry_count - 1].modified = 1;
    }
    
    return 0;
}

static int ini_write_kv_modified(FILE* f, const ini_kv_t* kv) {
    if (!kv->raw_line || kv->modified) {
        fprintf(f, "%s = %s\n", kv->key, kv->value ? kv->value : "");
    } else {
        fprintf(f, "%s\n", kv->raw_line);
    }
    return 0;
}

int ini_save(const ini_parser_t* parser, const char* path) {
    if (!parser || !path) return -1;
    
    FILE* f = fopen(path, "w");
    if (!f) return -1;
    
    for (int i = 0; i < parser->leading_count; i++) {
        fprintf(f, "%s\n", parser->leading_lines[i]);
    }
    
    for (int i = 0; i < parser->section_count; i++) {
        const ini_section_t* sec = &parser->sections[i];
        
        if (sec->raw_header) {
            fprintf(f, "%s\n", sec->raw_header);
        } else if (sec->name && *sec->name) {
            fprintf(f, "[%s]\n", sec->name);
        }
        
        for (int j = 0; j < sec->entry_count; j++) {
            const ini_kv_t* kv = &sec->entries[j];
            ini_write_kv_modified(f, kv);
        }
        
        for (int j = 0; j < sec->trailing_count; j++) {
            fprintf(f, "%s\n", sec->trailing_lines[j]);
        }
    }
    
    fclose(f);
    return 0;
}

const char* ini_get_error_message(const ini_parser_t* parser) {
    if (!parser) return NULL;
    return parser->last_error.message;
}

int ini_get_error_line(const ini_parser_t* parser) {
    if (!parser) return 0;
    return parser->last_error.line_number;
}
