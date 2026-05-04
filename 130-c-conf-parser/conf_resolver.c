#include "conf.h"
#include <string.h>
#include <stdio.h>
#include <ctype.h>

typedef struct {
    char section[CONF_MAX_KEY_LEN];
    char key[CONF_MAX_KEY_LEN];
} RefPath;

static void str_tolower(char* str) {
    if (!str) return;
    for (; *str; ++str) {
        *str = (char)tolower((unsigned char)*str);
    }
}

static bool parse_reference(const char* input, const char** start, const char** end,
                            char* ref_section, size_t section_max,
                            char* ref_key, size_t key_max) {
    const char* p = input;
    while (*p) {
        if (*p == '$' && *(p + 1) == '{') {
            *start = p;
            p += 2;
            const char* content_start = p;
            while (*p && *p != '}') p++;
            if (*p != '}') return false;
            *end = p + 1;
            size_t content_len = p - content_start;
            char content[CONF_MAX_KEY_LEN * 2];
            if (content_len >= sizeof(content)) content_len = sizeof(content) - 1;
            strncpy(content, content_start, content_len);
            content[content_len] = '\0';
            char* dot = strchr(content, '.');
            if (dot) {
                *dot = '\0';
                strncpy(ref_section, content, section_max - 1);
                ref_section[section_max - 1] = '\0';
                strncpy(ref_key, dot + 1, key_max - 1);
                ref_key[key_max - 1] = '\0';
            } else {
                ref_section[0] = '\0';
                strncpy(ref_key, content, key_max - 1);
                ref_key[key_max - 1] = '\0';
            }
            str_tolower(ref_section);
            str_tolower(ref_key);
            return true;
        }
        p++;
    }
    return false;
}

static bool is_in_path(const RefPath* path, int depth, const char* section, const char* key) {
    for (int i = 0; i < depth; i++) {
        if (strcmp(path[i].section, section) == 0 && 
            strcmp(path[i].key, key) == 0) {
            return true;
        }
    }
    return false;
}

static ConfItem* find_item_with_priority(ConfContext* ctx, const char* current_section, 
                                          const char* ref_section, const char* key) {
    ConfSection* section;
    if (ref_section[0] != '\0') {
        section = conf_find_section(ctx, ref_section);
    } else if (current_section[0] != '\0') {
        section = conf_find_section(ctx, current_section);
        if (section) {
            ConfItem* item = conf_find_item(section, key);
            if (item) return item;
        }
        section = conf_find_section(ctx, "");
    } else {
        section = conf_find_section(ctx, "");
    }
    if (!section) return NULL;
    return conf_find_item(section, key);
}

static ConfError resolve_item_value(ConfContext* ctx, const char* current_section, 
                                     ConfItem* item, char* output, size_t output_size,
                                     RefPath* path, int recursion_depth) {
    if (recursion_depth >= CONF_MAX_RECURSION) {
        return CONF_ERR_CIRCULAR_REF;
    }
    char raw_value[CONF_MAX_VALUE_LEN];
    switch (item->type) {
        case CONF_TYPE_STRING:
            strncpy(raw_value, item->value.str_val, CONF_MAX_VALUE_LEN - 1);
            raw_value[CONF_MAX_VALUE_LEN - 1] = '\0';
            break;
        case CONF_TYPE_INT:
            snprintf(raw_value, sizeof(raw_value), "%ld", item->value.int_val);
            break;
        case CONF_TYPE_FLOAT:
            snprintf(raw_value, sizeof(raw_value), "%g", item->value.float_val);
            break;
        case CONF_TYPE_BOOL:
            strncpy(raw_value, item->value.bool_val ? "true" : "false", sizeof(raw_value) - 1);
            raw_value[sizeof(raw_value) - 1] = '\0';
            break;
        default:
            raw_value[0] = '\0';
    }
    char temp_output[CONF_MAX_VALUE_LEN];
    const char* src = raw_value;
    char* dst = temp_output;
    size_t dst_remaining = sizeof(temp_output) - 1;
    while (*src && dst_remaining > 0) {
        const char* ref_start;
        const char* ref_end;
        char ref_section[CONF_MAX_KEY_LEN];
        char ref_key[CONF_MAX_KEY_LEN];
        if (parse_reference(src, &ref_start, &ref_end, 
                            ref_section, sizeof(ref_section),
                            ref_key, sizeof(ref_key))) {
            while (src < ref_start && dst_remaining > 0) {
                *dst++ = *src++;
                dst_remaining--;
            }
            char actual_ref_section[CONF_MAX_KEY_LEN];
            if (ref_section[0] == '\0' && current_section[0] != '\0') {
                strncpy(actual_ref_section, current_section, CONF_MAX_KEY_LEN - 1);
                actual_ref_section[CONF_MAX_KEY_LEN - 1] = '\0';
            } else {
                strncpy(actual_ref_section, ref_section, CONF_MAX_KEY_LEN - 1);
                actual_ref_section[CONF_MAX_KEY_LEN - 1] = '\0';
            }
            if (is_in_path(path, recursion_depth, actual_ref_section, ref_key)) {
                return CONF_ERR_CIRCULAR_REF;
            }
            ConfItem* ref_item = find_item_with_priority(ctx, current_section, 
                                                          ref_section, ref_key);
            if (ref_item) {
                strncpy(path[recursion_depth].section, actual_ref_section, CONF_MAX_KEY_LEN - 1);
                path[recursion_depth].section[CONF_MAX_KEY_LEN - 1] = '\0';
                strncpy(path[recursion_depth].key, ref_key, CONF_MAX_KEY_LEN - 1);
                path[recursion_depth].key[CONF_MAX_KEY_LEN - 1] = '\0';
                char ref_resolved[CONF_MAX_VALUE_LEN];
                ConfError err = resolve_item_value(ctx, current_section, ref_item, 
                                                    ref_resolved, sizeof(ref_resolved),
                                                    path, recursion_depth + 1);
                if (err != CONF_OK) {
                    return err;
                }
                size_t ref_len = strlen(ref_resolved);
                if (ref_len > dst_remaining) {
                    ref_len = dst_remaining;
                }
                memcpy(dst, ref_resolved, ref_len);
                dst += ref_len;
                dst_remaining -= ref_len;
            }
            src = ref_end;
        } else {
            *dst++ = *src++;
            dst_remaining--;
        }
    }
    *dst = '\0';
    size_t temp_len = strlen(temp_output);
    if (temp_len >= output_size) {
        temp_len = output_size - 1;
    }
    memcpy(output, temp_output, temp_len);
    output[temp_len] = '\0';
    return CONF_OK;
}

ConfError conf_resolve_references(ConfContext* ctx, const char* section, const char* input,
                                   char* output, size_t output_size, int recursion_depth) {
    if (!ctx || !input || !output || output_size == 0) {
        return CONF_ERR_MEMORY;
    }
    RefPath path[CONF_MAX_RECURSION];
    memset(path, 0, sizeof(path));
    const char* src = input;
    char* dst = output;
    size_t dst_remaining = output_size - 1;
    while (*src && dst_remaining > 0) {
        const char* ref_start;
        const char* ref_end;
        char ref_section[CONF_MAX_KEY_LEN];
        char ref_key[CONF_MAX_KEY_LEN];
        if (parse_reference(src, &ref_start, &ref_end,
                            ref_section, sizeof(ref_section),
                            ref_key, sizeof(ref_key))) {
            while (src < ref_start && dst_remaining > 0) {
                *dst++ = *src++;
                dst_remaining--;
            }
            ConfItem* ref_item = find_item_with_priority(ctx, section ? section : "",
                                                          ref_section, ref_key);
            if (ref_item) {
                char actual_ref_section[CONF_MAX_KEY_LEN];
                if (ref_section[0] == '\0' && section && section[0] != '\0') {
                    strncpy(actual_ref_section, section, CONF_MAX_KEY_LEN - 1);
                    actual_ref_section[CONF_MAX_KEY_LEN - 1] = '\0';
                } else {
                    strncpy(actual_ref_section, ref_section, CONF_MAX_KEY_LEN - 1);
                    actual_ref_section[CONF_MAX_KEY_LEN - 1] = '\0';
                }
                strncpy(path[recursion_depth].section, actual_ref_section, CONF_MAX_KEY_LEN - 1);
                path[recursion_depth].section[CONF_MAX_KEY_LEN - 1] = '\0';
                strncpy(path[recursion_depth].key, ref_key, CONF_MAX_KEY_LEN - 1);
                path[recursion_depth].key[CONF_MAX_KEY_LEN - 1] = '\0';
                char ref_resolved[CONF_MAX_VALUE_LEN];
                ConfError err = resolve_item_value(ctx, section ? section : "", ref_item,
                                                    ref_resolved, sizeof(ref_resolved),
                                                    path, recursion_depth + 1);
                if (err != CONF_OK) {
                    return err;
                }
                size_t ref_len = strlen(ref_resolved);
                if (ref_len > dst_remaining) {
                    ref_len = dst_remaining;
                }
                memcpy(dst, ref_resolved, ref_len);
                dst += ref_len;
                dst_remaining -= ref_len;
            }
            src = ref_end;
        } else {
            *dst++ = *src++;
            dst_remaining--;
        }
    }
    *dst = '\0';
    return CONF_OK;
}
