#include "conf.h"
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <ctype.h>
#include <sys/stat.h>

static void str_tolower(char* str) {
    if (!str) return;
    for (; *str; ++str) {
        *str = (char)tolower((unsigned char)*str);
    }
}

static bool is_numeric(const char* str, bool* is_float) {
    if (!str || !*str) return false;
    const char* p = str;
    bool has_dot = false;
    bool has_e = false;
    bool has_digits = false;
    if (*p == '+' || *p == '-') p++;
    while (*p) {
        if (isdigit((unsigned char)*p)) {
            has_digits = true;
            p++;
        } else if (*p == '.' && !has_dot && !has_e) {
            has_dot = true;
            p++;
        } else if ((*p == 'e' || *p == 'E') && !has_e && has_digits) {
            has_e = true;
            p++;
            if (*p == '+' || *p == '-') p++;
            has_digits = false;
        } else {
            return false;
        }
    }
    if (!has_digits) return false;
    *is_float = has_dot || has_e;
    return true;
}

static bool is_boolean(const char* str, bool* value) {
    if (!str) return false;
    char lower[16];
    size_t len = strlen(str);
    if (len >= sizeof(lower)) return false;
    for (size_t i = 0; i <= len; i++) {
        lower[i] = (char)tolower((unsigned char)str[i]);
    }
    if (strcmp(lower, "true") == 0 || strcmp(lower, "yes") == 0 || 
        strcmp(lower, "on") == 0 || strcmp(lower, "1") == 0) {
        *value = true;
        return true;
    }
    if (strcmp(lower, "false") == 0 || strcmp(lower, "no") == 0 || 
        strcmp(lower, "off") == 0 || strcmp(lower, "0") == 0) {
        *value = false;
        return true;
    }
    return false;
}

ConfValueType conf_infer_type(const char* value) {
    if (!value) return CONF_TYPE_UNKNOWN;
    bool bool_val;
    if (is_boolean(value, &bool_val)) {
        return CONF_TYPE_BOOL;
    }
    bool is_float;
    if (is_numeric(value, &is_float)) {
        return is_float ? CONF_TYPE_FLOAT : CONF_TYPE_INT;
    }
    return CONF_TYPE_STRING;
}

void conf_parse_value(ConfItem* item, const char* raw_value) {
    if (!item || !raw_value) return;
    item->type = conf_infer_type(raw_value);
    switch (item->type) {
        case CONF_TYPE_STRING:
            strncpy(item->value.str_val, raw_value, CONF_MAX_VALUE_LEN - 1);
            item->value.str_val[CONF_MAX_VALUE_LEN - 1] = '\0';
            break;
        case CONF_TYPE_INT:
            item->value.int_val = atol(raw_value);
            break;
        case CONF_TYPE_FLOAT:
            item->value.float_val = atof(raw_value);
            break;
        case CONF_TYPE_BOOL: {
            bool b;
            is_boolean(raw_value, &b);
            item->value.bool_val = b;
            break;
        }
        default:
            item->value.str_val[0] = '\0';
    }
}

static void parse_key_path(const char* key, char* section, size_t section_max,
                            char* item_key, size_t key_max) {
    if (!key || !section || !item_key) {
        if (section) section[0] = '\0';
        if (item_key) item_key[0] = '\0';
        return;
    }
    const char* dot = strchr(key, '.');
    if (dot) {
        size_t section_len = dot - key;
        if (section_len >= section_max) section_len = section_max - 1;
        strncpy(section, key, section_len);
        section[section_len] = '\0';
        strncpy(item_key, dot + 1, key_max - 1);
        item_key[key_max - 1] = '\0';
    } else {
        section[0] = '\0';
        strncpy(item_key, key, key_max - 1);
        item_key[key_max - 1] = '\0';
    }
    str_tolower(section);
    str_tolower(item_key);
}

static ConfItem* find_item_with_path(ConfContext* ctx, const char* section, const char* key) {
    if (!ctx || !key) return NULL;
    ConfSection* sec = conf_find_section(ctx, section);
    if (sec) {
        ConfItem* item = conf_find_item(sec, key);
        if (item) return item;
    }
    if (section[0] != '\0') {
        ConfSection* global = conf_find_section(ctx, "");
        if (global) {
            return conf_find_item(global, key);
        }
    }
    return NULL;
}

static void item_to_string(ConfItem* item, char* buf, size_t buf_size) {
    if (!item || !buf || buf_size == 0) return;
    switch (item->type) {
        case CONF_TYPE_STRING:
            strncpy(buf, item->value.str_val, buf_size - 1);
            buf[buf_size - 1] = '\0';
            break;
        case CONF_TYPE_INT:
            snprintf(buf, buf_size, "%ld", item->value.int_val);
            break;
        case CONF_TYPE_FLOAT:
            snprintf(buf, buf_size, "%g", item->value.float_val);
            break;
        case CONF_TYPE_BOOL:
            strncpy(buf, item->value.bool_val ? "true" : "false", buf_size - 1);
            buf[buf_size - 1] = '\0';
            break;
        default:
            buf[0] = '\0';
    }
}

ConfError conf_get_string_ex(ConfContext* ctx, const char* section, const char* key,
                              const char* default_val, char* out_buf, size_t buf_size) {
    if (!ctx || !key || !out_buf || buf_size == 0) {
        return CONF_ERR_MEMORY;
    }
    ConfItem* item = find_item_with_path(ctx, section ? section : "", key);
    if (!item) {
        if (default_val) {
            strncpy(out_buf, default_val, buf_size - 1);
            out_buf[buf_size - 1] = '\0';
        } else {
            out_buf[0] = '\0';
        }
        return CONF_ERR_KEY_NOT_FOUND;
    }
    char str_value[CONF_MAX_VALUE_LEN];
    item_to_string(item, str_value, sizeof(str_value));
    char resolved[CONF_MAX_VALUE_LEN];
    ConfError err = conf_resolve_references(ctx, section ? section : "", 
                                              str_value, 
                                              resolved, sizeof(resolved), 0);
    if (err != CONF_OK) {
        if (default_val) {
            strncpy(out_buf, default_val, buf_size - 1);
            out_buf[buf_size - 1] = '\0';
        } else {
            out_buf[0] = '\0';
        }
        return err;
    }
    strncpy(out_buf, resolved, buf_size - 1);
    out_buf[buf_size - 1] = '\0';
    return CONF_OK;
}

const char* conf_get_string(ConfContext* ctx, const char* key, const char* default_val) {
    if (!ctx || !key) return default_val;
    static char thread_local_buf[CONF_MAX_VALUE_LEN];
    char section[CONF_MAX_KEY_LEN];
    char item_key[CONF_MAX_KEY_LEN];
    parse_key_path(key, section, sizeof(section), item_key, sizeof(item_key));
    ConfError err = conf_get_string_ex(ctx, section, item_key, default_val, 
                                        thread_local_buf, sizeof(thread_local_buf));
    (void)err;
    return thread_local_buf;
}

ConfError conf_get_int_ex(ConfContext* ctx, const char* section, const char* key,
                           long default_val, long* out_val) {
    if (!ctx || !key || !out_val) {
        return CONF_ERR_MEMORY;
    }
    ConfItem* item = find_item_with_path(ctx, section ? section : "", key);
    if (!item) {
        *out_val = default_val;
        return CONF_ERR_KEY_NOT_FOUND;
    }
    switch (item->type) {
        case CONF_TYPE_INT:
            *out_val = item->value.int_val;
            return CONF_OK;
        case CONF_TYPE_FLOAT:
            *out_val = (long)item->value.float_val;
            return CONF_OK;
        case CONF_TYPE_STRING: {
            char resolved[CONF_MAX_VALUE_LEN];
            ConfError err = conf_resolve_references(ctx, section ? section : "",
                                                      item->value.str_val,
                                                      resolved, sizeof(resolved), 0);
            if (err != CONF_OK) {
                *out_val = default_val;
                return err;
            }
            bool is_float;
            if (is_numeric(resolved, &is_float)) {
                if (is_float) {
                    *out_val = (long)atof(resolved);
                } else {
                    *out_val = atol(resolved);
                }
                return CONF_OK;
            }
            *out_val = default_val;
            return CONF_ERR_TYPE_MISMATCH;
        }
        case CONF_TYPE_BOOL:
            *out_val = item->value.bool_val ? 1L : 0L;
            return CONF_OK;
        default:
            *out_val = default_val;
            return CONF_ERR_TYPE_MISMATCH;
    }
}

long conf_get_int(ConfContext* ctx, const char* key, long default_val) {
    if (!ctx || !key) return default_val;
    char section[CONF_MAX_KEY_LEN];
    char item_key[CONF_MAX_KEY_LEN];
    parse_key_path(key, section, sizeof(section), item_key, sizeof(item_key));
    long result;
    conf_get_int_ex(ctx, section, item_key, default_val, &result);
    return result;
}

ConfError conf_get_float_ex(ConfContext* ctx, const char* section, const char* key,
                             double default_val, double* out_val) {
    if (!ctx || !key || !out_val) {
        return CONF_ERR_MEMORY;
    }
    ConfItem* item = find_item_with_path(ctx, section ? section : "", key);
    if (!item) {
        *out_val = default_val;
        return CONF_ERR_KEY_NOT_FOUND;
    }
    switch (item->type) {
        case CONF_TYPE_FLOAT:
            *out_val = item->value.float_val;
            return CONF_OK;
        case CONF_TYPE_INT:
            *out_val = (double)item->value.int_val;
            return CONF_OK;
        case CONF_TYPE_STRING: {
            char resolved[CONF_MAX_VALUE_LEN];
            ConfError err = conf_resolve_references(ctx, section ? section : "",
                                                      item->value.str_val,
                                                      resolved, sizeof(resolved), 0);
            if (err != CONF_OK) {
                *out_val = default_val;
                return err;
            }
            bool is_float;
            if (is_numeric(resolved, &is_float)) {
                *out_val = atof(resolved);
                return CONF_OK;
            }
            *out_val = default_val;
            return CONF_ERR_TYPE_MISMATCH;
        }
        case CONF_TYPE_BOOL:
            *out_val = item->value.bool_val ? 1.0 : 0.0;
            return CONF_OK;
        default:
            *out_val = default_val;
            return CONF_ERR_TYPE_MISMATCH;
    }
}

double conf_get_float(ConfContext* ctx, const char* key, double default_val) {
    if (!ctx || !key) return default_val;
    char section[CONF_MAX_KEY_LEN];
    char item_key[CONF_MAX_KEY_LEN];
    parse_key_path(key, section, sizeof(section), item_key, sizeof(item_key));
    double result;
    conf_get_float_ex(ctx, section, item_key, default_val, &result);
    return result;
}

ConfError conf_get_bool_ex(ConfContext* ctx, const char* section, const char* key,
                            bool default_val, bool* out_val) {
    if (!ctx || !key || !out_val) {
        return CONF_ERR_MEMORY;
    }
    ConfItem* item = find_item_with_path(ctx, section ? section : "", key);
    if (!item) {
        *out_val = default_val;
        return CONF_ERR_KEY_NOT_FOUND;
    }
    switch (item->type) {
        case CONF_TYPE_BOOL:
            *out_val = item->value.bool_val;
            return CONF_OK;
        case CONF_TYPE_INT:
            *out_val = (item->value.int_val != 0);
            return CONF_OK;
        case CONF_TYPE_FLOAT:
            *out_val = (item->value.float_val != 0.0);
            return CONF_OK;
        case CONF_TYPE_STRING: {
            char resolved[CONF_MAX_VALUE_LEN];
            ConfError err = conf_resolve_references(ctx, section ? section : "",
                                                      item->value.str_val,
                                                      resolved, sizeof(resolved), 0);
            if (err != CONF_OK) {
                *out_val = default_val;
                return err;
            }
            bool b;
            if (is_boolean(resolved, &b)) {
                *out_val = b;
                return CONF_OK;
            }
            bool is_float;
            if (is_numeric(resolved, &is_float)) {
                if (is_float) {
                    *out_val = (atof(resolved) != 0.0);
                } else {
                    *out_val = (atol(resolved) != 0);
                }
                return CONF_OK;
            }
            *out_val = default_val;
            return CONF_ERR_TYPE_MISMATCH;
        }
        default:
            *out_val = default_val;
            return CONF_ERR_TYPE_MISMATCH;
    }
}

bool conf_get_bool(ConfContext* ctx, const char* key, bool default_val) {
    if (!ctx || !key) return default_val;
    char section[CONF_MAX_KEY_LEN];
    char item_key[CONF_MAX_KEY_LEN];
    parse_key_path(key, section, sizeof(section), item_key, sizeof(item_key));
    bool result;
    conf_get_bool_ex(ctx, section, item_key, default_val, &result);
    return result;
}

void conf_set_string(ConfContext* ctx, const char* section, const char* key, const char* value) {
    if (!ctx || !key || !value) return;
    ConfSection* sec = conf_get_or_create_section(ctx, section ? section : "");
    if (!sec) return;
    ConfItem* item = conf_find_item(sec, key);
    if (!item) {
        if (sec->item_count >= CONF_MAX_ITEMS) return;
        item = &sec->items[sec->item_count];
        strncpy(item->key, key, CONF_MAX_KEY_LEN - 1);
        item->key[CONF_MAX_KEY_LEN - 1] = '\0';
        str_tolower(item->key);
        sec->item_count++;
    }
    strncpy(item->value.str_val, value, CONF_MAX_VALUE_LEN - 1);
    item->value.str_val[CONF_MAX_VALUE_LEN - 1] = '\0';
    item->type = CONF_TYPE_STRING;
}

void conf_set_int(ConfContext* ctx, const char* section, const char* key, long value) {
    if (!ctx || !key) return;
    ConfSection* sec = conf_get_or_create_section(ctx, section ? section : "");
    if (!sec) return;
    ConfItem* item = conf_find_item(sec, key);
    if (!item) {
        if (sec->item_count >= CONF_MAX_ITEMS) return;
        item = &sec->items[sec->item_count];
        strncpy(item->key, key, CONF_MAX_KEY_LEN - 1);
        item->key[CONF_MAX_KEY_LEN - 1] = '\0';
        str_tolower(item->key);
        sec->item_count++;
    }
    item->value.int_val = value;
    item->type = CONF_TYPE_INT;
}

void conf_set_float(ConfContext* ctx, const char* section, const char* key, double value) {
    if (!ctx || !key) return;
    ConfSection* sec = conf_get_or_create_section(ctx, section ? section : "");
    if (!sec) return;
    ConfItem* item = conf_find_item(sec, key);
    if (!item) {
        if (sec->item_count >= CONF_MAX_ITEMS) return;
        item = &sec->items[sec->item_count];
        strncpy(item->key, key, CONF_MAX_KEY_LEN - 1);
        item->key[CONF_MAX_KEY_LEN - 1] = '\0';
        str_tolower(item->key);
        sec->item_count++;
    }
    item->value.float_val = value;
    item->type = CONF_TYPE_FLOAT;
}

void conf_set_bool(ConfContext* ctx, const char* section, const char* key, bool value) {
    if (!ctx || !key) return;
    ConfSection* sec = conf_get_or_create_section(ctx, section ? section : "");
    if (!sec) return;
    ConfItem* item = conf_find_item(sec, key);
    if (!item) {
        if (sec->item_count >= CONF_MAX_ITEMS) return;
        item = &sec->items[sec->item_count];
        strncpy(item->key, key, CONF_MAX_KEY_LEN - 1);
        item->key[CONF_MAX_KEY_LEN - 1] = '\0';
        str_tolower(item->key);
        sec->item_count++;
    }
    item->value.bool_val = value;
    item->type = CONF_TYPE_BOOL;
}

bool conf_needs_reload(ConfContext* ctx) {
    if (!ctx || ctx->filepath[0] == '\0') return false;
    struct stat st;
    if (stat(ctx->filepath, &st) != 0) return false;
    return (st.st_mtime != ctx->mtime);
}

ConfError conf_reload(ConfContext* ctx) {
    if (!ctx || ctx->filepath[0] == '\0') {
        return CONF_ERR_FILE_NOT_FOUND;
    }
    return conf_load_file(ctx, ctx->filepath);
}
