#include "conf.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include <sys/stat.h>

static void str_tolower(char* str) {
    if (!str) return;
    for (; *str; ++str) {
        *str = (char)tolower((unsigned char)*str);
    }
}

static char* str_trim(char* str) {
    if (!str) return NULL;
    while (isspace((unsigned char)*str)) str++;
    if (*str == '\0') return str;
    char* end = str + strlen(str) - 1;
    while (end > str && isspace((unsigned char)*end)) end--;
    end[1] = '\0';
    return str;
}

static bool is_blank_line(const char* line) {
    while (*line) {
        if (!isspace((unsigned char)*line)) return false;
        line++;
    }
    return true;
}

static bool is_comment_line(const char* line) {
    while (isspace((unsigned char)*line)) line++;
    return (*line == '#' || *line == ';');
}

static bool is_section_header(const char* line, char* section_name, size_t max_len) {
    const char* start = line;
    while (isspace((unsigned char)*start)) start++;
    if (*start != '[') return false;
    start++;
    const char* end = start;
    while (*end && *end != ']') end++;
    if (*end != ']') return false;
    size_t len = end - start;
    if (len >= max_len) len = max_len - 1;
    strncpy(section_name, start, len);
    section_name[len] = '\0';
    str_tolower(section_name);
    return true;
}

static bool is_key_value(const char* line, char* key, size_t key_max, 
                         char* value, size_t value_max) {
    const char* p = line;
    while (isspace((unsigned char)*p)) p++;
    const char* key_start = p;
    while (*p && *p != '=' && !isspace((unsigned char)*p)) p++;
    size_t key_len = p - key_start;
    if (key_len == 0 || key_len >= key_max) return false;
    strncpy(key, key_start, key_len);
    key[key_len] = '\0';
    str_tolower(key);
    while (isspace((unsigned char)*p)) p++;
    if (*p != '=') return false;
    p++;
    while (isspace((unsigned char)*p)) p++;
    const char* val_start = p;
    size_t val_len = strlen(val_start);
    if (val_len >= value_max) val_len = value_max - 1;
    strncpy(value, val_start, val_len);
    value[val_len] = '\0';
    return true;
}

static int skip_bom(FILE* fp) {
    unsigned char bom[3];
    long pos = ftell(fp);
    if (fread(bom, 1, 3, fp) != 3) {
        fseek(fp, pos, SEEK_SET);
        return 0;
    }
    if (bom[0] == 0xEF && bom[1] == 0xBB && bom[2] == 0xBF) {
        return 3;
    }
    fseek(fp, pos, SEEK_SET);
    return 0;
}

static char* read_line_with_crlf(FILE* fp, char* buffer, size_t max_len) {
    if (feof(fp)) return NULL;
    char* p = buffer;
    size_t remaining = max_len - 1;
    int c;
    while (remaining > 0 && (c = fgetc(fp)) != EOF) {
        if (c == '\r') {
            int next = fgetc(fp);
            if (next != '\n' && next != EOF) {
                ungetc(next, fp);
            }
            break;
        }
        if (c == '\n') {
            break;
        }
        *p++ = (char)c;
        remaining--;
    }
    *p = '\0';
    if (p == buffer && feof(fp)) return NULL;
    return buffer;
}

ConfContext* conf_create(void) {
    ConfContext* ctx = (ConfContext*)calloc(1, sizeof(ConfContext));
    if (!ctx) return NULL;
    ctx->section_count = 0;
    ctx->mtime = 0;
    ctx->filepath[0] = '\0';
    ctx->last_error[0] = '\0';
    ConfSection* global = conf_get_or_create_section(ctx, "");
    if (!global) {
        free(ctx);
        return NULL;
    }
    return ctx;
}

void conf_destroy(ConfContext* ctx) {
    if (ctx) {
        free(ctx);
    }
}

ConfSection* conf_find_section(ConfContext* ctx, const char* name) {
    if (!ctx || !name) return NULL;
    char lower_name[CONF_MAX_KEY_LEN];
    strncpy(lower_name, name, CONF_MAX_KEY_LEN - 1);
    lower_name[CONF_MAX_KEY_LEN - 1] = '\0';
    str_tolower(lower_name);
    for (size_t i = 0; i < ctx->section_count; i++) {
        if (strcmp(ctx->sections[i].name, lower_name) == 0) {
            return &ctx->sections[i];
        }
    }
    return NULL;
}

ConfItem* conf_find_item(ConfSection* section, const char* key) {
    if (!section || !key) return NULL;
    char lower_key[CONF_MAX_KEY_LEN];
    strncpy(lower_key, key, CONF_MAX_KEY_LEN - 1);
    lower_key[CONF_MAX_KEY_LEN - 1] = '\0';
    str_tolower(lower_key);
    for (size_t i = 0; i < section->item_count; i++) {
        if (strcmp(section->items[i].key, lower_key) == 0) {
            return &section->items[i];
        }
    }
    return NULL;
}

ConfSection* conf_get_or_create_section(ConfContext* ctx, const char* name) {
    if (!ctx) return NULL;
    ConfSection* existing = conf_find_section(ctx, name);
    if (existing) return existing;
    if (ctx->section_count >= CONF_MAX_SECTIONS) return NULL;
    ConfSection* new_section = &ctx->sections[ctx->section_count];
    strncpy(new_section->name, name, CONF_MAX_KEY_LEN - 1);
    new_section->name[CONF_MAX_KEY_LEN - 1] = '\0';
    str_tolower(new_section->name);
    new_section->item_count = 0;
    ctx->section_count++;
    return new_section;
}

static void add_or_update_item(ConfSection* section, const char* key, const char* value) {
    if (!section || !key || !value) return;
    ConfItem* item = conf_find_item(section, key);
    if (!item) {
        if (section->item_count >= CONF_MAX_ITEMS) return;
        item = &section->items[section->item_count];
        strncpy(item->key, key, CONF_MAX_KEY_LEN - 1);
        item->key[CONF_MAX_KEY_LEN - 1] = '\0';
        str_tolower(item->key);
        section->item_count++;
    }
    conf_parse_value(item, value);
}

static void set_error(ConfContext* ctx, ConfError err, const char* msg) {
    if (!ctx) return;
    const char* err_str = conf_error_string(err);
    if (msg && msg[0]) {
        snprintf(ctx->last_error, sizeof(ctx->last_error), "%s: %s", err_str, msg);
    } else {
        snprintf(ctx->last_error, sizeof(ctx->last_error), "%s", err_str);
    }
}

static bool starts_with_triple_quote(const char* str) {
    return (str[0] == '"' && str[1] == '"' && str[2] == '"');
}

static bool ends_with_triple_quote(const char* str) {
    size_t len = strlen(str);
    if (len < 3) return false;
    return (str[len-3] == '"' && str[len-2] == '"' && str[len-1] == '"');
}

ConfError conf_load_file(ConfContext* ctx, const char* filepath) {
    if (!ctx || !filepath) return CONF_ERR_MEMORY;
    FILE* fp = fopen(filepath, "rb");
    if (!fp) {
        set_error(ctx, CONF_ERR_FILE_NOT_FOUND, filepath);
        return CONF_ERR_FILE_NOT_FOUND;
    }
    struct stat st;
    if (stat(filepath, &st) != 0) {
        fclose(fp);
        set_error(ctx, CONF_ERR_FILE_READ, "cannot stat file");
        return CONF_ERR_FILE_READ;
    }
    for (size_t i = 0; i < ctx->section_count; i++) {
        ctx->sections[i].item_count = 0;
    }
    ctx->section_count = 0;
    ConfSection* global = conf_get_or_create_section(ctx, "");
    if (!global) {
        fclose(fp);
        set_error(ctx, CONF_ERR_MEMORY, "cannot create global section");
        return CONF_ERR_MEMORY;
    }
    ConfSection* current_section = global;
    skip_bom(fp);
    char line[CONF_MAX_VALUE_LEN * 2];
    char multiline_value[CONF_MAX_VALUE_LEN];
    bool in_multiline = false;
    char multiline_key[CONF_MAX_KEY_LEN];
    int line_num = 0;
    while (read_line_with_crlf(fp, line, sizeof(line)) != NULL) {
        line_num++;
        if (in_multiline) {
            char* trimmed = str_trim(line);
            if (ends_with_triple_quote(trimmed)) {
                size_t len = strlen(trimmed);
                trimmed[len - 3] = '\0';
                size_t current_len = strlen(multiline_value);
                if (current_len > 0 && current_len < CONF_MAX_VALUE_LEN - 1) {
                    multiline_value[current_len] = '\n';
                    multiline_value[current_len + 1] = '\0';
                    current_len++;
                }
                strncat(multiline_value, trimmed, CONF_MAX_VALUE_LEN - current_len - 1);
                add_or_update_item(current_section, multiline_key, multiline_value);
                in_multiline = false;
            } else {
                size_t current_len = strlen(multiline_value);
                if (current_len > 0 && current_len < CONF_MAX_VALUE_LEN - 1) {
                    multiline_value[current_len] = '\n';
                    multiline_value[current_len + 1] = '\0';
                    current_len++;
                }
                strncat(multiline_value, trimmed, CONF_MAX_VALUE_LEN - current_len - 1);
            }
            continue;
        }
        if (is_blank_line(line)) continue;
        if (is_comment_line(line)) continue;
        char section_name[CONF_MAX_KEY_LEN];
        if (is_section_header(line, section_name, sizeof(section_name))) {
            current_section = conf_get_or_create_section(ctx, section_name);
            if (!current_section) {
                fclose(fp);
                set_error(ctx, CONF_ERR_MEMORY, "too many sections");
                return CONF_ERR_MEMORY;
            }
            continue;
        }
        char key[CONF_MAX_KEY_LEN];
        char value[CONF_MAX_VALUE_LEN];
        if (is_key_value(line, key, sizeof(key), value, sizeof(value))) {
            char* trimmed_value = str_trim(value);
            if (starts_with_triple_quote(trimmed_value)) {
                if (ends_with_triple_quote(trimmed_value) && strlen(trimmed_value) >= 6) {
                    size_t len = strlen(trimmed_value);
                    memmove(trimmed_value, trimmed_value + 3, len - 5);
                    trimmed_value[len - 6] = '\0';
                    add_or_update_item(current_section, key, trimmed_value);
                } else {
                    in_multiline = true;
                    strncpy(multiline_key, key, CONF_MAX_KEY_LEN - 1);
                    multiline_key[CONF_MAX_KEY_LEN - 1] = '\0';
                    multiline_value[0] = '\0';
                    if (strlen(trimmed_value) > 3) {
                        strcpy(multiline_value, trimmed_value + 3);
                    }
                }
            } else {
                add_or_update_item(current_section, key, trimmed_value);
            }
        } else {
            char err_msg[256];
            snprintf(err_msg, sizeof(err_msg), "line %d: syntax error", line_num);
            set_error(ctx, CONF_ERR_SYNTAX, err_msg);
        }
    }
    fclose(fp);
    strncpy(ctx->filepath, filepath, sizeof(ctx->filepath) - 1);
    ctx->filepath[sizeof(ctx->filepath) - 1] = '\0';
    ctx->mtime = st.st_mtime;
    return CONF_OK;
}

const char* conf_error_string(ConfError err) {
    switch (err) {
        case CONF_OK: return "OK";
        case CONF_ERR_FILE_NOT_FOUND: return "File not found";
        case CONF_ERR_FILE_READ: return "File read error";
        case CONF_ERR_SYNTAX: return "Syntax error";
        case CONF_ERR_SECTION_NOT_FOUND: return "Section not found";
        case CONF_ERR_KEY_NOT_FOUND: return "Key not found";
        case CONF_ERR_TYPE_MISMATCH: return "Type mismatch";
        case CONF_ERR_CIRCULAR_REF: return "Circular reference detected";
        case CONF_ERR_MEMORY: return "Memory error";
        case CONF_ERR_INVALID_FORMAT: return "Invalid format";
        default: return "Unknown error";
    }
}

const char* conf_last_error(ConfContext* ctx) {
    if (!ctx || ctx->last_error[0] == '\0') {
        return "No error";
    }
    return ctx->last_error;
}
